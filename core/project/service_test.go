package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/project/mocks"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	pat "github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/assert"
)

func mockService(t *testing.T) (*mocks.Repository, *mocks.UserService, *mocks.ServiceuserService,
	*mocks.RelationService, *mocks.PolicyService, *mocks.AuthnService, *mocks.GroupService, *mocks.RoleService) {
	t.Helper()

	repo := mocks.NewRepository(t)
	relationService := mocks.NewRelationService(t)
	userService := mocks.NewUserService(t)
	suserService := mocks.NewServiceuserService(t)
	policyService := mocks.NewPolicyService(t)
	authnService := mocks.NewAuthnService(t)
	groupService := mocks.NewGroupService(t)
	roleService := mocks.NewRoleService(t)
	return repo, userService, suserService, relationService, policyService, authnService, groupService, roleService
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				repo.EXPECT().GetByID(ctx, tid.String()).Return(project.Project{
					ID:   tid.String(),
					Name: "test",
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				repo.EXPECT().GetByName(ctx, "test").Return(project.Project{
					ID:   tid.String(),
					Name: "test",
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
			name:    "fail to create project if no principal found",
			prj:     testProj,
			wantErr: true,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				authnService.EXPECT().GetPrincipal(ctx).Return(authenticate.Principal{}, errors.New("not found"))
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
		{
			name:    "create project successfully with it's respective policies",
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
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

				relationService.EXPECT().Create(ctx, relation.Relation{
					Object: relation.Object{
						ID:        "project-id",
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						ID:        "org-id",
						Namespace: schema.OrganizationNamespace,
					},
					RelationName: schema.OrganizationRelationName,
				}).Return(relation.Relation{}, nil)

				policyService.EXPECT().Create(ctx, policy.Policy{
					RoleID:        project.OwnerRole,
					ResourceID:    "project-id",
					ResourceType:  schema.ProjectNamespace,
					PrincipalID:   "test-user",
					PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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

func TestService_ListByUser(t *testing.T) {
	ctx := context.Background()
	type args struct {
		principal authenticate.Principal
		flt       project.Filter
	}
	tests := []struct {
		name    string
		setup   func() *project.Service
		args    args
		want    []project.Project
		wantErr bool
	}{
		{
			name: "list all projects by user successfully",
			args: args{
				principal: authenticate.Principal{ID: "user-id", Type: schema.UserPrincipal},
				flt:       project.Filter{},
			},
			want: []project.Project{
				{
					ID:   "project-id",
					Name: "test",
					Organization: organization.Organization{
						ID: "org-id",
					},
				},
				{
					ID:   "project-id-2",
					Name: "test-2",
					Organization: organization.Organization{
						ID: "org-id",
					},
				},
				{
					ID:   "project-id-3",
					Name: "test-3",
					Organization: organization.Organization{
						ID: "org-id-2",
					},
				},
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				relationService.EXPECT().LookupResources(ctx, relation.Relation{
					Object: relation.Object{
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						Namespace: schema.UserPrincipal,
						ID:        "user-id",
					},
					RelationName: project.MemberPermission,
				}).Return([]string{"project-id", "project-id-2", "project-id-3"}, nil)

				repo.EXPECT().List(ctx, project.Filter{
					ProjectIDs: []string{"project-id", "project-id-2", "project-id-3"},
				}).Return([]project.Project{
					{
						ID:   "project-id",
						Name: "test",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
					{
						ID:   "project-id-2",
						Name: "test-2",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
					{
						ID:   "project-id-3",
						Name: "test-3",
						Organization: organization.Organization{
							ID: "org-id-2",
						},
					},
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
		{
			name: "list all projects by user with non-inherited policies (with no groups)",
			args: args{
				principal: authenticate.Principal{ID: "user-id", Type: schema.UserPrincipal},
				flt: project.Filter{
					NonInherited: true,
				},
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				policyService.EXPECT().List(ctx, policy.Filter{
					PrincipalType: schema.UserPrincipal,
					PrincipalID:   "user-id",
					ResourceType:  schema.ProjectNamespace,
				}).Return([]policy.Policy{
					{
						ResourceID:    "project-id",
						ResourceType:  schema.ProjectNamespace,
						PrincipalID:   "user-id",
						PrincipalType: schema.UserPrincipal,
					},
				}, nil)

				groupService.EXPECT().ListByUser(ctx, authenticate.Principal{ID: "user-id", Type: schema.UserPrincipal}, group.Filter{}).Return([]group.Group{}, nil)

				repo.EXPECT().List(ctx, project.Filter{
					ProjectIDs:   []string{"project-id"},
					NonInherited: true,
				}).Return([]project.Project{
					{
						ID:   "project-id",
						Name: "test",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
		{
			name: "list all projects by user with non-inherited policies (with groups)",
			args: args{
				principal: authenticate.Principal{ID: "user-id", Type: schema.UserPrincipal},
				flt: project.Filter{
					NonInherited: true,
				},
			},
			want: []project.Project{
				{
					ID:   "project-id",
					Name: "test",
					Organization: organization.Organization{
						ID: "org-id",
					},
				},
				{
					ID:   "project-id-2",
					Name: "test-2",
					Organization: organization.Organization{
						ID: "org-id",
					},
				},
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				policyService.EXPECT().List(ctx, policy.Filter{
					PrincipalType: schema.UserPrincipal,
					PrincipalID:   "user-id",
					ResourceType:  schema.ProjectNamespace,
				}).Return([]policy.Policy{
					{
						ResourceID:    "project-id",
						ResourceType:  schema.ProjectNamespace,
						PrincipalID:   "user-id",
						PrincipalType: schema.UserPrincipal,
					},
				}, nil)

				groupService.EXPECT().ListByUser(ctx, authenticate.Principal{ID: "user-id", Type: schema.UserPrincipal}, group.Filter{}).Return([]group.Group{
					{
						ID: "group-id",
					},
				}, nil)

				policyService.EXPECT().List(ctx, policy.Filter{
					PrincipalType: schema.GroupPrincipal,
					PrincipalIDs:  []string{"group-id"},
					ResourceType:  schema.ProjectNamespace,
				}).Return([]policy.Policy{
					{
						ResourceID:    "project-id-2",
						ResourceType:  schema.ProjectNamespace,
						PrincipalID:   "group-id",
						PrincipalType: schema.GroupPrincipal,
					},
				}, nil)

				repo.EXPECT().List(ctx, project.Filter{
					ProjectIDs:   []string{"project-id", "project-id-2"},
					NonInherited: true,
				}).Return([]project.Project{
					{
						ID:   "project-id",
						Name: "test",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
					{
						ID:   "project-id-2",
						Name: "test-2",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
		{
			name: "PAT principal should resolve to user and intersect with PAT project scope",
			args: args{
				principal: authenticate.Principal{
					ID:   "pat-456",
					Type: schema.PATPrincipal,
					PAT:  &pat.PAT{ID: "pat-456", UserID: "user-id", OrgID: "org-1"},
				},
				flt: project.Filter{},
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				// LookupResources for user's project memberships (resolved from PAT)
				relationService.EXPECT().LookupResources(ctx, relation.Relation{
					Object: relation.Object{
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						Namespace: schema.UserPrincipal,
						ID:        "user-id",
					},
					RelationName: project.MemberPermission,
				}).Return([]string{"project-id", "project-id-2", "project-id-3"}, nil)

				// LookupResources for PAT's project scope
				relationService.EXPECT().LookupResources(ctx, relation.Relation{
					Object: relation.Object{
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						ID:        "pat-456",
						Namespace: schema.PATPrincipal,
					},
					RelationName: schema.GetPermission,
				}).Return([]string{"project-id"}, nil)

				// Repo called with intersection
				repo.EXPECT().List(ctx, project.Filter{
					ProjectIDs: []string{"project-id"},
				}).Return([]project.Project{
					{
						ID:   "project-id",
						Name: "test",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
		{
			name: "PAT principal with non-inherited should resolve to user and intersect",
			args: args{
				principal: authenticate.Principal{
					ID:   "pat-456",
					Type: schema.PATPrincipal,
					PAT:  &pat.PAT{ID: "pat-456", UserID: "user-id", OrgID: "org-1"},
				},
				flt: project.Filter{
					NonInherited: true,
				},
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				// Direct policies for user (resolved from PAT)
				policyService.EXPECT().List(ctx, policy.Filter{
					PrincipalType: schema.UserPrincipal,
					PrincipalID:   "user-id",
					ResourceType:  schema.ProjectNamespace,
				}).Return([]policy.Policy{
					{
						ResourceID:    "project-id",
						ResourceType:  schema.ProjectNamespace,
						PrincipalID:   "user-id",
						PrincipalType: schema.UserPrincipal,
					},
					{
						ResourceID:    "project-id-2",
						ResourceType:  schema.ProjectNamespace,
						PrincipalID:   "user-id",
						PrincipalType: schema.UserPrincipal,
					},
				}, nil)

				// Group lookup uses user-only principal (no double PAT filtering)
				groupService.EXPECT().ListByUser(ctx, authenticate.Principal{ID: "user-id", Type: schema.UserPrincipal}, group.Filter{}).Return([]group.Group{}, nil)

				// PAT scope intersection
				relationService.EXPECT().LookupResources(ctx, relation.Relation{
					Object: relation.Object{
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						ID:        "pat-456",
						Namespace: schema.PATPrincipal,
					},
					RelationName: schema.GetPermission,
				}).Return([]string{"project-id"}, nil)

				// Repo called with intersection result
				repo.EXPECT().List(ctx, project.Filter{
					ProjectIDs:   []string{"project-id"},
					NonInherited: true,
				}).Return([]project.Project{
					{
						ID:   "project-id",
						Name: "test",
						Organization: organization.Organization{
							ID: "org-id",
						},
					},
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
		{
			name: "PAT principal with no project overlap returns empty",
			args: args{
				principal: authenticate.Principal{
					ID:   "pat-456",
					Type: schema.PATPrincipal,
					PAT:  &pat.PAT{ID: "pat-456", UserID: "user-id", OrgID: "org-1"},
				},
				flt: project.Filter{},
			},
			want:    []project.Project{},
			wantErr: false,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				// User has projects
				relationService.EXPECT().LookupResources(ctx, relation.Relation{
					Object: relation.Object{
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						Namespace: schema.UserPrincipal,
						ID:        "user-id",
					},
					RelationName: project.MemberPermission,
				}).Return([]string{"project-id-1"}, nil)

				// PAT scoped to different projects
				relationService.EXPECT().LookupResources(ctx, relation.Relation{
					Object: relation.Object{
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						ID:        "pat-456",
						Namespace: schema.PATPrincipal,
					},
					RelationName: schema.GetPermission,
				}).Return([]string{"project-id-2"}, nil)

				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.ListByUser(ctx, tt.args.principal, tt.args.flt)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListByUser() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ListByUser() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_ListUsers(t *testing.T) {
	ctx := context.Background()
	type args struct {
		id               string
		permissionFilter string
	}
	tests := []struct {
		name    string
		setup   func() *project.Service
		args    args
		want    []user.User
		wantErr bool
	}{
		{
			name: "list all users of a project without permission filter",
			args: args{
				id: "project-id",
			},
			want: []user.User{
				{
					ID: "user-id",
				},
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				repo.EXPECT().GetByName(ctx, "project-id").Return(project.Project{
					ID: "project-id",
				}, nil)

				policyService.EXPECT().List(ctx, policy.Filter{
					ProjectID: "project-id",
				}).Return([]policy.Policy{
					{
						PrincipalID:   "user-id",
						PrincipalType: schema.UserPrincipal,
						ResourceID:    "project-id",
						ResourceType:  schema.ProjectNamespace,
					},
				}, nil)

				userService.EXPECT().GetByIDs(ctx, []string{"user-id"}).Return([]user.User{
					{
						ID: "user-id",
					},
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.ListUsers(ctx, tt.args.id, tt.args.permissionFilter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ListUsers() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_ListServiceUsers(t *testing.T) {
	ctx := context.Background()
	type args struct {
		id               string
		permissionFilter string
	}
	tests := []struct {
		name    string
		setup   func() *project.Service
		args    args
		want    []serviceuser.ServiceUser
		wantErr bool
	}{
		{
			name: "list no users if none found",
			args: args{
				id: "project-id",
			},
			want:    []serviceuser.ServiceUser{},
			wantErr: false,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				repo.EXPECT().GetByName(ctx, "project-id").Return(project.Project{
					ID: "project-id",
				}, nil)

				relationService.EXPECT().LookupSubjects(ctx, relation.Relation{
					Object: relation.Object{
						ID:        "project-id",
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						Namespace: schema.ServiceUserPrincipal,
					},
					RelationName: "",
				}).Return([]string{}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
		{
			name: "list all users of a project without permission filter",
			args: args{
				id: "project-id",
			},
			want: []serviceuser.ServiceUser{
				{
					ID: "user-id",
				},
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				repo.EXPECT().GetByName(ctx, "project-id").Return(project.Project{
					ID: "project-id",
				}, nil)

				relationService.EXPECT().LookupSubjects(ctx, relation.Relation{
					Object: relation.Object{
						ID:        "project-id",
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						Namespace: schema.ServiceUserPrincipal,
					},
					RelationName: "",
				}).Return([]string{"user-id"}, nil)

				suserService.EXPECT().FilterSudos(ctx, []string{"user-id"}).Return([]string{"user-id"}, nil)
				suserService.EXPECT().GetByIDs(ctx, []string{"user-id"}).Return([]serviceuser.ServiceUser{
					{
						ID: "user-id",
					},
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.ListServiceUsers(ctx, tt.args.id, tt.args.permissionFilter)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListUsers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ListUsers() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_ListGroups(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		id      string
		want    []group.Group
		wantErr bool
		setup   func() *project.Service
	}{
		{
			name:    "list no groups if none found",
			id:      "project-id",
			want:    []group.Group{},
			wantErr: false,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				repo.EXPECT().GetByName(ctx, "project-id").Return(project.Project{
					ID: "project-id",
				}, nil)

				policyService.EXPECT().List(ctx, policy.Filter{
					ProjectID:     "project-id",
					PrincipalType: schema.GroupPrincipal,
				}).Return([]policy.Policy{}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
		{
			name: "list all groups of a project",
			id:   "project-id",
			want: []group.Group{
				{
					ID: "group-id",
				},
			},
			wantErr: false,
			setup: func() *project.Service {
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				repo.EXPECT().GetByName(ctx, "project-id").Return(project.Project{
					ID: "project-id",
				}, nil)

				policyService.EXPECT().List(ctx, policy.Filter{
					ProjectID:     "project-id",
					PrincipalType: schema.GroupPrincipal,
				}).Return([]policy.Policy{
					{
						PrincipalID: "group-id",
					},
				}, nil)

				groupService.EXPECT().GetByIDs(ctx, []string{"group-id"}).Return([]group.Group{
					{
						ID: "group-id",
					},
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.ListGroups(ctx, tt.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListGroups() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ListGroups() mismatch (-want +got):\n%s", diff)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				_ = roleService
				relationService.EXPECT().Delete(ctx, relation.Relation{
					Object: relation.Object{
						ID:        "project-id",
						Namespace: schema.ProjectNamespace,
					},
				}).Return(nil)
				repo.EXPECT().Delete(ctx, "project-id").Return(nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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

func TestService_SetMemberRole(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New().String()
	orgID := uuid.New().String()
	userID := uuid.New().String()
	suID := uuid.New().String()
	groupID := uuid.New().String()
	roleID := uuid.New().String()

	tests := []struct {
		name          string
		projectID     string
		principalID   string
		principalType string
		roleID        string
		setup         func(*mocks.Repository, *mocks.UserService, *mocks.ServiceuserService, *mocks.GroupService, *mocks.PolicyService, *mocks.RoleService)
		wantErr       error
	}{
		{
			name:          "should return error if project does not exist",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{}, project.ErrNotExist)
			},
			wantErr: project.ErrNotExist,
		},
		{
			name:          "should return error if user does not exist",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{}, user.ErrNotExist)
			},
			wantErr: user.ErrNotExist,
		},
		{
			name:          "should return error if service user does not exist",
			projectID:     projectID,
			principalID:   suID,
			principalType: schema.ServiceUserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				suserSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{}, serviceuser.ErrNotExist)
			},
			wantErr: serviceuser.ErrNotExist,
		},
		{
			name:          "should return error if group does not exist",
			projectID:     projectID,
			principalID:   groupID,
			principalType: schema.GroupPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				groupSvc.EXPECT().Get(ctx, groupID).Return(group.Group{}, group.ErrNotExist)
			},
			wantErr: group.ErrNotExist,
		},
		{
			name:          "should return error for invalid principal type",
			projectID:     projectID,
			principalID:   userID,
			principalType: "invalid",
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
			},
			wantErr: project.ErrInvalidPrincipalType,
		},
		{
			name:          "should return error if user is not an org member",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{}, nil)
			},
			wantErr: project.ErrNotOrgMember,
		},
		{
			name:          "should return error if role does not exist",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "p1"}}, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{}, role.ErrNotExist)
			},
			wantErr: role.ErrNotExist,
		},
		{
			name:          "should return error if role scope is not project",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID:         orgID,
					PrincipalID:   userID,
					PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "p1"}}, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.OrganizationNamespace}}, nil)
			},
			wantErr: project.ErrInvalidProjectRole,
		},
		{
			name:          "should succeed for user with no existing project policies",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "org-p1"}}, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: roleID, ResourceID: projectID, ResourceType: schema.ProjectNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
			},
			wantErr: nil,
		},
		{
			name:          "should replace existing policies on role change",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				userSvc.EXPECT().GetByID(ctx, userID).Return(user.User{ID: userID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					OrgID: orgID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "org-p1"}}, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "old-p1"}, {ID: "old-p2"}}, nil)
				policySvc.EXPECT().Delete(ctx, "old-p1").Return(nil)
				policySvc.EXPECT().Delete(ctx, "old-p2").Return(nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: roleID, ResourceID: projectID, ResourceType: schema.ProjectNamespace,
					PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{}, nil)
			},
			wantErr: nil,
		},
		{
			name:          "should return error if service user belongs to different org",
			projectID:     projectID,
			principalID:   suID,
			principalType: schema.ServiceUserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				suserSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: "different-org"}, nil)
			},
			wantErr: project.ErrNotOrgMember,
		},
		{
			name:          "should return error if group belongs to different org",
			projectID:     projectID,
			principalID:   groupID,
			principalType: schema.GroupPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				groupSvc.EXPECT().Get(ctx, groupID).Return(group.Group{ID: groupID, OrganizationID: "different-org"}, nil)
			},
			wantErr: project.ErrNotOrgMember,
		},
		{
			name:          "should succeed for service user principal",
			projectID:     projectID,
			principalID:   suID,
			principalType: schema.ServiceUserPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				suserSvc.EXPECT().Get(ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID}, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal,
				}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: roleID, ResourceID: projectID, ResourceType: schema.ProjectNamespace,
					PrincipalID: suID, PrincipalType: schema.ServiceUserPrincipal,
				}).Return(policy.Policy{}, nil)
			},
			wantErr: nil,
		},
		{
			name:          "should succeed for group principal",
			projectID:     projectID,
			principalID:   groupID,
			principalType: schema.GroupPrincipal,
			roleID:        roleID,
			setup: func(repo *mocks.Repository, userSvc *mocks.UserService, suserSvc *mocks.ServiceuserService, groupSvc *mocks.GroupService, policySvc *mocks.PolicyService, roleSvc *mocks.RoleService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID, Organization: organization.Organization{ID: orgID}}, nil)
				groupSvc.EXPECT().Get(ctx, groupID).Return(group.Group{ID: groupID, OrganizationID: orgID}, nil)
				roleSvc.EXPECT().Get(ctx, roleID).Return(role.Role{ID: roleID, Scopes: []string{schema.ProjectNamespace}}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: groupID, PrincipalType: schema.GroupPrincipal,
				}).Return([]policy.Policy{}, nil)
				policySvc.EXPECT().Create(ctx, policy.Policy{
					RoleID: roleID, ResourceID: projectID, ResourceType: schema.ProjectNamespace,
					PrincipalID: groupID, PrincipalType: schema.GroupPrincipal,
				}).Return(policy.Policy{}, nil)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			userSvc := mocks.NewUserService(t)
			suserSvc := mocks.NewServiceuserService(t)
			groupSvc := mocks.NewGroupService(t)
			policySvc := mocks.NewPolicyService(t)
			roleSvc := mocks.NewRoleService(t)
			relationSvc := mocks.NewRelationService(t)
			authnSvc := mocks.NewAuthnService(t)

			if tt.setup != nil {
				tt.setup(repo, userSvc, suserSvc, groupSvc, policySvc, roleSvc)
			}

			svc := project.NewService(repo, relationSvc, userSvc, policySvc, authnSvc, suserSvc, groupSvc, roleSvc)
			err := svc.SetMemberRole(ctx, tt.projectID, tt.principalID, tt.principalType, tt.roleID)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestService_RemoveMember(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New().String()
	userID := uuid.New().String()

	tests := []struct {
		name          string
		projectID     string
		principalID   string
		principalType string
		setup         func(*mocks.Repository, *mocks.PolicyService)
		wantErr       error
	}{
		{
			name:          "should return error if project does not exist",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			setup: func(repo *mocks.Repository, policySvc *mocks.PolicyService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{}, project.ErrNotExist)
			},
			wantErr: project.ErrNotExist,
		},
		{
			name:          "should return error for invalid principal type",
			projectID:     projectID,
			principalID:   userID,
			principalType: "app/invalid",
			setup: func(repo *mocks.Repository, policySvc *mocks.PolicyService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID}, nil)
			},
			wantErr: project.ErrInvalidPrincipalType,
		},
		{
			name:          "should return error if principal has no project policies",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			setup: func(repo *mocks.Repository, policySvc *mocks.PolicyService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{}, nil)
			},
			wantErr: project.ErrNotMember,
		},
		{
			name:          "should delete all project policies for the principal",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.UserPrincipal,
			setup: func(repo *mocks.Repository, policySvc *mocks.PolicyService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "p1"}, {ID: "p2"}}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
				policySvc.EXPECT().Delete(ctx, "p2").Return(nil)
			},
			wantErr: nil,
		},
		{
			name:          "should work for service user principal",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.ServiceUserPrincipal,
			setup: func(repo *mocks.Repository, policySvc *mocks.PolicyService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.ServiceUserPrincipal,
				}).Return([]policy.Policy{{ID: "p1"}}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
			},
			wantErr: nil,
		},
		{
			name:          "should work for group principal",
			projectID:     projectID,
			principalID:   userID,
			principalType: schema.GroupPrincipal,
			setup: func(repo *mocks.Repository, policySvc *mocks.PolicyService) {
				repo.EXPECT().GetByID(ctx, projectID).Return(project.Project{ID: projectID}, nil)
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.GroupPrincipal,
				}).Return([]policy.Policy{{ID: "p1"}}, nil)
				policySvc.EXPECT().Delete(ctx, "p1").Return(nil)
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := mocks.NewRepository(t)
			policySvc := mocks.NewPolicyService(t)
			relationSvc := mocks.NewRelationService(t)
			userSvc := mocks.NewUserService(t)
			suserSvc := mocks.NewServiceuserService(t)
			groupSvc := mocks.NewGroupService(t)
			roleSvc := mocks.NewRoleService(t)
			authnSvc := mocks.NewAuthnService(t)

			if tt.setup != nil {
				tt.setup(repo, policySvc)
			}

			svc := project.NewService(repo, relationSvc, userSvc, policySvc, authnSvc, suserSvc, groupSvc, roleSvc)
			err := svc.RemoveMember(ctx, tt.projectID, tt.principalID, tt.principalType)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
