package postgres

import (
	"context"
	"database/sql"

	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	svc "github.com/raystack/frontier/core/aggregates/orgusers"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

var ErrBadInput = errors.New("bad input")

func wrapBadOperatorError(operator, column string) error {
	return fmt.Errorf("%w: unsupported operator %s for column %s", ErrBadInput, operator, column)
}

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
	RoleNames   pq.StringArray `db:"role_names"`
	RoleTitles  pq.StringArray `db:"role_titles"`
	RoleIDs     pq.StringArray `db:"role_ids"`
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
		RoleNames:   u.RoleNames,
		RoleTitles:  u.RoleTitles,
		RoleIDs:     u.RoleIDs,
		OrgID:       u.OrgID.String,
		OrgJoinedAt: u.OrgJoinedAt.Time,
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
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

// prepare a query by joining policy, users and roles tables
// combines all roles of a user as a comma separated string
func (r OrgUsersRepository) prepareDataQuery(orgID string, input *rql.Query) (string, []interface{}, error) {
	baseQuery := r.buildBaseQuery(orgID)

	if err := r.validateFilters(input.Filters); err != nil {
		return "", nil, err
	}

	queryWithFilters, err := r.applyFilters(baseQuery, orgID, input.Filters)
	if err != nil {
		return "", nil, err
	}

	queryWithSearch, err := r.addRQLSearchInQuery(queryWithFilters, input)
	if err != nil {
		return "", nil, err
	}

	queryWithSort, err := r.addRQLSortInQuery(queryWithSearch, input)
	if err != nil {
		return "", nil, err
	}

	return queryWithSort.Offset(uint(input.Offset)).Limit(uint(input.Limit)).ToSQL()
}

func (r OrgUsersRepository) buildBaseQuery(orgID string) *goqu.SelectDataset {
	querySelects := []interface{}{
		goqu.I(TABLE_POLICIES + "." + COLUMN_RESOURCE_ID).As(COLUMN_ORG_ID),
		goqu.I(TABLE_USERS + "." + COLUMN_ID).As(COLUMN_ID),
		goqu.I(TABLE_USERS + "." + COLUMN_NAME).As(COLUMN_NAME),
		goqu.I(TABLE_USERS + "." + COLUMN_TITLE).As(COLUMN_TITLE),
		goqu.I(TABLE_USERS + "." + COLUMN_EMAIL).As(COLUMN_EMAIL),
		goqu.I(TABLE_USERS + "." + COLUMN_STATE).As(COLUMN_STATE),
		goqu.I(TABLE_USERS + "." + COLUMN_AVATAR).As(COLUMN_AVATAR),
		goqu.MIN(goqu.I(TABLE_POLICIES + "." + COLUMN_POLICY_CREATED_AT)).As(COLUMN_ORG_JOINED_DATE),
		goqu.L("ARRAY_AGG(?)", goqu.I(TABLE_ROLES+"."+COLUMN_NAME)).As(COLUMN_ROLE_NAMES),
		goqu.L("ARRAY_AGG(COALESCE(?, ''))", goqu.I(TABLE_ROLES+"."+COLUMN_TITLE)).As(COLUMN_ROLE_TITLES),
		goqu.L("ARRAY_AGG(?)", goqu.I(TABLE_ROLES+"."+COLUMN_ID).Cast("TEXT")).As(COLUMN_ROLE_IDS),
	}

	baseConditions := []goqu.Expression{
		goqu.I(TABLE_POLICIES + "." + COLUMN_RESOURCE_ID).Eq(orgID),
		goqu.I(TABLE_POLICIES + "." + COLUMN_RESOURCE_TYPE).Eq("app/organization"),
		goqu.I(TABLE_POLICIES + "." + COLUMN_PRINCIPAL_TYPE).Eq("app/user"),
		goqu.I(TABLE_USERS + "." + COLUMN_DELETED_AT).IsNull(),
		goqu.I(TABLE_ROLES + "." + COLUMN_DELETED_AT).IsNull(),
	}

	return dialect.From(TABLE_POLICIES).Prepared(true).
		Join(
			goqu.T(TABLE_USERS),
			goqu.On(goqu.I(TABLE_USERS+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID))),
		).
		LeftJoin(
			goqu.T(TABLE_ROLES),
			goqu.On(goqu.I(TABLE_ROLES+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID))),
		).
		Where(baseConditions...).
		Select(querySelects...).
		GroupBy(r.getGroupByColumns()...)
}

