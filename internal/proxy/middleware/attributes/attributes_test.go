package attributes

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/raystack/salt/log"
	"github.com/raystack/shield/core/organization"
	"github.com/raystack/shield/core/project"
	"github.com/raystack/shield/core/rule"
	"github.com/raystack/shield/internal/proxy/middleware"
)

var testPermissionAttributesMap = map[string]any{
	"project":       "85be2dfe-7b13-42aa-96f8-f040afb0bbb3",
	"team":          "e16f46cf-6e1e-4802-967a-ea4008ee0ca3",
	"resource":      "p-gojek-test-firehose-nb-test-11",
	"resource_type": "firehose",
	"user":          "nihar.b.interns@aux.gojek.com",
	"namespace":     "odin",
	"prefix":        "/api",
	"organization":  "39e63abd-0fb0-4f5a-ac24-92bf83e1f920",
}

var testProjectMap = map[string]project.Project{
	"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71": {
		ID:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
		Name: "Prj 1",
		Slug: "prj-1",
		Metadata: map[string]any{
			"email": "org1@org1.com",
		},
		Organization: organization.Organization{
			ID:   "org1",
			Name: "Org 1",
			Slug: "Org Slug 1",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"85be2dfe-7b13-42aa-96f8-f040afb0bbb3": {
		ID:   "85be2dfe-7b13-42aa-96f8-f040afb0bbb3",
		Name: "Prj 2",
		Slug: "prj-2",
		Metadata: map[string]any{
			"email": "org1@org2.com",
		},
		Organization: organization.Organization{
			ID:   "39e63abd-0fb0-4f5a-ac24-92bf83e1f920",
			Name: "Org 2",
			Slug: "Org Slug 2",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
	"project-3-slug": {
		ID:   "c3772d61-faa1-4d8d-fff3-c8fa5a1fdc4b",
		Name: "Prj 3",
		Slug: "project-3-slug",
		Metadata: map[string]any{
			"email": "org1@org2.com",
		},
		Organization: organization.Organization{
			ID:   "org2",
			Name: "Org 2",
			Slug: "Org Slug 2",
		},
		CreatedAt: time.Time{},
		UpdatedAt: time.Time{},
	},
}

func TestExtractMiddleware(t *testing.T) {
	t.Parallel()

	odinRule := rule.Rule{
		Frontend: rule.Frontend{
			URL:    "/api/firehoses",
			Method: "POST",
		},
		Backend: rule.Backend{
			URL:       "http://localhost:3000",
			Namespace: "odin",
			Prefix:    "/api",
		},
		Middlewares: rule.MiddlewareSpecs{
			rule.MiddlewareSpec{
				Name: "attributes",
				Config: map[string]interface{}{
					"attributes": map[string]any{
						"project": map[string]any{
							"key":    "X-Shield-Project",
							"source": "request",
							"type":   "header",
						},
						"resource": map[string]any{
							"key":    "name",
							"source": "request",
							"type":   "json_payload",
						},
						"resource_type": map[string]any{
							"type":  "constant",
							"value": "firehose",
						},
						"team": map[string]any{
							"key":    "X-Shield-Group",
							"source": "request",
							"type":   "header",
						},
					},
				},
			},
		},
	}

	postBody, _ := json.Marshal(map[string]string{
		"name": "p-gojek-test-firehose-nb-test-11",
	})

	table := []struct {
		title                string
		permissionAttributes map[string]any
		rw                   http.ResponseWriter
		req                  *http.Request
		name                 string
		a                    Attributes
		want                 map[string]any
		ok                   bool
	}{
		{
			title:                "success",
			permissionAttributes: testPermissionAttributesMap,
			rw:                   nil,
			req: &http.Request{
				Method: "POST",
				Header: map[string][]string{
					"X-Shield-Project": {"85be2dfe-7b13-42aa-96f8-f040afb0bbb3"},
					"X-Auth-Email":     {"nihar.b.interns@aux.gojek.com"},
					"X-Shield-Group":   {"e16f46cf-6e1e-4802-967a-ea4008ee0ca3"},
				},
				Body: ioutil.NopCloser(bytes.NewBuffer(postBody)),
			},
			a: Attributes{
				log:                    log.NewNoop(),
				next:                   &mockNextHandler{},
				identityProxyHeaderKey: "X-Auth-Email",
				projectService: mockProject{
					GetProjectFunc: func(ctx context.Context, id string) (project.Project, error) {
						return testProjectMap[id], nil
					},
				},
			},
			want: testPermissionAttributesMap,
			ok:   true,
		},
		{
			title:                "error in getting organization",
			permissionAttributes: testPermissionAttributesMap,
			rw:                   nil,
			req: &http.Request{
				Method: "POST",
				Header: map[string][]string{
					"X-Shield-Project": {"a5ae3dfe-7b13-42aa-96f8-f040afb0bbb3"},
					"X-Auth-Email":     {"nihar.b.interns@aux.gojek.com"},
					"X-Shield-Group":   {"e16f46cf-6e1e-4802-967a-ea4008ee0ca3"},
				},
				Body: ioutil.NopCloser(bytes.NewBuffer(postBody)),
			},
			a: Attributes{
				log:                    log.NewNoop(),
				next:                   &mockNextHandler{},
				identityProxyHeaderKey: "X-Auth-Email",
				projectService: mockProject{
					GetProjectFunc: func(ctx context.Context, id string) (project.Project, error) {
						return project.Project{}, fmt.Errorf("Project not found")
					},
				},
			},
			want: map[string]any{
				"project":       "a5ae3dfe-7b13-42aa-96f8-f040afb0bbb3",
				"team":          "e16f46cf-6e1e-4802-967a-ea4008ee0ca3",
				"resource":      "p-gojek-test-firehose-nb-test-11",
				"resource_type": "firehose",
				"user":          "nihar.b.interns@aux.gojek.com",
				"namespace":     "odin",
				"prefix":        "/api",
				"organization":  "",
			},
			ok: true,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			middleware.EnrichRule(tt.req, &odinRule)
			middleware.EnrichPathParams(tt.req, map[string]string{})
			w := httptest.NewRecorder()

			tt.a.ServeHTTP(w, tt.req)
			ctx := tt.a.next.(*mockNextHandler).getContext()
			attrMap, ok := GetAttributesFromContext(ctx)
			assert.EqualValues(t, tt.ok, ok)
			assert.EqualValues(t, tt.want, attrMap)
		})
	}
}

type mockNextHandler struct {
	ctx context.Context
}

func (m *mockNextHandler) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	m.ctx = req.Context()
	return
}

func (m *mockNextHandler) getContext() context.Context {
	return m.ctx
}

type mockProject struct {
	GetProjectFunc func(ctx context.Context, id string) (project.Project, error)
}

func (m mockProject) Get(ctx context.Context, id string) (project.Project, error) {
	return m.GetProjectFunc(ctx, id)
}
