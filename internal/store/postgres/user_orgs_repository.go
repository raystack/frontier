package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	svc "github.com/raystack/frontier/core/aggregates/userorgs"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	// Organizations Table Columns
	COLUMN_ORG_NAME   = "org_name"
	COLUMN_ORG_TITLE  = "org_title"
	COLUMN_ORG_AVATAR = "org_avatar"

	// Custom Aggregated Columns
	COLUMN_PROJECT_COUNT = "project_count"
	COLUMN_ORG_JOINED_ON = "org_joined_on"

	STATE_ENABLED              = "enabled"
	ALIAS_PROJECT_COUNTS       = "project_counts"
	RESOURCE_TYPE_ORGANIZATION = "app/organization"
)

type UserOrgsRepository struct {
	dbc *db.Client
}

type UserOrgs struct {
	OrgID        sql.NullString `db:"org_id"`
	OrgTitle     sql.NullString `db:"org_title"`
	OrgName      sql.NullString `db:"org_name"`
	OrgAvatar    sql.NullString `db:"org_avatar"`
	ProjectCount sql.NullInt64  `db:"project_count"`
	RoleNames    pq.StringArray `db:"role_names"`
	RoleTitles   pq.StringArray `db:"role_titles"`
	RoleIDs      pq.StringArray `db:"role_ids"`
	OrgJoinedOn  sql.NullTime   `db:"org_joined_on"`
	UserID       sql.NullString `db:"principal_id"`
}

type UserOrgsGroup struct {
	Name sql.NullString      `db:"name"`
	Data []UserOrgsGroupData `db:"data"`
}

type UserOrgsGroupData struct {
	Name  sql.NullString `db:"values"`
	Count int            `db:"count"`
}

func (u *UserOrgs) transformToAggregatedUserOrg() svc.AggregatedUserOrganization {
	return svc.AggregatedUserOrganization{
		OrgID:        u.OrgID.String,
		OrgTitle:     u.OrgTitle.String,
		OrgName:      u.OrgName.String,
		OrgAvatar:    u.OrgAvatar.String,
		ProjectCount: u.ProjectCount.Int64,
		RoleNames:    u.RoleNames,
		RoleTitles:   u.RoleTitles,
		RoleIDs:      u.RoleIDs,
		OrgJoinedOn:  u.OrgJoinedOn.Time,
		UserID:       u.UserID.String,
	}
}

func NewUserOrgsRepository(dbc *db.Client) *UserOrgsRepository {
	return &UserOrgsRepository{
		dbc: dbc,
	}
}

func (r UserOrgsRepository) Search(ctx context.Context, principalID string, rql *rql.Query) (svc.UserOrgs, error) {
	// Check for unsupported operations
	if len(rql.Filters) > 0 {
		return svc.UserOrgs{}, fmt.Errorf("%w: filters not supported", ErrBadInput)
	}
	if len(rql.Sort) > 0 {
		return svc.UserOrgs{}, fmt.Errorf("%w: sorting not supported", ErrBadInput)
	}
	if rql.Search != "" {
		return svc.UserOrgs{}, fmt.Errorf("%w: search not supported", ErrBadInput)
	}
	if len(rql.GroupBy) > 0 {
		return svc.UserOrgs{}, fmt.Errorf("%w: group_by not supported", ErrBadInput)
	}
	if rql.Limit != 0 {
		return svc.UserOrgs{}, fmt.Errorf("%w: limit not supported", ErrBadInput)
	}
	if rql.Offset != 0 {
		return svc.UserOrgs{}, fmt.Errorf("%w: offset not supported", ErrBadInput)
	}

	query, params, err := r.buildBaseQuery(principalID)
	if err != nil {
		return svc.UserOrgs{}, err
	}

	var userOrgs []UserOrgs
	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_POLICIES, "GetUserOrgs", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &userOrgs, query, params...)
		})
	})

	if err != nil {
		return svc.UserOrgs{}, err
	}

	// Transform the results
	organizations := make([]svc.AggregatedUserOrganization, 0, len(userOrgs))
	for _, org := range userOrgs {
		organizations = append(organizations, org.transformToAggregatedUserOrg())
	}

	return svc.UserOrgs{
		Organizations: organizations,
		Group: svc.Group{
			Name: "",
			Data: []svc.GroupData{},
		},
		Pagination: svc.Page{
			Limit:  rql.Limit,
			Offset: rql.Offset,
		},
	}, nil
}