func (r OrgUsersRepository) getGroupByColumns() []interface{} {
	return []interface{}{
		goqu.I(TABLE_POLICIES + "." + COLUMN_RESOURCE_ID),
		goqu.I(TABLE_USERS + "." + COLUMN_ID),
		goqu.I(TABLE_USERS + "." + COLUMN_NAME),
		goqu.I(TABLE_USERS + "." + COLUMN_TITLE),
		goqu.I(TABLE_USERS + "." + COLUMN_EMAIL),
		goqu.I(TABLE_USERS + "." + COLUMN_STATE),
		goqu.I(TABLE_USERS + "." + COLUMN_CREATED_AT),
		goqu.I(TABLE_USERS + "." + COLUMN_UPDATED_AT),
	}
}

func (r OrgUsersRepository) validateFilters(filters []rql.Filter) error {
	supportedFilters := []string{
		COLUMN_NAME, COLUMN_TITLE, COLUMN_EMAIL, COLUMN_STATE,
		COLUMN_ORG_JOINED_DATE, COLUMN_ROLE_NAMES, COLUMN_ROLE_TITLES, COLUMN_ROLE_IDS,
	}

	for _, filter := range filters {
		if !slices.Contains(supportedFilters, filter.Name) {
			return fmt.Errorf("%s is not supported in filters", filter.Name)
		}
	}
	return nil
}

func (r OrgUsersRepository) applyFilters(query *goqu.SelectDataset, orgID string, filters []rql.Filter) (*goqu.SelectDataset, error) {
	conditions := []goqu.Expression{}

	for _, filter := range filters {
		if isRoleFilter(filter.Name) {
			roleCondition, err := r.buildRoleFilterCondition(orgID, filter)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, roleCondition...)
		} else {
			condition, err := r.buildNonRoleFilterCondition(filter)
			if err != nil {
				return nil, err
			}
			conditions = append(conditions, condition)
		}
	}

	return query.Where(conditions...), nil
}

func isRoleFilter(columnName string) bool {
	roleFilters := []string{COLUMN_ROLE_NAMES, COLUMN_ROLE_TITLES, COLUMN_ROLE_IDS}
	return slices.Contains(roleFilters, columnName)
}

func (r OrgUsersRepository) buildRoleFilterCondition(orgID string, filter rql.Filter) ([]goqu.Expression, error) {
	if !slices.Contains([]string{OPERATOR_EQ, OPERATOR_NOT_EQ}, filter.Operator) {
		return nil, wrapBadOperatorError(filter.Operator, filter.Name)
	}

	columnName := r.getRoleColumnName(filter.Name)

	if filter.Operator == OPERATOR_EQ {
		roleSubquery := r.buildRoleExistsSubquery(orgID, columnName, filter.Value)
		return []goqu.Expression{goqu.L("EXISTS ?", roleSubquery)}, nil
	}

	roleNotExistsSubquery := r.buildRoleExistsSubquery(orgID, columnName, filter.Value)
	hasRoleSubquery := r.buildHasAnyRoleSubquery(orgID)

	return []goqu.Expression{
		goqu.L("NOT EXISTS ?", roleNotExistsSubquery),
		goqu.L("EXISTS ?", hasRoleSubquery),
	}, nil
}

func (r OrgUsersRepository) getRoleColumnName(filterName string) string {
	switch filterName {
	case COLUMN_ROLE_NAMES:
		return COLUMN_NAME
	case COLUMN_ROLE_TITLES:
		return COLUMN_TITLE
	case COLUMN_ROLE_IDS:
		return COLUMN_ID
	default:
		return ""
	}
}

