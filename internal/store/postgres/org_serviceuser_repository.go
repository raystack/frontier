package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	svc "github.com/raystack/frontier/core/aggregates/orgserviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
)

type ServiceUserRow struct {
	ID          string `db:"id"`
	Title       string `db:"title"`
	OrgID       string `db:"org_id"`
	ProjectData string `db:"project_data"`
	// only here so the aggregated project titles, which rql filters and searches on,
	// have somewhere to land when the wrapper select returns them
	Projects  string       `db:"projects"`
	CreatedAt sql.NullTime `db:"created_at"`
}

func (c *ServiceUserRow) transformToAggregatedServiceUser(orgID string) svc.AggregatedServiceUser {
	var projects []svc.Project
	if c.ProjectData != "" && c.ProjectData != "null" {
		// Parse JSON array of project objects
		err := json.Unmarshal([]byte(c.ProjectData), &projects)
		if err != nil {
			// If JSON parsing fails, return empty projects array
			projects = []svc.Project{}
		}
	}

	return svc.AggregatedServiceUser{
		ID:        c.ID,
		OrgID:     orgID,
		Title:     c.Title,
		CreatedAt: c.CreatedAt.Time,
		Projects:  projects,
	}
}

// the base query joins serviceusers, policies and projects, and all three carry an
// id, a title and a created_at. rql is therefore applied to a wrapper select over the
// base query, where only these aliased output columns are visible and unambiguous.
const (
	ORG_SERVICE_USER_BASE_ALIAS = "org_service_users"
	COLUMN_PROJECTS             = "projects"
)

var (
	orgServiceUserRQLFilterSupportedColumns = []string{COLUMN_ID, COLUMN_TITLE, COLUMN_ORG_ID, COLUMN_CREATED_AT, COLUMN_PROJECTS}
	orgServiceUserRQLSearchSupportedColumns = []string{COLUMN_TITLE, COLUMN_PROJECTS}
	orgServiceUserRQLSortSupportedColumns   = []string{COLUMN_ID, COLUMN_TITLE, COLUMN_CREATED_AT}
)

type OrgServiceUserRepository struct {
	dbc *db.Client
}

func NewOrgServiceUserRepository(dbc *db.Client) *OrgServiceUserRepository {
	return &OrgServiceUserRepository{
		dbc: dbc,
	}
}

func (r OrgServiceUserRepository) Search(ctx context.Context, orgID string, rqlQuery *rql.Query) (svc.OrganizationServiceUsers, error) {
	dataQuery, params, page, err := r.prepareDataQuery(orgID, rqlQuery)
	if err != nil {
		return svc.OrganizationServiceUsers{}, err
	}

	var serviceUsers []ServiceUserRow

	txOpts := sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  true,
	}

	err = r.dbc.WithTxn(ctx, txOpts, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_SERVICE_USERS, "SearchOrgServiceUsers", func(ctx context.Context) error {
			return r.dbc.SelectContext(ctx, &serviceUsers, dataQuery, params...)
		})
	})

	if err != nil {
		return svc.OrganizationServiceUsers{}, err
	}

	res := make([]svc.AggregatedServiceUser, 0)
	for _, serviceuser := range serviceUsers {
		transformed := serviceuser.transformToAggregatedServiceUser(orgID)
		res = append(res, transformed)
	}

	return svc.OrganizationServiceUsers{
		ServiceUsers: res,
		Pagination: svc.Page{
			Offset: page.Offset,
			Limit:  page.Limit,
		},
	}, nil
}

