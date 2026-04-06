package postgres

import (
	"testing"
	"time"

	svc "github.com/raystack/frontier/core/aggregates/orgpats"
	patmodels "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

func TestOrgPATsRepository_buildCountQuery(t *testing.T) {
	r := OrgPATsRepository{}

	tests := []struct {
		name    string
		orgID   string
		rql     *rql.Query
		wantErr bool
	}{
		{
			name:  "basic count",
			orgID: "org-1",
			rql:   &rql.Query{Limit: 30, Offset: 0},
		},
		{
			name:  "count with search",
			orgID: "org-1",
			rql:   &rql.Query{Limit: 30, Offset: 0, Search: "aman"},
		},
		{
			name:  "count with filter",
			orgID: "org-1",
			rql: &rql.Query{
				Limit:  30,
				Offset: 0,
				Filters: []rql.Filter{
					{Name: "title", Operator: opILike, Value: "%test%"},
				},
			},
		},
		{
			name:  "count with unsupported filter",
			orgID: "org-1",
			rql: &rql.Query{
				Limit:  30,
				Offset: 0,
				Filters: []rql.Filter{
					{Name: "invalid_field", Operator: "eq", Value: "x"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params, err := r.buildCountQuery(tt.orgID, tt.rql)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, sql)
			assert.Contains(t, sql, "COUNT(*)")
			assert.NotContains(t, sql, "ORDER BY")
			assert.NotEmpty(t, params)
		})
	}
}

func TestOrgPATsRepository_buildDataQuery(t *testing.T) {
	r := OrgPATsRepository{}

	tests := []struct {
		name    string
		orgID   string
		rql     *rql.Query
		wantErr bool
	}{
		{
			name:  "basic data query with pagination",
			orgID: "org-1",
			rql:   &rql.Query{Limit: 30, Offset: 0},
		},
		{
			name:  "data query with search",
			orgID: "org-1",
			rql:   &rql.Query{Limit: 10, Offset: 5, Search: "test"},
		},
		{
			name:  "data query with ilike filter",
			orgID: "org-1",
			rql: &rql.Query{
				Limit:  30,
				Offset: 0,
				Filters: []rql.Filter{
					{Name: "created_by_email", Operator: opILike, Value: "%@pixxel%"},
				},
			},
		},
		{
			name:  "data query with sort",
			orgID: "org-1",
			rql: &rql.Query{
				Limit:  30,
				Offset: 0,
				Sort:   []rql.Sort{{Name: "title", Order: "asc"}},
			},
		},
		{
			name:  "data query with unsupported sort",
			orgID: "org-1",
			rql: &rql.Query{
				Limit:  30,
				Offset: 0,
				Sort:   []rql.Sort{{Name: "role_name", Order: "asc"}},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, params, err := r.buildDataQuery(tt.orgID, tt.rql)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, sql)
			assert.Contains(t, sql, "pat_id")
			assert.Contains(t, sql, "created_by_title")
			assert.Contains(t, sql, "pol")
			assert.NotEmpty(t, params)
		})
	}
}

func TestOrgPATsRepository_addFilter(t *testing.T) {
	r := OrgPATsRepository{}
	baseQuery := dialect.From("test")

	tests := []struct {
		name    string
		filter  rql.Filter
		wantErr bool
	}{
		{name: "empty operator", filter: rql.Filter{Name: "title", Operator: opEmpty}},
		{name: "notempty operator", filter: rql.Filter{Name: "title", Operator: opNotEmpty}},
		{name: "like operator", filter: rql.Filter{Name: "title", Operator: opLike, Value: "test%"}},
		{name: "notlike operator", filter: rql.Filter{Name: "title", Operator: opNotLike, Value: "%test"}},
		{name: "ilike operator", filter: rql.Filter{Name: "created_by_email", Operator: opILike, Value: "%aman%"}},
		{name: "notilike operator", filter: rql.Filter{Name: "created_by_title", Operator: opNotILike, Value: "%admin%"}},
		{name: "eq operator", filter: rql.Filter{Name: "id", Operator: "eq", Value: "uuid-1"}},
		{name: "unsupported field", filter: rql.Filter{Name: "bad_field", Operator: "eq", Value: "x"}, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := r.addFilter(baseQuery, tt.filter)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOrgPATsRepository_groupRows(t *testing.T) {
	r := OrgPATsRepository{}

	strPtr := func(s string) *string { return &s }
	now := time.Now()

	tests := []struct {
		name     string
		rows     []OrgPATRow
		wantPATs int
	}{
		{
			name:     "empty rows",
			rows:     []OrgPATRow{},
			wantPATs: 0,
		},
		{
			name: "single PAT with no policies",
			rows: []OrgPATRow{
				{
					PATID: "pat-1", PATTitle: "my-token",
					CreatedByID: "user-1", CreatedByTitle: "John", CreatedByEmail: "john@test.com",
					CreatedAt: now, ExpiresAt: now,
				},
			},
			wantPATs: 1,
		},
		{
			name: "single PAT with org scope",
			rows: []OrgPATRow{
				{
					PATID: "pat-1", PATTitle: "my-token",
					CreatedByID: "user-1", CreatedByTitle: "John", CreatedByEmail: "john@test.com",
					CreatedAt: now, ExpiresAt: now,
					RoleID: strPtr("role-1"), ResourceType: strPtr(schema.OrganizationNamespace),
					ResourceID: strPtr("org-1"),
				},
			},
			wantPATs: 1,
		},
		{
			name: "single PAT with multiple project scopes",
			rows: []OrgPATRow{
				{
					PATID: "pat-1", PATTitle: "my-token",
					CreatedByID: "user-1", CreatedByTitle: "John", CreatedByEmail: "john@test.com",
					CreatedAt: now, ExpiresAt: now,
					RoleID: strPtr("role-1"), ResourceType: strPtr(schema.ProjectNamespace),
					ResourceID: strPtr("proj-1"),
				},
				{
					PATID: "pat-1", PATTitle: "my-token",
					CreatedByID: "user-1", CreatedByTitle: "John", CreatedByEmail: "john@test.com",
					CreatedAt: now, ExpiresAt: now,
					RoleID: strPtr("role-1"), ResourceType: strPtr(schema.ProjectNamespace),
					ResourceID: strPtr("proj-2"),
				},
			},
			wantPATs: 1,
		},
		{
			name: "single PAT with all-projects scope",
			rows: []OrgPATRow{
				{
					PATID: "pat-1", PATTitle: "my-token",
					CreatedByID: "user-1", CreatedByTitle: "John", CreatedByEmail: "john@test.com",
					CreatedAt: now, ExpiresAt: now,
					RoleID: strPtr("role-1"), ResourceType: strPtr(schema.OrganizationNamespace),
					ResourceID: strPtr("org-1"), GrantRelation: strPtr(schema.PATGrantRelationName),
				},
			},
			wantPATs: 1,
		},
		{
			name: "multiple PATs",
			rows: []OrgPATRow{
				{
					PATID: "pat-1", PATTitle: "token-1",
					CreatedByID: "user-1", CreatedByTitle: "John", CreatedByEmail: "john@test.com",
					CreatedAt: now, ExpiresAt: now,
					RoleID: strPtr("role-1"), ResourceType: strPtr(schema.OrganizationNamespace),
					ResourceID: strPtr("org-1"),
				},
				{
					PATID: "pat-2", PATTitle: "token-2",
					CreatedByID: "user-2", CreatedByTitle: "Jane", CreatedByEmail: "jane@test.com",
					CreatedAt: now, ExpiresAt: now,
					RoleID: strPtr("role-2"), ResourceType: strPtr(schema.ProjectNamespace),
					ResourceID: strPtr("proj-1"),
				},
			},
			wantPATs: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := r.groupRows(tt.rows)
			assert.Len(t, result, tt.wantPATs)
		})
	}

	// Detailed assertions for specific cases
	t.Run("project scope collects resource IDs", func(t *testing.T) {
		rows := []OrgPATRow{
			{
				PATID: "pat-1", PATTitle: "token", CreatedByID: "u1", CreatedByTitle: "U", CreatedByEmail: "u@t.c",
				CreatedAt: now, ExpiresAt: now,
				RoleID: strPtr("role-1"), ResourceType: strPtr(schema.ProjectNamespace), ResourceID: strPtr("proj-1"),
			},
			{
				PATID: "pat-1", PATTitle: "token", CreatedByID: "u1", CreatedByTitle: "U", CreatedByEmail: "u@t.c",
				CreatedAt: now, ExpiresAt: now,
				RoleID: strPtr("role-1"), ResourceType: strPtr(schema.ProjectNamespace), ResourceID: strPtr("proj-2"),
			},
		}
		result := r.groupRows(rows)
		assert.Len(t, result, 1)
		assert.Len(t, result[0].Scopes, 1)
		assert.Equal(t, schema.ProjectNamespace, result[0].Scopes[0].ResourceType)
		assert.ElementsMatch(t, []string{"proj-1", "proj-2"}, result[0].Scopes[0].ResourceIDs)
	})

	t.Run("all-projects scope has nil resource IDs", func(t *testing.T) {
		rows := []OrgPATRow{
			{
				PATID: "pat-1", PATTitle: "token", CreatedByID: "u1", CreatedByTitle: "U", CreatedByEmail: "u@t.c",
				CreatedAt: now, ExpiresAt: now,
				RoleID: strPtr("role-1"), ResourceType: strPtr(schema.OrganizationNamespace),
				ResourceID: strPtr("org-1"), GrantRelation: strPtr(schema.PATGrantRelationName),
			},
		}
		result := r.groupRows(rows)
		assert.Len(t, result, 1)
		assert.Len(t, result[0].Scopes, 1)
		assert.Equal(t, schema.ProjectNamespace, result[0].Scopes[0].ResourceType)
		assert.Nil(t, result[0].Scopes[0].ResourceIDs)
	})

	t.Run("mixed scopes for same PAT", func(t *testing.T) {
		rows := []OrgPATRow{
			{
				PATID: "pat-1", PATTitle: "token", CreatedByID: "u1", CreatedByTitle: "U", CreatedByEmail: "u@t.c",
				CreatedAt: now, ExpiresAt: now,
				RoleID: strPtr("role-org"), ResourceType: strPtr(schema.OrganizationNamespace), ResourceID: strPtr("org-1"),
			},
			{
				PATID: "pat-1", PATTitle: "token", CreatedByID: "u1", CreatedByTitle: "U", CreatedByEmail: "u@t.c",
				CreatedAt: now, ExpiresAt: now,
				RoleID: strPtr("role-proj"), ResourceType: strPtr(schema.ProjectNamespace), ResourceID: strPtr("proj-1"),
			},
		}
		result := r.groupRows(rows)
		assert.Len(t, result, 1)
		assert.Len(t, result[0].Scopes, 2)

		scopeTypes := make(map[string]bool)
		for _, sc := range result[0].Scopes {
			scopeTypes[sc.ResourceType] = true
		}
		assert.True(t, scopeTypes[schema.OrganizationNamespace])
		assert.True(t, scopeTypes[schema.ProjectNamespace])
	})

	t.Run("preserves PAT order", func(t *testing.T) {
		rows := []OrgPATRow{
			{PATID: "pat-a", PATTitle: "a", CreatedByID: "u1", CreatedByTitle: "U", CreatedByEmail: "u@t.c", CreatedAt: now, ExpiresAt: now},
			{PATID: "pat-b", PATTitle: "b", CreatedByID: "u2", CreatedByTitle: "V", CreatedByEmail: "v@t.c", CreatedAt: now, ExpiresAt: now},
			{PATID: "pat-c", PATTitle: "c", CreatedByID: "u3", CreatedByTitle: "W", CreatedByEmail: "w@t.c", CreatedAt: now, ExpiresAt: now},
		}
		result := r.groupRows(rows)
		assert.Equal(t, "pat-a", result[0].ID)
		assert.Equal(t, "pat-b", result[1].ID)
		assert.Equal(t, "pat-c", result[2].ID)
	})

	t.Run("sets UserID for all-projects resolution", func(t *testing.T) {
		rows := []OrgPATRow{
			{PATID: "pat-1", PATTitle: "t", CreatedByID: "user-42", CreatedByTitle: "U", CreatedByEmail: "u@t.c", CreatedAt: now, ExpiresAt: now},
		}
		result := r.groupRows(rows)
		assert.Equal(t, "user-42", result[0].UserID)
	})
}

func TestOrgPATsRepository_groupRows_scopeKeyGrouping(t *testing.T) {
	r := OrgPATsRepository{}
	strPtr := func(s string) *string { return &s }
	now := time.Now()

	t.Run("same role different resource types create separate scopes", func(t *testing.T) {
		// A role that supports both org and project scope (future case)
		rows := []OrgPATRow{
			{
				PATID: "pat-1", PATTitle: "t", CreatedByID: "u1", CreatedByTitle: "U", CreatedByEmail: "u@t.c",
				CreatedAt: now, ExpiresAt: now,
				RoleID: strPtr("role-dual"), ResourceType: strPtr(schema.OrganizationNamespace), ResourceID: strPtr("org-1"),
			},
			{
				PATID: "pat-1", PATTitle: "t", CreatedByID: "u1", CreatedByTitle: "U", CreatedByEmail: "u@t.c",
				CreatedAt: now, ExpiresAt: now,
				RoleID: strPtr("role-dual"), ResourceType: strPtr(schema.ProjectNamespace), ResourceID: strPtr("proj-1"),
			},
		}
		result := r.groupRows(rows)
		assert.Len(t, result, 1)
		assert.Len(t, result[0].Scopes, 2)

		for _, sc := range result[0].Scopes {
			assert.Equal(t, "role-dual", sc.RoleID)
		}
	})
}

// Verify the domain model has the correct structure
func TestAggregatedPAT_Structure(t *testing.T) {
	pat := svc.AggregatedPAT{
		ID:    "pat-1",
		Title: "test-token",
		CreatedBy: svc.CreatedBy{
			ID:    "user-1",
			Title: "John Doe",
			Email: "john@test.com",
		},
		Scopes: []patmodels.PATScope{
			{RoleID: "role-1", ResourceType: "app/organization", ResourceIDs: []string{"org-1"}},
		},
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(24 * time.Hour),
		UserID:    "user-1",
	}

	assert.Equal(t, "pat-1", pat.ID)
	assert.Equal(t, "John Doe", pat.CreatedBy.Title)
	assert.Len(t, pat.Scopes, 1)
	assert.Nil(t, pat.LastUsedAt)
}
