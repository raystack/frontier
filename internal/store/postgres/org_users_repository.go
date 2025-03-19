package postgres

import (
	"context"
	"database/sql"

	"fmt"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	OPERATOR_EQ     = "eq"
	OPERATOR_NOT_EQ = "neq"
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
	fmt.Println(dataQuery)
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

// prepare a query by joining policy, users and roles tables
// combines all roles of a user as a comma separated string
func (r OrgUsersRepository) prepareDataQuery(orgID string, input *rql.Query) (string, []interface{}, error) {
	querySelects := []interface{}{
		goqu.I(TABLE_POLICIES + "." + COLUMN_RESOURCE_ID).As(COLUMN_ORG_ID),
		goqu.I(TABLE_USERS + "." + COLUMN_ID).As(COLUMN_ID),
		goqu.I(TABLE_USERS + "." + COLUMN_NAME).As(COLUMN_NAME),
		goqu.I(TABLE_USERS + "." + COLUMN_TITLE).As(COLUMN_TITLE),
		goqu.I(TABLE_USERS + "." + COLUMN_EMAIL).As(COLUMN_EMAIL),
		goqu.I(TABLE_USERS + "." + COLUMN_STATE).As(COLUMN_STATE),
		goqu.MIN(goqu.I(TABLE_POLICIES + "." + COLUMN_POLICY_CREATED_AT)).As(COLUMN_ORG_JOINED_DATE),
		goqu.L("STRING_AGG(?, ', ')", goqu.I(TABLE_ROLES+"."+COLUMN_NAME)).As(COLUMN_ROLE_NAMES),
		goqu.L("STRING_AGG(COALESCE(?, ''), ', ')", goqu.I(TABLE_ROLES+"."+COLUMN_TITLE)).As(COLUMN_ROLE_TITLES),
		goqu.L("STRING_AGG(?, ', ')", goqu.I(TABLE_ROLES+"."+COLUMN_ID).Cast("TEXT")).As(COLUMN_ROLE_IDS),
	}

	whereConditions := []goqu.Expression{
		goqu.C(COLUMN_RESOURCE_ID).Eq(orgID),
		goqu.C(COLUMN_RESOURCE_TYPE).Eq("app/organization"),
		goqu.C(COLUMN_PRINCIPAL_TYPE).Eq("app/user"),
		goqu.I(TABLE_USERS + "." + COLUMN_DELETED_AT).IsNull(),
		goqu.I(TABLE_ROLES + "." + COLUMN_DELETED_AT).IsNull(),
	}

	supportedFilters := []string{
		COLUMN_NAME,
		COLUMN_TITLE,
		COLUMN_EMAIL,
		COLUMN_STATE,
		COLUMN_ORG_JOINED_DATE,
		COLUMN_ROLE_NAMES,
		COLUMN_ROLE_TITLES,
		COLUMN_ROLE_IDS,
	}

	if len(input.Filters) != 0 {
		for _, filter := range input.Filters {
			if !slices.Contains(supportedFilters, filter.Name) {
				return "", nil, fmt.Errorf("%s is not supported in filters", filter.Name)
			}

			// Handle non-role filters in the outer query
			if !slices.Contains([]string{COLUMN_ROLE_NAMES, COLUMN_ROLE_TITLES, COLUMN_ROLE_IDS}, filter.Name) {
				supportedStringOperators := []string{"eq", "neq", "like", "in", "notin", "notlike", "empty", "notempty"}

				if !slices.Contains(supportedStringOperators, filter.Operator) {
					return "", nil, fmt.Errorf("unsupported operator %s for non-role filter column %s", filter.Operator, filter.Name)
				}

				var columnName string
				switch filter.Name {
				case COLUMN_NAME, COLUMN_TITLE, COLUMN_EMAIL, COLUMN_STATE:
					columnName = fmt.Sprintf("%s.%s", TABLE_USERS, filter.Name)
				case COLUMN_ORG_JOINED_DATE:
					columnName = fmt.Sprintf("%s.%s", TABLE_POLICIES, COLUMN_POLICY_CREATED_AT)
				}

				switch filter.Operator {
				case OPERATOR_EMPTY:
					whereConditions = append(whereConditions, goqu.L(fmt.Sprintf("coalesce(%s, '') = ''", columnName)))
				case OPERATOR_NOT_EMPTY:
					whereConditions = append(whereConditions, goqu.L(fmt.Sprintf("coalesce(%s, '') != ''", columnName)))
				case OPERATOR_IN, OPERATOR_NOT_IN:
					whereConditions = append(whereConditions,
						goqu.Ex{columnName: goqu.Op{filter.Operator: strings.Split(filter.Value.(string), ",")}})
				case OPERATOR_LIKE:
					searchPattern := "%" + filter.Value.(string) + "%"
					whereConditions = append(whereConditions, goqu.Cast(goqu.I(columnName), "TEXT").ILike(searchPattern))
				case OPERATOR_NOT_LIKE:
					searchPattern := "%" + filter.Value.(string) + "%"
					whereConditions = append(whereConditions, goqu.Cast(goqu.I(columnName), "TEXT").ILike(searchPattern))
				default: // eq, neq
					whereConditions = append(whereConditions, goqu.Ex{columnName: goqu.Op{filter.Operator: filter.Value.(string)}})
				}
			} else {
				// Handle role-related filters using subqueries
				if !slices.Contains([]string{OPERATOR_EQ, OPERATOR_NOT_EQ}, filter.Operator) {
					return "", nil, fmt.Errorf("unsupported operator %s for role filter column %s", filter.Operator, filter.Name)
				}

				var columnName string
				switch filter.Name {
				case COLUMN_ROLE_NAMES:
					columnName = COLUMN_NAME
				case COLUMN_ROLE_TITLES:
					columnName = COLUMN_TITLE
				case COLUMN_ROLE_IDS:
					columnName = COLUMN_ID
				}

				if filter.Operator == OPERATOR_EQ {
					roleSubquery := dialect.From(TABLE_POLICIES).
						Join(
							goqu.T(TABLE_ROLES),
							goqu.On(goqu.I(TABLE_ROLES+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID))),
						).
						Where(
							goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(goqu.I(TABLE_USERS+"."+COLUMN_ID)),
							goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).Eq(orgID),
							goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_TYPE).Eq("app/organization"),
							goqu.I(TABLE_ROLES+"."+columnName).Eq(filter.Value),
						).
						Select(goqu.L("1")).
						Limit(1)

					whereConditions = append(whereConditions, goqu.L("EXISTS ?", roleSubquery))
				} else {
					roleNotExistsSubquery := dialect.From(TABLE_POLICIES).
						Join(
							goqu.T(TABLE_ROLES),
							goqu.On(goqu.I(TABLE_ROLES+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID))),
						).
						Where(
							goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(goqu.I(TABLE_USERS+"."+COLUMN_ID)),
							goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).Eq(orgID),
							goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_TYPE).Eq("app/organization"),
							goqu.I(TABLE_ROLES+"."+columnName).Eq(filter.Value),
						).
						Select(goqu.L("1")).
						Limit(1)

					hasRoleSubquery := dialect.From(TABLE_POLICIES).
						Join(
							goqu.T(TABLE_ROLES),
							goqu.On(goqu.I(TABLE_ROLES+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID))),
						).
						Where(
							goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(goqu.I(TABLE_USERS+"."+COLUMN_ID)),
							goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).Eq(orgID),
							goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_TYPE).Eq("app/organization"),
						).
						Select(goqu.L("1")).
						Limit(1)

					whereConditions = append(whereConditions,
						goqu.L("NOT EXISTS ?", roleNotExistsSubquery),
						goqu.L("EXISTS ?", hasRoleSubquery),
					)
				}
			}
		}
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
		Where(whereConditions...).
		Select(querySelects...).
		GroupBy(
			goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID),
			goqu.I(TABLE_USERS+"."+COLUMN_ID),
			goqu.I(TABLE_USERS+"."+COLUMN_NAME),
			goqu.I(TABLE_USERS+"."+COLUMN_TITLE),
			goqu.I(TABLE_USERS+"."+COLUMN_EMAIL),
			goqu.I(TABLE_USERS+"."+COLUMN_STATE),
			goqu.I(TABLE_USERS+"."+COLUMN_CREATED_AT),
			goqu.I(TABLE_USERS+"."+COLUMN_UPDATED_AT),
		)

	return usersWithRolesQuery.ToSQL()
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
			goqu.I(TABLE_USERS+"."+COLUMN_DELETED_AT).IsNull(),
			goqu.I(TABLE_ROLES+"."+COLUMN_DELETED_AT).IsNull(),
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
