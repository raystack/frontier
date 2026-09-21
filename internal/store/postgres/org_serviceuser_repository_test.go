package postgres

import (
	"testing"

	svc "github.com/raystack/frontier/core/aggregates/orgserviceuser"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

const orgServiceUserBaseSQL = `SELECT * FROM (` +
	`SELECT "serviceusers"."id" AS "id", "serviceusers"."title" AS "title", "serviceusers"."org_id" AS "org_id", "serviceusers"."created_at" AS "created_at", ` +
	`COALESCE(JSONB_AGG(DISTINCT JSONB_BUILD_OBJECT('id', projects.id, 'title', projects.title, 'name', projects.name)) FILTER (WHERE projects.id IS NOT NULL), '[]') AS "project_data", ` +
	`COALESCE(STRING_AGG(DISTINCT projects.title, ', ') FILTER (WHERE projects.id IS NOT NULL), '') AS "projects" ` +
	`FROM "serviceusers" ` +
	`LEFT JOIN "policies" ON (("serviceusers"."id" = "policies"."principal_id") AND ("policies"."principal_type" = $1) AND ("policies"."resource_type" = $2) AND ("policies"."deleted_at" IS NULL)) ` +
	`LEFT JOIN "projects" ON (("policies"."resource_id" = "projects"."id") AND ("projects"."deleted_at" IS NULL)) ` +
	`WHERE (("serviceusers"."org_id" = $3) AND ("serviceusers"."deleted_at" IS NULL)) ` +
	`GROUP BY "serviceusers"."id"` +
	`) AS "org_service_users"`

func baseParams(rest ...any) []any {
	return append([]any{"app/serviceuser", "app/project", "org1"}, rest...)
}

func TestOrgServiceUserRepository_prepareDataQuery(t *testing.T) {
	tests := []struct {
		name       string
		rql        *rql.Query
		wantSQL    string
		wantParams []any
		wantPage   utils.Page
		wantErr    string
	}{
		{
			name:       "no options falls back to title asc and the default limit",
			rql:        &rql.Query{},
			wantSQL:    orgServiceUserBaseSQL + ` ORDER BY "title" ASC, "id" ASC LIMIT $4`,
			wantParams: baseParams(int64(50)),
			wantPage:   utils.Page{Limit: 50, Offset: 0},
		},
		{
			name:       "nil query behaves like an empty one",
			rql:        nil,
			wantSQL:    orgServiceUserBaseSQL + ` ORDER BY "title" ASC, "id" ASC LIMIT $4`,
			wantParams: baseParams(int64(50)),
			wantPage:   utils.Page{Limit: 50, Offset: 0},
		},
		{
			name:       "requested sort leads the order by",
			rql:        &rql.Query{Limit: 50, Sort: []rql.Sort{{Name: "created_at", Order: "desc"}}},
			wantSQL:    orgServiceUserBaseSQL + ` ORDER BY "created_at" DESC, "id" ASC LIMIT $4`,
			wantParams: baseParams(int64(50)),
			wantPage:   utils.Page{Limit: 50, Offset: 0},
		},
		{
			name:       "sorting by title descending really is descending",
			rql:        &rql.Query{Limit: 10, Sort: []rql.Sort{{Name: "title", Order: "desc"}}},
			wantSQL:    orgServiceUserBaseSQL + ` ORDER BY "title" DESC, "id" ASC LIMIT $4`,
			wantParams: baseParams(int64(10)),
			wantPage:   utils.Page{Limit: 10, Offset: 0},
		},
		{
			name:       "several sort terms keep their order and still end on id",
			rql:        &rql.Query{Limit: 10, Sort: []rql.Sort{{Name: "title", Order: "asc"}, {Name: "created_at", Order: "desc"}}},
			wantSQL:    orgServiceUserBaseSQL + ` ORDER BY "title" ASC, "created_at" DESC, "id" ASC LIMIT $4`,
			wantParams: baseParams(int64(10)),
			wantPage:   utils.Page{Limit: 10, Offset: 0},
		},
		{
			name:       "search matches on title or project titles",
			rql:        &rql.Query{Search: "test", Limit: 10, Offset: 5},
			wantSQL:    orgServiceUserBaseSQL + ` WHERE ((CAST("title" AS TEXT) ILIKE $4) OR (CAST("projects" AS TEXT) ILIKE $5)) ORDER BY "title" ASC, "id" ASC LIMIT $6 OFFSET $7`,
			wantParams: baseParams("%test%", "%test%", int64(10), int64(5)),
			wantPage:   utils.Page{Limit: 10, Offset: 5},
		},
		{
			name:       "equality filter on title",
			rql:        &rql.Query{Limit: 10, Filters: []rql.Filter{{Name: "title", Operator: "eq", Value: "svc"}}},
			wantSQL:    orgServiceUserBaseSQL + ` WHERE ("title" = $4) ORDER BY "title" ASC, "id" ASC LIMIT $5`,
			wantParams: baseParams("svc", int64(10)),
			wantPage:   utils.Page{Limit: 10, Offset: 0},
		},
		{
			name:       "empty filter on title",
			rql:        &rql.Query{Limit: 10, Filters: []rql.Filter{{Name: "title", Operator: "empty", Value: ""}}},
			wantSQL:    orgServiceUserBaseSQL + ` WHERE (("title" IS NULL) OR ("title" = $4)) ORDER BY "title" ASC, "id" ASC LIMIT $5`,
			wantParams: baseParams("", int64(10)),
			wantPage:   utils.Page{Limit: 10, Offset: 0},
		},
		{
			name:       "datetime filter on created_at",
			rql:        &rql.Query{Limit: 10, Filters: []rql.Filter{{Name: "created_at", Operator: "gt", Value: "2026-01-01"}}},
			wantSQL:    orgServiceUserBaseSQL + ` WHERE ("created_at" > $4) ORDER BY "title" ASC, "id" ASC LIMIT $5`,
			wantParams: baseParams("2026-01-01", int64(10)),
			wantPage:   utils.Page{Limit: 10, Offset: 0},
		},
		{
			name:       "filter on projects, which the admin ui offers",
			rql:        &rql.Query{Limit: 10, Filters: []rql.Filter{{Name: "projects", Operator: "ilike", Value: "%Live%"}}},
			wantSQL:    orgServiceUserBaseSQL + ` WHERE ("projects" ILIKE $4) ORDER BY "title" ASC, "id" ASC LIMIT $5`,
			wantParams: baseParams("%Live%", int64(10)),
			wantPage:   utils.Page{Limit: 10, Offset: 0},
		},
		{
			name:    "group_by is rejected rather than producing broken sql",
			rql:     &rql.Query{Limit: 10, GroupBy: []string{"title"}},
			wantErr: "bad input: group_by is not supported",
		},
		{
			name:    "filtering on a column the query does not select is rejected",
			rql:     &rql.Query{Limit: 10, Filters: []rql.Filter{{Name: "state", Operator: "eq", Value: "enabled"}}},
			wantErr: "bad input: state is not supported in filters",
		},
		{
			name:    "sorting on a column the query does not select is rejected",
			rql:     &rql.Query{Limit: 10, Sort: []rql.Sort{{Name: "state", Order: "asc"}}},
			wantErr: "bad input: state is not supported in sort",
		},
		{
			name:       "a limit of zero falls back to the default rather than returning everything",
			rql:        &rql.Query{Limit: 0, Offset: 20},
			wantSQL:    orgServiceUserBaseSQL + ` ORDER BY "title" ASC, "id" ASC LIMIT $4 OFFSET $5`,
			wantParams: baseParams(int64(50), int64(20)),
			wantPage:   utils.Page{Limit: 50, Offset: 20},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &OrgServiceUserRepository{}
			gotSQL, gotParams, gotPage, err := r.prepareDataQuery("org1", tt.rql)

			if tt.wantErr != "" {
				assert.EqualError(t, err, tt.wantErr)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.wantSQL, gotSQL)
			assert.Equal(t, tt.wantParams, gotParams)
			assert.Equal(t, tt.wantPage, gotPage)
		})
	}
}

func TestServiceUserRow_transformToAggregatedServiceUser(t *testing.T) {
	tests := []struct {
		name         string
		row          ServiceUserRow
		wantProjects []svc.Project
	}{
		{
			name: "service user with no project policy",
			row: ServiceUserRow{
				ID:          "su1",
				Title:       "no project service user",
				ProjectData: "[]",
			},
			wantProjects: []svc.Project{},
		},
		{
			name: "service user with projects",
			row: ServiceUserRow{
				ID:          "su2",
				Title:       "service user with projects",
				ProjectData: `[{"id":"p1","title":"Project One","name":"project-one"}]`,
			},
			wantProjects: []svc.Project{
				{ID: "p1", Title: "Project One", Name: "project-one"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.row.transformToAggregatedServiceUser("org1")

			assert.Equal(t, tt.row.ID, got.ID)
			assert.Equal(t, tt.row.Title, got.Title)
			assert.Equal(t, "org1", got.OrgID)
			assert.Equal(t, tt.wantProjects, got.Projects)
		})
	}
}
