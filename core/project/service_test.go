package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/project/mocks"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

func mockService(t *testing.T) (*mocks.Repository, *mocks.RelationService, *mocks.PolicyService, *mocks.AuthnService) {
	t.Helper()

	repo := mocks.NewRepository(t)
	relationService := mocks.NewRelationService(t)
	policyService := mocks.NewPolicyService(t)
	authnService := mocks.NewAuthnService(t)
	return repo, relationService, policyService, authnService
}

func TestService_Get(t *testing.T) {
	ctx := context.Background()
	tid := uuid.New()
	tests := []struct {
		name     string
		setup    func() *project.Service
		idOrName string
		want     project.Project
		wantErr  bool
	}{
		{
			name:     "get project by id",
			idOrName: tid.String(),
			want: project.Project{
				ID:   tid.String(),
				Name: "test",
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				repo.EXPECT().GetByID(ctx, tid.String()).Return(project.Project{
					ID:   tid.String(),
					Name: "test",
				}, nil)
				return project.NewService(repo, relationService, policyService, authnService)
			},
		},
		{
			name:     "get project by name",
			idOrName: "test",
			want: project.Project{
				ID:   tid.String(),
				Name: "test",
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				repo.EXPECT().GetByName(ctx, "test").Return(project.Project{
					ID:   tid.String(),
					Name: "test",
				}, nil)
				return project.NewService(repo, relationService, policyService, authnService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Get(ctx, tt.idOrName)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Get() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	testProj := project.Project{
		Name: "test",
		Organization: organization.Organization{
			ID: "org-id",
		},
	}
	tests := []struct {
		name    string
		prj     project.Project
		want    project.Project
		wantErr bool
		setup   func() *project.Service
	}{
		{
			name:    "fail to create project when membership service is not set",
			prj:     testProj,
			wantErr: true,
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				// Intentionally skip SetMembershipService.
				return project.NewService(repo, relationService, policyService, authnService)
			},
		},
		{
			name:    "fail to create project if no principal found",
			prj:     testProj,
			wantErr: true,
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				authnService.EXPECT().GetPrincipal(ctx).Return(authenticate.Principal{}, errors.New("not found"))
				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(mocks.NewMembershipService(t))
				return svc
			},
		},
		{
			name:    "create project successfully and add creator as owner via membership",
			prj:     testProj,
			wantErr: false,
			want: project.Project{
				ID:   "project-id",
				Name: "test",
				Organization: organization.Organization{
					ID: "org-id",
				},
			},
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				authnService.EXPECT().GetPrincipal(ctx).Return(authenticate.Principal{
					ID:   "test-user",
					Type: schema.UserPrincipal,
				}, nil)

				repo.EXPECT().Create(ctx, testProj).Return(project.Project{
					ID:   "project-id",
					Name: "test",
					Organization: organization.Organization{
						ID: "org-id",
					},
				}, nil)

				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().OnProjectCreated(ctx, "project-id", "org-id", "test-user", schema.UserPrincipal).Return(nil)

				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(membershipService)
				return svc
			},
		},
		{
			name:    "delete the project row when membership setup fails",
			prj:     testProj,
			wantErr: true,
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				authnService.EXPECT().GetPrincipal(ctx).Return(authenticate.Principal{
					ID:   "test-user",
					Type: schema.UserPrincipal,
				}, nil)

				repo.EXPECT().Create(ctx, testProj).Return(project.Project{
					ID:   "project-id",
					Name: "test",
					Organization: organization.Organization{
						ID: "org-id",
					},
				}, nil)

				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().OnProjectCreated(ctx, "project-id", "org-id", "test-user", schema.UserPrincipal).Return(errors.New("spicedb unavailable"))
				repo.EXPECT().Delete(ctx, "project-id").Return(nil)

				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(membershipService)
				return svc
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Create(ctx, tt.prj)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Create() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		setup   func() *project.Service
		f       project.Filter
		want    []project.Project
		wantErr bool
	}{
		{
			name: "list projects with org successfully",
			f: project.Filter{
				OrgID: "org-id",
			},
			want: []project.Project{
				{
					ID:   "project-id",
					Name: "test",
					Organization: organization.Organization{
						ID: "org-id",
					},
				},
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				repo.EXPECT().List(ctx, project.Filter{
					OrgID: "org-id",
				}).Return([]project.Project{
					{
						ID:   "project-id",
						Name: "test",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
				}, nil)
				return project.NewService(repo, relationService, policyService, authnService)
			},
		},
		{
			name: "list projects with member count of project",
			f: project.Filter{
				WithMemberCount: true,
			},
			want: []project.Project{
				{
					ID:   "project-id",
					Name: "test",
					Organization: organization.Organization{
						ID: "org-id",
					},
					MemberCount: 1,
				},
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				repo.EXPECT().List(ctx, project.Filter{
					WithMemberCount: true,
				}).Return([]project.Project{
					{
						ID:   "project-id",
						Name: "test",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
				}, nil)

				policyService.EXPECT().ProjectMemberCount(ctx, []string{"project-id"}).Return([]policy.MemberCount{
					{
						ID:    "project-id",
						Count: 1,
					},
				}, nil)
				return project.NewService(repo, relationService, policyService, authnService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.List(ctx, tt.f)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("List() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_List_WithPrincipal(t *testing.T) {
	ctx := context.Background()
	userPrincipal := authenticate.Principal{ID: "68f86fec-eb87-49f0-9be0-8d99b00a4a9c", Type: schema.UserPrincipal}

	tests := []struct {
		name      string
		setup     func(*testing.T) *project.Service
		filter    project.Filter
		want      []project.Project
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "errors when membership service is not wired",
			filter: project.Filter{Principal: &userPrincipal},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				// Intentionally skip SetMembershipService.
				return project.NewService(repo, relationService, policyService, authnService)
			},
			wantErr: true,
		},
		{
			name:   "returns ErrInvalidUUID when Principal has empty ID",
			filter: project.Filter{Principal: &authenticate.Principal{Type: schema.UserPrincipal}},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				return project.NewService(repo, relationService, policyService, authnService)
			},
			wantErr:   true,
			wantErrIs: project.ErrInvalidUUID,
		},
		{
			name:   "returns ErrInvalidUUID when Principal ID is not a valid UUID",
			filter: project.Filter{Principal: &authenticate.Principal{ID: "not-a-uuid", Type: schema.UserPrincipal}},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				return project.NewService(repo, relationService, policyService, authnService)
			},
			wantErr:   true,
			wantErrIs: project.ErrInvalidUUID,
		},
		{
			name:   "returns ErrInvalidPrincipalType when Principal has empty Type",
			filter: project.Filter{Principal: &authenticate.Principal{ID: "68f86fec-eb87-49f0-9be0-8d99b00a4a9c"}},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				return project.NewService(repo, relationService, policyService, authnService)
			},
			wantErr:   true,
			wantErrIs: project.ErrInvalidPrincipalType,
		},
		{
			name:   "returns projects from the membership shim",
			filter: project.Filter{Principal: &userPrincipal},
			want: []project.Project{
				{ID: "p1", Name: "p1"},
				{ID: "p2", Name: "p2"},
			},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "", false).
					Return([]string{"p1", "p2"}, nil)
				repo.EXPECT().
					List(ctx, project.Filter{Principal: &userPrincipal, ProjectIDs: []string{"p1", "p2"}}).
					Return([]project.Project{{ID: "p1", Name: "p1"}, {ID: "p2", Name: "p2"}}, nil)
				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(membershipService)
				return svc
			},
		},
		{
			name:   "passes OrgID and NonInherited through to the shim",
			filter: project.Filter{Principal: &userPrincipal, OrgID: "org-1", NonInherited: true},
			want:   []project.Project{{ID: "p1", Organization: organization.Organization{ID: "org-1"}}},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "org-1", true).
					Return([]string{"p1"}, nil)
				repo.EXPECT().
					List(ctx, project.Filter{Principal: &userPrincipal, OrgID: "org-1", NonInherited: true, ProjectIDs: []string{"p1"}}).
					Return([]project.Project{{ID: "p1", Organization: organization.Organization{ID: "org-1"}}}, nil)
				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(membershipService)
				return svc
			},
		},
		{
			name:   "intersects shim result with caller-supplied ProjectIDs",
			filter: project.Filter{Principal: &userPrincipal, ProjectIDs: []string{"p2", "p3", "p4"}},
			want:   []project.Project{{ID: "p2"}, {ID: "p3"}},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "", false).
					Return([]string{"p1", "p2", "p3"}, nil)
				repo.EXPECT().
					List(ctx, project.Filter{Principal: &userPrincipal, ProjectIDs: []string{"p2", "p3"}}).
					Return([]project.Project{{ID: "p2"}, {ID: "p3"}}, nil)
				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(membershipService)
				return svc
			},
		},
		{
			name:   "short-circuits to empty slice when shim returns no IDs",
			filter: project.Filter{Principal: &userPrincipal},
			want:   []project.Project{},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "", false).
					Return(nil, nil)
				// repo.List must NOT be called.
				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(membershipService)
				return svc
			},
		},
		{
			name:   "propagates membership shim error",
			filter: project.Filter{Principal: &userPrincipal},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "", false).
					Return(nil, errors.New("membership boom"))
				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(membershipService)
				return svc
			},
			wantErr: true,
		},
		{
			name:   "composes Filter.Principal with WithMemberCount enrichment",
			filter: project.Filter{Principal: &userPrincipal, OrgID: "org-1", WithMemberCount: true},
			want: []project.Project{
				{ID: "p1", Organization: organization.Organization{ID: "org-1"}, MemberCount: 5},
				{ID: "p2", Organization: organization.Organization{ID: "org-1"}, MemberCount: 2},
			},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, relationService, policyService, authnService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "org-1", false).
					Return([]string{"p1", "p2"}, nil)
				repo.EXPECT().
					List(ctx, project.Filter{Principal: &userPrincipal, OrgID: "org-1", WithMemberCount: true, ProjectIDs: []string{"p1", "p2"}}).
					Return([]project.Project{
						{ID: "p1", Organization: organization.Organization{ID: "org-1"}},
						{ID: "p2", Organization: organization.Organization{ID: "org-1"}},
					}, nil)
				policyService.EXPECT().
					ProjectMemberCount(ctx, []string{"p1", "p2"}).
					Return([]policy.MemberCount{
						{ID: "p1", Count: 5},
						{ID: "p2", Count: 2},
					}, nil)
				svc := project.NewService(repo, relationService, policyService, authnService)
				svc.SetMembershipService(membershipService)
				return svc
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup(t)
			got, err := s.List(ctx, tt.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("List() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("List() error = %v, want errors.Is(%v)", err, tt.wantErrIs)
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("List() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_DeleteModel(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		id      string
		wantErr bool
		setup   func() *project.Service
	}{
		{
			name:    "delete relations before deleting project successfully",
			id:      "project-id",
			wantErr: false,
			setup: func() *project.Service {
				repo, relationService, policyService, authnService := mockService(t)
				relationService.EXPECT().Delete(ctx, relation.Relation{
					Object: relation.Object{
						ID:        "project-id",
						Namespace: schema.ProjectNamespace,
					},
				}).Return(nil)
				repo.EXPECT().Delete(ctx, "project-id").Return(nil)
				return project.NewService(repo, relationService, policyService, authnService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if err := s.DeleteModel(ctx, tt.id); (err != nil) != tt.wantErr {
				t.Errorf("DeleteModel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