func (r UserOrgsRepository) buildBaseQuery(principalID string) (string, []interface{}, error) {
	projectCountSubquery := dialect.From(TABLE_PROJECTS).
		Select(
			goqu.I(COLUMN_ORG_ID),
			goqu.COUNT(COLUMN_ID).As(COLUMN_PROJECT_COUNT),
		).
		Where(
			goqu.I(COLUMN_DELETED_AT).IsNull(),
			goqu.I(COLUMN_STATE).Eq(STATE_ENABLED),
		).
		GroupBy(COLUMN_ORG_ID).
		As(ALIAS_PROJECT_COUNTS)

	querySelects := []interface{}{
		goqu.I(TABLE_POLICIES + "." + COLUMN_PRINCIPAL_ID),
		goqu.I(TABLE_POLICIES + "." + COLUMN_RESOURCE_ID).As(COLUMN_ORG_ID),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_NAME).As(COLUMN_ORG_NAME),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_TITLE).As(COLUMN_ORG_TITLE),
		goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_AVATAR).As(COLUMN_ORG_AVATAR),
		goqu.MIN(goqu.I(TABLE_POLICIES + "." + COLUMN_CREATED_AT)).As(COLUMN_ORG_JOINED_ON),
		goqu.L("ARRAY_AGG(?)", goqu.I(TABLE_ROLES+"."+COLUMN_NAME)).As(COLUMN_ROLE_NAMES),
		goqu.L("ARRAY_AGG(?)", goqu.I(TABLE_ROLES+"."+COLUMN_TITLE)).As(COLUMN_ROLE_TITLES),
		goqu.L("ARRAY_AGG(?)", goqu.I(TABLE_ROLES+"."+COLUMN_ID)).As(COLUMN_ROLE_IDS),
		goqu.COALESCE(goqu.I(ALIAS_PROJECT_COUNTS+"."+COLUMN_PROJECT_COUNT), 0).As(COLUMN_PROJECT_COUNT),
	}

	baseConditions := []goqu.Expression{
		goqu.I(TABLE_POLICIES + "." + COLUMN_RESOURCE_TYPE).Eq(RESOURCE_TYPE_ORGANIZATION),
		goqu.I(TABLE_POLICIES + "." + COLUMN_PRINCIPAL_ID).Eq(principalID),
	}

	return dialect.From(TABLE_POLICIES).Prepared(true).
		Join(
			goqu.T(TABLE_ROLES),
			goqu.On(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID).Eq(goqu.I(TABLE_ROLES+"."+COLUMN_ID))),
		).
		Join(
			goqu.T(TABLE_ORGANIZATIONS),
			goqu.On(goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).Eq(goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID))),
		).
		Join(
			goqu.T(TABLE_USERS),
			goqu.On(goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(goqu.I(TABLE_USERS+"."+COLUMN_ID))),
		).
		LeftJoin(
			projectCountSubquery,
			goqu.On(goqu.I(ALIAS_PROJECT_COUNTS+"."+COLUMN_ORG_ID).Eq(goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_ID))),
		).
		Where(baseConditions...).
		Select(querySelects...).
		GroupBy(
			goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID),
			goqu.I(TABLE_USERS+"."+COLUMN_EMAIL),
			goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_NAME),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_TITLE),
			goqu.I(TABLE_ORGANIZATIONS+"."+COLUMN_AVATAR),
			goqu.I(ALIAS_PROJECT_COUNTS+"."+COLUMN_PROJECT_COUNT),
		).
		Order(goqu.I(TABLE_ORGANIZATIONS + "." + COLUMN_NAME).Asc()).ToSQL()
}