func (r OrgServiceUserRepository) prepareDataQuery(orgID string, rqlQuery *rql.Query) (string, []any, utils.Page, error) {
	query := dialect.From(r.buildBaseQuery(orgID).As(ORG_SERVICE_USER_BASE_ALIAS)).Prepared(true)

	if rqlQuery == nil {
		rqlQuery = &rql.Query{}
	}

	// the response has no group block, and the shared sort helper would otherwise
	// order by a group_by column that the wrapper select does not have
	if len(rqlQuery.GroupBy) > 0 {
		return "", nil, utils.Page{}, fmt.Errorf("%w: group_by is not supported", ErrBadInput)
	}

	query, err := utils.AddRQLFiltersInQuery(query, rqlQuery, orgServiceUserRQLFilterSupportedColumns, svc.AggregatedServiceUser{})
	if err != nil {
		return "", nil, utils.Page{}, fmt.Errorf("%w: %w", ErrBadInput, err)
	}

	query, err = utils.AddRQLSearchInQuery(query, rqlQuery, orgServiceUserRQLSearchSupportedColumns)
	if err != nil {
		return "", nil, utils.Page{}, fmt.Errorf("%w: %w", ErrBadInput, err)
	}

	for _, sortItem := range rqlQuery.Sort {
		if !slices.Contains(orgServiceUserRQLSortSupportedColumns, sortItem.Name) {
			return "", nil, utils.Page{}, fmt.Errorf("%w: %s is not supported in sort", ErrBadInput, sortItem.Name)
		}
	}
	query, err = utils.AddRQLSortInQuery(query, rqlQuery)
	if err != nil {
		return "", nil, utils.Page{}, fmt.Errorf("%w: %w", ErrBadInput, err)
	}

	// the request decides the order, so title asc is only used when nothing was asked for
	if len(rqlQuery.Sort) == 0 {
		query = query.OrderAppend(goqu.C(COLUMN_TITLE).Asc())
	}
	// the last term makes the order total, so paging by offset cannot repeat or skip
	// a row when two rows tie on the sorted column
	query = query.OrderAppend(goqu.C(COLUMN_ID).Asc())

	query, page := utils.AddRQLPaginationInQuery(query, rqlQuery)

	sql, params, err := query.ToSQL()
	return sql, params, page, err
}

func (r OrgServiceUserRepository) buildBaseQuery(orgID string) *goqu.SelectDataset {
	return dialect.From(TABLE_SERVICE_USERS).Prepared(true).
		Select(
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_ID).As("id"),
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_TITLE).As("title"),
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_ORG_ID).As("org_id"),
			goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_CREATED_AT).As("created_at"),
			// a service user without any project policy still has to show up, so the
			// aggregate skips the null rows the left join produces and falls back to [].
			// two roles on one project produce two joined rows, so it is deduped, which
			// needs jsonb because the json type has no equality operator.
			goqu.L("COALESCE(JSONB_AGG(DISTINCT JSONB_BUILD_OBJECT('id', "+TABLE_PROJECTS+"."+COLUMN_ID+", 'title', "+TABLE_PROJECTS+"."+COLUMN_TITLE+", 'name', "+TABLE_PROJECTS+"."+COLUMN_NAME+")) FILTER (WHERE "+TABLE_PROJECTS+"."+COLUMN_ID+" IS NOT NULL), '[]')").As("project_data"),
			// the admin ui renders the project titles joined by a comma and lets you
			// filter on that column, so the same text is exposed for rql to filter on
			goqu.L("COALESCE(STRING_AGG(DISTINCT "+TABLE_PROJECTS+"."+COLUMN_TITLE+", ', ') FILTER (WHERE "+TABLE_PROJECTS+"."+COLUMN_ID+" IS NOT NULL), '')").As("projects"),
		).
		LeftJoin(
			goqu.T(TABLE_POLICIES),
			goqu.On(
				goqu.I(TABLE_SERVICE_USERS+"."+COLUMN_ID).Eq(goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_ID)),
				goqu.I(TABLE_POLICIES+"."+COLUMN_PRINCIPAL_TYPE).Eq(schema.ServiceUserPrincipal),
				goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_TYPE).Eq(schema.ProjectNamespace),
			),
		).
		LeftJoin(
			goqu.T(TABLE_PROJECTS),
			goqu.On(
				goqu.I(TABLE_POLICIES+"."+COLUMN_RESOURCE_ID).Eq(goqu.I(TABLE_PROJECTS+"."+COLUMN_ID)),
				goqu.I(TABLE_PROJECTS+"."+COLUMN_DELETED_AT).IsNull(),
			),
		).
		Where(goqu.Ex{
			TABLE_SERVICE_USERS + "." + COLUMN_ORG_ID: orgID,
		}).
		GroupBy(
			TABLE_SERVICE_USERS + "." + COLUMN_ID,
		)
}
