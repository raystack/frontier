package postgres

import (
	"context"
	"database/sql"
	"fmt"

	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/projectusers"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

const (
	COLUMN_PROJECT_ID     = "project_id"
	COLUMN_PROJECT_JOINED = "project_joined_at"
	RESOURCE_TYPE_PROJECT = "app/project"
	TABLE_BASE            = "base"
)

type ProjectUsersRepository struct {
	dbc *db.Client
}

type ProjectUsers struct {
	UserID          sql.NullString `db:"id"`
	UserName        sql.NullString `db:"name"`
	UserEmail       sql.NullString `db:"email"`
	UserTitle       sql.NullString `db:"title"`
	UserState       sql.NullString `db:"state"`
	RoleNames       sql.NullString `db:"role_names"`
	RoleTitles      sql.NullString `db:"role_titles"`
	RoleIDs         sql.NullString `db:"role_ids"`
	ProjectID       sql.NullString `db:"project_id"`
	ProjectJoinedAt sql.NullTime   `db:"project_joined_at"`
}

func (u *ProjectUsers) transformToAggregatedUser() svc.AggregatedUser {
	return svc.AggregatedUser{
		ID:              u.UserID.String,
		Name:            u.UserName.String,
		Email:           u.UserEmail.String,
		Title:           u.UserTitle.String,
		State:           user.State(u.UserState.String),
		RoleNames:       strings.Split(u.RoleNames.String, ","),
		RoleTitles:      strings.Split(u.RoleTitles.String, ","),
		RoleIDs:         strings.Split(u.RoleIDs.String, ","),
		ProjectID:       u.ProjectID.String,
		ProjectJoinedAt: u.ProjectJoinedAt.Time,
	}
}

func NewProjectUsersRepository(dbc *db.Client) *ProjectUsersRepository {
	return &ProjectUsersRepository{
		dbc: dbc,
	}
}

func (r ProjectUsersRepository) Search(ctx context.Context, projectID string, rql *rql.Query) (svc.ProjectUsers, error) {
	dataQuery, params, err := r.prepareDataQuery(projectID, rql)
	fmt.Println(dataQuery)
	if err != nil {
		return svc.ProjectUsers{}, err
	}

	var projectUsers []ProjectUsers

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		err = r.dbc.WithTimeout(ctx, TABLE_POLICIES, "GetProjectUsers", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &projectUsers, dataQuery, params...)
		})
		return err
	})

	if err != nil {
		return svc.ProjectUsers{}, err
	}

	res := make([]svc.AggregatedUser, 0)
	for _, user := range projectUsers {
		res = append(res, user.transformToAggregatedUser())
	}
	return svc.ProjectUsers{
		Users: res,
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r ProjectUsersRepository) prepareDataQuery(projectID string, rql *rql.Query) (string, []interface{}, error) {
	query := r.buildBaseQuery(projectID)

	if rql.Search != "" {
		query = r.addSearch(query, rql.Search)
	}

	return query.Offset(uint(rql.Offset)).Limit(uint(rql.Limit)).ToSQL()
}

func (r ProjectUsersRepository) buildBaseQuery(projectID string) *goqu.SelectDataset {
	return dialect.From(TABLE_POLICIES).Prepared(true).
		Select(
			goqu.I(TABLE_USERS+"."+COLUMN_ID),
			goqu.I(TABLE_USERS+"."+COLUMN_NAME),
			goqu.I(TABLE_USERS+"."+COLUMN_EMAIL),
			goqu.I(TABLE_USERS+"."+COLUMN_TITLE),
			goqu.I(TABLE_USERS+"."+COLUMN_STATE),
			goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).As(COLUMN_PROJECT_ID),
			goqu.MIN(goqu.I(TABLE_POLICIES+"."+COLUMN_CREATED_AT)).As(COLUMN_PROJECT_JOINED),
			goqu.L("string_agg(DISTINCT "+TABLE_ROLES+"."+COLUMN_NAME+", ',')").As(COLUMN_ROLE_NAMES),
			goqu.L("string_agg(DISTINCT "+TABLE_ROLES+"."+COLUMN_TITLE+", ',')").As(COLUMN_ROLE_TITLES),
			goqu.L("string_agg(DISTINCT "+TABLE_ROLES+"."+COLUMN_ID+"::text, ',')").As(COLUMN_ROLE_IDS),
		).
		InnerJoin(
			goqu.T(TABLE_USERS),
			goqu.On(goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID).Eq(goqu.I(TABLE_USERS+"."+COLUMN_ID))),
		).
		InnerJoin(
			goqu.T(TABLE_ROLES),
			goqu.On(goqu.I(TABLE_POLICIES+"."+COLUMN_ROLE_ID).Eq(goqu.I(TABLE_ROLES+"."+COLUMN_ID))),
		).
		Where(goqu.Ex{
			TABLE_POLICIES + "." + COLUMN_RESOURCE_ID:    projectID,
			TABLE_POLICIES + "." + COLUMN_RESOURCE_TYPE:  RESOURCE_TYPE_PROJECT,
			TABLE_POLICIES + "." + COLUMN_PRINCIPAL_TYPE: PRINCIPAL_TYPE_USER,
		}).
		GroupBy(
			TABLE_USERS+"."+COLUMN_ID,
			TABLE_USERS+"."+COLUMN_NAME,
			TABLE_USERS+"."+COLUMN_EMAIL,
			TABLE_USERS+"."+COLUMN_TITLE,
			TABLE_USERS+"."+COLUMN_STATE,
			TABLE_POLICIES+"."+COLUMN_RESOURCE_ID,
		)
}

func (r ProjectUsersRepository) addSearch(query *goqu.SelectDataset, search string) *goqu.SelectDataset {
	searchPattern := "%" + search + "%"

	baseSubquery := query.As(TABLE_BASE)

	return dialect.From(baseSubquery).Prepared(true).
		Where(
			goqu.Or(
				// User field searches
				goqu.I(TABLE_BASE+"."+COLUMN_NAME).ILike(searchPattern),
				goqu.I(TABLE_BASE+"."+COLUMN_EMAIL).ILike(searchPattern),
				goqu.I(TABLE_BASE+"."+COLUMN_TITLE).ILike(searchPattern),
				goqu.I(TABLE_BASE+"."+COLUMN_STATE).ILike(searchPattern),
				// Search on already aggregated role columns
				goqu.I(TABLE_BASE+"."+COLUMN_ROLE_NAMES).ILike(searchPattern),
				goqu.I(TABLE_BASE+"."+COLUMN_ROLE_TITLES).ILike(searchPattern),
				goqu.I(TABLE_BASE+"."+COLUMN_ROLE_IDS).ILike(searchPattern),
			),
		)
}
