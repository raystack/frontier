package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	svc "github.com/raystack/frontier/core/aggregates/userprojects"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	// Table aliases
	TABLE_ALIAS_PROJECT      = "p"
	TABLE_ALIAS_SUB_PROJECT  = "p2"
	TABLE_ALIAS_POLICIES     = "pol"
	TABLE_ALIAS_SUB_POLICIES = "pol2"
	TABLE_ALIAS_USERS        = "u"

	// Resource and Principal types
	TYPE_PROJECT = "app/project"
	TYPE_USER    = "app/user"

	// Column aliases for projections
	COLUMN_PROJECT_NAME    = "project_name"
	COLUMN_PROJECT_TITLE   = "project_title"
	COLUMN_PROJECT_CREATED = "project_created_on"
	COLUMN_USER_IDS        = "user_ids"
	COLUMN_USER_NAMES      = "user_names"
	COLUMN_USER_TITLES     = "user_titles"
	COLUMN_USER_AVATARS    = "user_avatars"
)

type UserProjectsRepository struct {
	dbc *db.Client
}

type UserProjects struct {
	ProjectID        sql.NullString `db:"project_id"`
	ProjectTitle     sql.NullString `db:"project_title"`
	ProjectName      sql.NullString `db:"project_name"`
	ProjectCreatedOn sql.NullTime   `db:"project_created_on"`
	UserIDs          pq.StringArray `db:"user_ids"`
	UserNames        pq.StringArray `db:"user_names"`
	UserTitles       pq.StringArray `db:"user_titles"`
	UserAvatars      pq.StringArray `db:"user_avatars"`
	OrgID            sql.NullString `db:"org_id"`
	UserID           sql.NullString `db:"user_id"`
}

func (p *UserProjects) transformToAggregatedProject() svc.AggregatedProject {
	return svc.AggregatedProject{
		ProjectID:    p.ProjectID.String,
		ProjectTitle: p.ProjectTitle.String,
		ProjectName:  p.ProjectName.String,
		CreatedOn:    p.ProjectCreatedOn.Time,
		UserIDs:      []string(p.UserIDs),
		UserNames:    []string(p.UserNames),
		UserTitles:   []string(p.UserTitles),
		UserAvatars:  []string(p.UserAvatars),
		OrgID:        p.OrgID.String,
		UserID:       p.UserID.String,
	}
}

func NewUserProjectsRepository(dbc *db.Client) *UserProjectsRepository {
	return &UserProjectsRepository{
		dbc: dbc,
	}
}
func (r UserProjectsRepository) Search(ctx context.Context, userID string, orgID string, rql *rql.Query) (svc.UserProjects, error) {
	query, err := r.prepareDataQuery(userID, orgID, rql)
	if err != nil {
		return svc.UserProjects{}, err
	}

	dataQuery, params, err := query.ToSQL()
	if err != nil {
		return svc.UserProjects{}, err
	}

	var userProjects []UserProjects

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, "projects", "GetUserProjects", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &userProjects, dataQuery, params...)
		})
	})

	if err != nil {
		return svc.UserProjects{}, err
	}

	res := make([]svc.AggregatedProject, 0)
	for _, project := range userProjects {
		transformedProject := project.transformToAggregatedProject()
		transformedProject.OrgID = orgID
		transformedProject.UserID = userID
		res = append(res, transformedProject)
	}

	return svc.UserProjects{
		Projects: res,
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r UserProjectsRepository) prepareDataQuery(userID string, orgID string, rql *rql.Query) (*goqu.SelectDataset, error) {
	query := r.buildBaseQuery(userID, orgID)

	if rql.Search != "" {
		query = r.addSearch(query, rql.Search)
	}

	return query.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)), nil
}

