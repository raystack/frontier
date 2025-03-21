package postgres

import (
	"context"
	"database/sql"

	// "errors"
	// "fmt"
	// "slices"
	// "strings"

	// "github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	svc "github.com/raystack/frontier/core/aggregates/orgprojects"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
)

type OrgProjectsRepository struct {
	dbc *db.Client
}

type OrgProjects struct {
	ProjectID      sql.NullString `db:"id"`
	ProjectName    sql.NullString `db:"name"`
	ProjectTitle   sql.NullString `db:"title"`
	ProjectState   sql.NullString `db:"state"`
	MemberCount    sql.NullInt64  `db:"member_count"`
	ProjectAvatars pq.StringArray `db:"avatars"`
	CreatedAt      sql.NullTime   `db:"created_at"`
	OrganizationID sql.NullString `db:"organization_id"`
}

type OrgProjectsGroup struct {
	Name sql.NullString         `db:"name"`
	Data []OrgProjectsGroupData `db:"data"`
}

type OrgProjectsGroupData struct {
	Name  sql.NullString `db:"values"`
	Count int            `db:"count"`
}

func (p *OrgProjects) transformToAggregatedProject() svc.AggregatedProject {
	return svc.AggregatedProject{
		ID:             p.ProjectID.String,
		Name:           p.ProjectName.String,
		Title:          p.ProjectTitle.String,
		State:          project.State(p.ProjectState.String),
		MemberCount:    p.MemberCount.Int64,
		Avatars:        p.ProjectAvatars,
		CreatedAt:      p.CreatedAt.Time,
		OrganizationID: p.OrganizationID.String,
	}
}

func NewOrgProjectsRepository(dbc *db.Client) *OrgProjectsRepository {
	return &OrgProjectsRepository{
		dbc: dbc,
	}
}

func (r OrgProjectsRepository) Search(ctx context.Context, orgID string, rql *rql.Query) (svc.OrgProjects, error) {
	return svc.OrgProjects{}, nil
	dataQuery, params, err := r.prepareDataQuery(orgID, rql)
	if err != nil {
		return svc.OrgProjects{}, err
	}

	var orgProjects []OrgProjects

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "GetOrgProjects", func(ctx context.Context) error {
			return tx.SelectContext(ctx, &orgProjects, dataQuery, params...)
		})

		if err != nil {
			return err
		}

		return err
	})

	if err != nil {
		return svc.OrgProjects{}, err
	}

	res := make([]svc.AggregatedProject, 0)
	for _, project := range orgProjects {
		res = append(res, project.transformToAggregatedProject())
	}
	return svc.OrgProjects{
		Projects: res,
		Pagination: svc.Page{
			Offset: rql.Offset,
			Limit:  rql.Limit,
		},
	}, nil
}

func (r OrgProjectsRepository) prepareDataQuery(orgID string, rql *rql.Query) (string, []interface{}, error) {
	return "", nil, nil
}
