package postgres

import (
	"context"
	"database/sql"

	"fmt"
	"slices"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	// Users Table Columns
	COLUMN_EMAIL      = "email"
	COLUMN_DELETED_AT = "deleted_at"

	// Policies Table Columns
	COLUMN_RESOURCE_ID       = "resource_id"
	COLUMN_RESOURCE_TYPE     = "resource_type"
	COLUMN_PRINCIPAL_ID      = "principal_id"
	COLUMN_PRINCIPAL_TYPE    = "principal_type"
	COLUMN_POLICY_CREATED_AT = "created_at"

	// Custom Aggregated Columns
	COLUMN_ROLE_NAMES      = "role_names"
	COLUMN_ROLE_TITLES     = "role_titles"
	COLUMN_ROLE_IDS        = "role_ids"
	COLUMN_ROLE_ID         = "role_id"
	COLUMN_ORG_JOINED_DATE = "org_joined_at"
)

type OrgUsersRepository struct {
	dbc *db.Client
}

type OrgUsers struct {
	UserID      sql.NullString `db:"id"`
	UserTitle   sql.NullString `db:"title"`
	UserName    sql.NullString `db:"name"`
	UserEmail   sql.NullString `db:"email"`
	UserState   sql.NullString `db:"state"`
	UserAvatar  sql.NullString `db:"avatar"`
	RoleNames   sql.NullString `db:"role_names"`
	RoleTitles  sql.NullString `db:"role_titles"`
	RoleIDs     sql.NullString `db:"role_ids"`
	OrgID       sql.NullString `db:"org_id"`
	OrgJoinedAt sql.NullTime   `db:"org_joined_at"`
}

type OrgUsersGroup struct {
	Name sql.NullString      `db:"name"`
	Data []OrgUsersGroupData `db:"data"`
}

type OrgUsersGroupData struct {
	Name  sql.NullString `db:"values"`
	Count int            `db:"count"`
}

func (u *OrgUsers) transformToAggregatedUser() svc.AggregatedUser {
	return svc.AggregatedUser{
		ID:          u.UserID.String,
		Name:        u.UserName.String,
		Title:       u.UserTitle.String,
		Avatar:      u.UserAvatar.String,
		Email:       u.UserEmail.String,
		State:       user.State(u.UserState.String),
		RoleNames:   u.RoleNames.String,
		RoleTitles:  u.RoleTitles.String,
		RoleIDs:     u.RoleIDs.String,
		OrgID:       u.OrgID.String,
		OrgJoinedAt: u.OrgJoinedAt.Time,
	}
}

func (o *OrgUsersGroup) transformToOrgUsersGroup() svc.Group {
	orgUsersGroupData := make([]svc.GroupData, 0)
	for _, groupDataItem := range o.Data {
		orgUsersGroupData = append(orgUsersGroupData, svc.GroupData{
			Name:  groupDataItem.Name.String,
			Count: groupDataItem.Count,
		})
	}
	return svc.Group{
		Name: o.Name.String,
		Data: orgUsersGroupData,
	}
}

func NewOrgUsersRepository(dbc *db.Client) *OrgUsersRepository {
	return &OrgUsersRepository{
		dbc: dbc,
	}
}