func (r OrgUsersRepository) buildRoleExistsSubquery(orgID string, columnName string, value interface{}) *goqu.SelectDataset {
	return dialect.From(TABLE_POLICIES).Prepared(true).
		Join(
			goqu.T(TABLE_ROLES),
			goqu.On(goqu.I(TABLE_ROLES+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID))),
		).
		Where(
			goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(goqu.I(TABLE_USERS+"."+COLUMN_ID)),
			goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).Eq(orgID),
			goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_TYPE).Eq("app/organization"),
			goqu.I(TABLE_ROLES+"."+columnName).Eq(value),
		).
		Select(goqu.L("1")).
		Limit(1)
}

func (r OrgUsersRepository) buildHasAnyRoleSubquery(orgID string) *goqu.SelectDataset {
	return dialect.From(TABLE_POLICIES).Prepared(true).
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
}

func (r OrgUsersRepository) buildNonRoleFilterCondition(filter rql.Filter) (goqu.Expression, error) {
	supportedStringOperators := []string{"eq", "neq", "like", "in", "notin", "notlike", "empty", "notempty"}
	if !slices.Contains(supportedStringOperators, filter.Operator) {
		return nil, wrapBadOperatorError(filter.Operator, filter.Name)
	}

	columnName := r.getNonRoleColumnName(filter.Name)

	switch filter.Operator {
	case "empty":
		return goqu.L(fmt.Sprintf("coalesce(%s, '') = ''", columnName)), nil
	case "notempty":
		return goqu.L(fmt.Sprintf("coalesce(%s, '') != ''", columnName)), nil
	case "in", "notin":
		return goqu.Ex{columnName: goqu.Op{filter.Operator: strings.Split(filter.Value.(string), ",")}}, nil
	case "like":
		searchPattern := "%" + filter.Value.(string) + "%"
		return goqu.Cast(goqu.I(columnName), "TEXT").ILike(searchPattern), nil
	case "notlike":
		searchPattern := "%" + filter.Value.(string) + "%"
		return goqu.Cast(goqu.I(columnName), "TEXT").NotILike(searchPattern), nil
	default: // eq, neq
		return goqu.Ex{columnName: goqu.Op{filter.Operator: filter.Value.(string)}}, nil
	}
}

func (r OrgUsersRepository) getNonRoleColumnName(filterName string) string {
	switch filterName {
	case COLUMN_NAME, COLUMN_TITLE, COLUMN_EMAIL, COLUMN_STATE:
		return fmt.Sprintf("%s.%s", TABLE_USERS, filterName)
	case COLUMN_ORG_JOINED_DATE:
		return fmt.Sprintf("%s.%s", TABLE_POLICIES, COLUMN_POLICY_CREATED_AT)
	default:
		return filterName
	}
}

func (r OrgUsersRepository) addRQLSearchInQuery(query *goqu.SelectDataset, rql *rql.Query) (*goqu.SelectDataset, error) {
	if rql.Search == "" {
		return query, nil
	}

	type columnMapping struct {
		name         string
		qualifiedCol string
	}

	// Define ordered column mappings
	columnMappings := []columnMapping{
		// User table columns
		{name: COLUMN_NAME, qualifiedCol: TABLE_USERS + "." + COLUMN_NAME},
		{name: COLUMN_TITLE, qualifiedCol: TABLE_USERS + "." + COLUMN_TITLE},
		{name: COLUMN_EMAIL, qualifiedCol: TABLE_USERS + "." + COLUMN_EMAIL},
		{name: COLUMN_STATE, qualifiedCol: TABLE_USERS + "." + COLUMN_STATE},
	}

	searchExpressions := make([]goqu.Expression, 0, len(columnMappings))
	searchPattern := "%" + rql.Search + "%"

	for _, mapping := range columnMappings {
		searchExpressions = append(searchExpressions,
			goqu.Cast(goqu.I(mapping.qualifiedCol), "TEXT").ILike(searchPattern),
		)
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