func (r UserProjectsRepository) buildBaseQuery(userID string, orgID string) *goqu.SelectDataset {
	subquery := dialect.From(goqu.T(TABLE_PROJECTS).As(TABLE_ALIAS_SUB_PROJECT)).
		Select(goqu.I(TABLE_ALIAS_SUB_PROJECT+"."+COLUMN_ID)).
		Join(
			goqu.T(TABLE_POLICIES).As(TABLE_ALIAS_SUB_POLICIES),
			goqu.On(goqu.I(TABLE_ALIAS_SUB_PROJECT+"."+COLUMN_ID).Eq(goqu.I(TABLE_ALIAS_SUB_POLICIES+"."+COLUMN_RESOURCE_ID))),
		).
		Where(goqu.And(
			goqu.I(TABLE_ALIAS_SUB_PROJECT+"."+COLUMN_ORG_ID).Eq(orgID),
			goqu.I(TABLE_ALIAS_SUB_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(userID),
			goqu.I(TABLE_ALIAS_SUB_POLICIES+"."+COLUMN_RESOURCE_TYPE).Eq(TYPE_PROJECT),
			goqu.I(TABLE_ALIAS_SUB_POLICIES+"."+COLUMN_PRINCIPAL_TYPE).Eq(TYPE_USER),
			goqu.I(TABLE_ALIAS_SUB_POLICIES+"."+COLUMN_DELETED_AT).IsNull(),
		))

	return dialect.From(goqu.T(TABLE_PROJECTS).As(TABLE_ALIAS_PROJECT)).Prepared(true).
		Select(
			goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_ID).As(COLUMN_PROJECT_ID),
			goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_TITLE).As(COLUMN_PROJECT_TITLE),
			goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_NAME).As(COLUMN_PROJECT_NAME),
			goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_CREATED_AT).As(COLUMN_PROJECT_CREATED),
			goqu.L(fmt.Sprintf("array_agg(DISTINCT %s.%s ORDER BY %s.%s)", TABLE_ALIAS_USERS, COLUMN_ID, TABLE_ALIAS_USERS, COLUMN_ID)).As(COLUMN_USER_IDS),
			goqu.L(fmt.Sprintf("array_agg(DISTINCT %s.%s ORDER BY %s.%s)", TABLE_ALIAS_USERS, COLUMN_AVATAR, TABLE_ALIAS_USERS, COLUMN_AVATAR)).As(COLUMN_USER_AVATARS),
			goqu.L(fmt.Sprintf("array_agg(DISTINCT %s.%s ORDER BY %s.%s)", TABLE_ALIAS_USERS, COLUMN_NAME, TABLE_ALIAS_USERS, COLUMN_NAME)).As(COLUMN_USER_NAMES),
			goqu.L(fmt.Sprintf("array_agg(DISTINCT %s.%s ORDER BY %s.%s)", TABLE_ALIAS_USERS, COLUMN_TITLE, TABLE_ALIAS_USERS, COLUMN_TITLE)).As(COLUMN_USER_TITLES),
		).
		Join(
			goqu.T(TABLE_POLICIES).As(TABLE_ALIAS_POLICIES),
			goqu.On(goqu.And(
				goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_ID).Eq(goqu.I(TABLE_ALIAS_POLICIES+"."+COLUMN_RESOURCE_ID)),
				goqu.I(TABLE_ALIAS_POLICIES+"."+COLUMN_RESOURCE_TYPE).Eq(TYPE_PROJECT),
				goqu.I(TABLE_ALIAS_POLICIES+"."+COLUMN_DELETED_AT).IsNull(),
			)),
		).
		Join(
			goqu.T(TABLE_USERS).As(TABLE_ALIAS_USERS),
			goqu.On(goqu.And(
				goqu.I(TABLE_ALIAS_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(goqu.I(TABLE_ALIAS_USERS+"."+COLUMN_ID)),
				goqu.I(TABLE_ALIAS_POLICIES+"."+COLUMN_PRINCIPAL_TYPE).Eq(TYPE_USER),
			)),
		).
		Where(goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_ID).In(subquery)).
		GroupBy(
			goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_ID),
			goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_NAME),
			goqu.I(TABLE_ALIAS_PROJECT+"."+COLUMN_CREATED_AT),
		).
		Order(goqu.I(TABLE_ALIAS_PROJECT + "." + COLUMN_NAME).Asc())
}

func (r UserProjectsRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchPattern := "%" + search + "%"

	baseSubquery := query.As(TABLE_BASE)

	return dialect.From(baseSubquery).Prepared(true).
		Where(
			goqu.Or(
				// Project field searches - only title and name as requested
				goqu.I(TABLE_BASE+"."+COLUMN_PROJECT_TITLE).ILike(searchPattern),
				goqu.I(TABLE_BASE+"."+COLUMN_PROJECT_NAME).ILike(searchPattern),
			),
		)
}