func (r OrgUsersRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrgUsers, error) {
	dataQuery, params, err := r.prepareDataQuery(orgID, rql)
	if err != nil {
		return svc.OrgUsers{}, err
	}

	var orgUsers []OrgUsers
	var orgUsersGroupData []OrgUsersGroupData
	var orgUsersGroup OrgUsersGroup

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "GetOrgUsers", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &orgUsers, dataQuery, params...)
		})

		if err != nil {
			return err
		}

		if len(rql.GroupBy) > 0 {
			groupByQuery, groupByParams, err := r.prepareGroupByQuery(orgID, rql)
			if err != nil {
				return err
			}

			err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "GetOrgusersWithGroup", func(ctx context.Context) error {
				return tx.SelectContext(ctx, &orgUsersGroupData, groupByQuery, groupByParams...)
			})

			if err != nil {
				return err
			}
			orgUsersGroup.Name = sql.NullString{String: rql.GroupBy[0]}
			orgUsersGroup.Data = orgUsersGroupData
		}
		return err
	})

	if err != nil {
		return svc.OrgUsers{}, err
	}

	res := make([]svc.AggregatedUser, 0)
	for _, user := range orgUsers {
		res = append(res, user.transformToAggregatedUser())
	}
	return svc.OrgUsers{
		Users: res,
		Group: orgUsersGroup.transformToOrgUsersGroup(),
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

// for each organization, fetch the last created billing_subscription entry
func (r OrgUsersRepository) prepareDataQuery(orgID string, rql *rql.Query) (string, []interface{}, error) {
	baseQ := r.getBaseQ(orgID)

	withFilterQ, err := r.addRQLFiltersInQuery(baseQ, rql)
	if err != nil {
		return "", nil, fmt.Errorf("addRQLFiltersInQuery: %w", err)
	}

	withFilterAndSearchQ, err := r.addRQLSearchInQuery(withFilterQ, rql)
	if err != nil {
		return "", nil, fmt.Errorf("addRQLSearchInQuery: %w", err)
	}

	withSortAndFilterAndSearchQ, err := r.addRQLSortInQuery(withFilterAndSearchQ, rql)
	if err != nil {
		return "", nil, fmt.Errorf("addRQLSortInQuery: %w", err)
	}

	//todo: add prepared true
	return withSortAndFilterAndSearchQ.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)).ToSQL()
}

// prepare a query by joining policy, users and roles tables
// combines all roles of a user as a comma separated string
func (r OrgUsersRepository) getBaseQ(orgID string) *goqu.SelectDataset {
	querySelects := []interface{}{
		goqu.I(TABLE_POLICIES + "." + COLUMN_RESOURCE_ID).As(COLUMN_ORG_ID),
		goqu.I(TABLE_USERS + "." + COLUMN_ID).As(COLUMN_ID),
		goqu.I(TABLE_USERS + "." + COLUMN_NAME).As(COLUMN_NAME),
		goqu.I(TABLE_USERS + "." + COLUMN_TITLE).As(COLUMN_TITLE),
		goqu.I(TABLE_USERS + "." + COLUMN_EMAIL).As(COLUMN_EMAIL),
		goqu.I(TABLE_USERS + "." + COLUMN_STATE).As(COLUMN_STATE),
		// goqu.I(TABLE_USERS + "." + COLUMN_AVATAR).As(COLUMN_AVATAR),
		goqu.MIN(goqu.I(TABLE_POLICIES + "." + COLUMN_CREATED_AT)).As(COLUMN_ORG_JOINED_DATE), // Earliest policy creation date
		goqu.L("STRING_AGG(?, ', ')", goqu.I(TABLE_ROLES+"."+COLUMN_NAME)).As(COLUMN_ROLE_NAMES),
		goqu.L("STRING_AGG(COALESCE(?, ''), ', ')", goqu.I(TABLE_ROLES+"."+COLUMN_TITLE)).As(COLUMN_ROLE_TITLES),
		goqu.L("STRING_AGG(?, ', ')", goqu.I(TABLE_ROLES+"."+COLUMN_ID).Cast("TEXT")).As(COLUMN_ROLE_IDS),
	}

	usersWithRolesQuery := dialect.From(TABLE_POLICIES).
		Join(
			goqu.T(TABLE_USERS),
			goqu.On(goqu.I(TABLE_USERS+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID))),
		).
		LeftJoin(
			goqu.T(TABLE_ROLES),
			goqu.On(goqu.I(TABLE_ROLES+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID))),
		).
		Where(
			goqu.C(COLUMN_RESOURCE_ID).Eq(orgID),
			goqu.C(COLUMN_RESOURCE_TYPE).Eq("app/organization"),
			goqu.C(COLUMN_PRINCIPAL_TYPE).Eq("app/user"),
			goqu.I(TABLE_USERS+"."+COLUMN_DELETED_AT).IsNull(),
			goqu.I(TABLE_ROLES+"."+COLUMN_DELETED_AT).IsNull(),
		).
		Select(querySelects...).
		GroupBy(
			goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID),
			goqu.I(TABLE_USERS+"."+COLUMN_ID),
		)

	return usersWithRolesQuery
}

