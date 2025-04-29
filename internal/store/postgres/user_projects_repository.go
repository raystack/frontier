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
	COLUMN_PROJECT_TITLE   = "project_title"
	COLUMN_PROJECT_NAME    = "project_name"
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
	query := r.buildBaseQuery(userID, orgID)

	if rql.Search != "" {
		query = r.addSearch(query, rql.Search)
	}

	dataQuery, params, err := query.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)).ToSQL()
	fmt.Println(dataQuery)
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
		res = append(res, project.transformToAggregatedProject())
	}

	return svc.UserProjects{
		Projects: res,
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r UserProjectsRepository) buildBaseQuery(userID string, orgID string) *goqu.SelectDataset {
    subquery := dialect.From(goqu.T(TABLE_PROJECTS).As("p2")).
        Select(goqu.I("p2." + COLUMN_ID)).
        Join(
            goqu.T(TABLE_POLICIES).As("pol2"),
            goqu.On(goqu.I("p2."+COLUMN_ID).Eq(goqu.I("pol2."+COLUMN_RESOURCE_ID))),
        ).
        Where(goqu.And(
            goqu.I("p2."+COLUMN_ORG_ID).Eq(orgID),
            goqu.I("pol2."+COLUMN_PRINCIPAL_ID).Eq(userID),
            goqu.I("pol2."+COLUMN_RESOURCE_TYPE).Eq("app/project"),
            goqu.I("pol2."+COLUMN_PRINCIPAL_TYPE).Eq("app/user"),
            goqu.I("pol2."+COLUMN_DELETED_AT).IsNull(),
        ))

    mainQuery := dialect.From(goqu.T(TABLE_PROJECTS).As("p")).Prepared(false).
        Select(
            goqu.I("p."+COLUMN_ID).As("project_id"),
            goqu.I("p."+COLUMN_NAME).As("project_title"),
            goqu.I("p."+COLUMN_CREATED_AT).As("project_created_on"),
            goqu.L("array_agg(DISTINCT u."+COLUMN_ID+" ORDER BY u."+COLUMN_ID+")").As("user_ids"),
            goqu.L("array_agg(DISTINCT u."+COLUMN_NAME+" ORDER BY u."+COLUMN_NAME+")").As("user_names"),
            goqu.L("array_agg(DISTINCT u.title ORDER BY u.title)").As("user_titles"),
        ).
        Join(
            goqu.T(TABLE_POLICIES).As("pol"),
            goqu.On(goqu.And(
                goqu.I("p."+COLUMN_ID).Eq(goqu.I("pol."+COLUMN_RESOURCE_ID)),
                goqu.I("pol."+COLUMN_RESOURCE_TYPE).Eq("app/project"),
                goqu.I("pol."+COLUMN_DELETED_AT).IsNull(),
            )),
        ).
        Join(
            goqu.T(TABLE_USERS).As("u"),
            goqu.On(goqu.And(
                goqu.I("pol."+COLUMN_PRINCIPAL_ID).Eq(goqu.I("u."+COLUMN_ID)),
                goqu.I("pol."+COLUMN_PRINCIPAL_TYPE).Eq("app/user"),
            )),
        ).
        Where(goqu.I("p."+COLUMN_ID).In(subquery)).
        GroupBy(
            goqu.I("p."+COLUMN_ID),
            goqu.I("p."+COLUMN_NAME),
            goqu.I("p."+COLUMN_CREATED_AT),
        ).
        Order(goqu.I("p."+COLUMN_NAME).Asc())

    return mainQuery
}

func (r UserProjectsRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchPattern := "%" + search + "%"

	baseSubquery := query.As(TABLE_BASE)

	return dialect.From(baseSubquery).Prepared(false).
		Where(
			goqu.Or(
				// Project field searches - only title and name as requested
				goqu.I(TABLE_BASE+"."+COLUMN_PROJECT_TITLE).ILike(searchPattern),
				goqu.I(TABLE_BASE+"."+COLUMN_PROJECT_NAME).ILike(searchPattern),
			),
		)
}