func (r OrgUsersRepository) prepareGroupByQuery(orgID string, rql *rql.Query) (string, []interface{}, error) {
	groupByQuerySelects := []interface{}{
        goqu.I(rql.GroupBy[0]).As(COLUMN_VALUES),
        goqu.COUNT(goqu.DISTINCT(goqu.I(TABLE_USERS + "." + COLUMN_ID))).As(COLUMN_COUNT),
    }

    usersWithRoleGroupByQ := dialect.From(TABLE_POLICIES).
        Join(
            goqu.T(TABLE_USERS),
            goqu.On(goqu.I(TABLE_USERS+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID))),
        ).
        LeftJoin(
            goqu.T(TABLE_ROLES),
            goqu.On(goqu.I(TABLE_ROLES+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID))),
        ).
        Where(
            goqu.C(COLUMN_RESOURCE_ID).Eq(orgID),
            goqu.C(COLUMN_RESOURCE_TYPE).Eq("app/organization"),
            goqu.C(COLUMN_PRINCIPAL_TYPE).Eq("app/user"),
            goqu.I(TABLE_USERS + "." + COLUMN_DELETED_AT).IsNull(),
            goqu.I(TABLE_ROLES + "." + COLUMN_DELETED_AT).IsNull(),
        )

	switch rql.GroupBy[0] {
	case "state":
		groupByQuerySelects[0] = goqu.I("users.state").As(COLUMN_VALUES)
		usersWithRoleGroupByQ = usersWithRoleGroupByQ.GroupBy("users.state").As(rql.GroupBy[0])
	default:
		usersWithRoleGroupByQ = usersWithRoleGroupByQ.GroupBy(rql.GroupBy[0])
	}

	return usersWithRoleGroupByQ.Select(groupByQuerySelects...).ToSQL()
}

func (r OrgUsersRepository) addRQLFiltersInQuery(query *goqu.SelectDataset, rqlInput *rql.Query) (*goqu.SelectDataset, error) {
	supportedFilters := []string{
		COLUMN_ID,
		COLUMN_NAME,
		COLUMN_TITLE,
		COLUMN_EMAIL,
		COLUMN_STATE,
		COLUMN_ORG_JOINED_DATE,
		COLUMN_ROLE_NAMES,
		COLUMN_ROLE_TITLES,
		COLUMN_ROLE_IDS,
	}

	for _, filter := range rqlInput.Filters {
		if !slices.Contains(supportedFilters, filter.Name) {
			return nil, fmt.Errorf("%s is not supported in filters", filter.Name)
		}
		datatype, err := rql.GetDataTypeOfField(filter.Name, svc.AggregatedUser{})
		if err != nil {
			return query, err
		}
		switch datatype {
		case "string":
			query = processStringDataType(filter, query)
		case "number":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(float32)},
			})
		case "bool":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(bool)},
			})
		case "datetime":
			query = query.Where(goqu.Ex{
				filter.Name: goqu.Op{filter.Operator: filter.Value.(string)},
			})
		}
	}
	return query, nil
}

func (r OrgUsersRepository) addRQLSearchInQuery(query *goqu.SelectDataset, rql *rql.Query) (*goqu.SelectDataset, error) {
	rqlSearchSupportedColumns := []string{
		COLUMN_ID,
		COLUMN_NAME,
		COLUMN_TITLE,
		COLUMN_EMAIL,
		COLUMN_STATE,
		COLUMN_ORG_JOINED_DATE,
		COLUMN_ROLE_NAMES,
		COLUMN_ROLE_TITLES,
	}

	searchExpressions := make([]goqu.Expression, 0)
	if rql.Search != "" {
		searchPattern := "%" + rql.Search + "%"
		for _, col := range rqlSearchSupportedColumns {
			searchExpressions = append(searchExpressions,
				goqu.Cast(goqu.I(col), "TEXT").ILike(searchPattern),
			)
		}
	}
	return query.Where(goqu.Or(searchExpressions...)), nil
}

func (r OrgUsersRepository) addRQLSortInQuery(query *goqu.SelectDataset, rql *rql.Query) (*goqu.SelectDataset, error) {
	// If there is a group by parameter added then sort the result
	// by group_by first key in asc order by default before any other sort column
	if len(rql.GroupBy) > 0 {
		query = query.OrderAppend(goqu.C(rql.GroupBy[0]).Asc())
	}

	for _, sortItem := range rql.Sort {
		switch sortItem.Order {
		case "asc":
			query = query.OrderAppend(goqu.C(sortItem.Name).Asc())
		case "desc":
			query = query.OrderAppend(goqu.C(sortItem.Name).Desc())
		default:
		}
	}
	return query, nil
}
