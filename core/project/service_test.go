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
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

func mockService(t *testing.T) (*mocks.Repository, *mocks.UserService, *mocks.ServiceuserService,
	*mocks.RelationService, *mocks.PolicyService, *mocks.AuthnService, *mocks.GroupService) {
	t.Helper()

	repo := mocks.NewRepository(t)
	relationService := mocks.NewRelationService(t)
	userService := mocks.NewUserService(t)
	suserService := mocks.NewServiceuserService(t)
	policyService := mocks.NewPolicyService(t)
	authnService := mocks.NewAuthnService(t)
	groupService := mocks.NewGroupService(t)
	return repo, userService, suserService, relationService, policyService, authnService, groupService
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
				repo.EXPECT().GetByID(ctx, tid.String()).Return(project.Project{
					ID:   tid.String(),
					Name: "test",
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
				repo.EXPECT().GetByName(ctx, "test").Return(project.Project{
					ID:   tid.String(),
					Name: "test",
				}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
				authnService.EXPECT().GetPrincipal(ctx).Return(authenticate.Principal{}, errors.New("not found"))
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
		principalID   string
		principalType string
		flt           project.Filter
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
				principalID:   "user-id",
				principalType: schema.UserPrincipal,
				flt:           project.Filter{},
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
			},
		},
		{
			name: "list all projects by user with non-inherited policies (with no groups)",
			args: args{
				principalID:   "user-id",
				principalType: schema.UserPrincipal,
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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

				groupService.EXPECT().ListByUser(ctx, "user-id", schema.UserPrincipal, group.Filter{}).Return([]group.Group{}, nil)

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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
			},
		},
		{
			name: "list all projects by user with non-inherited policies (with groups)",
			args: args{
				principalID:   "user-id",
				principalType: schema.UserPrincipal,
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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

				groupService.EXPECT().ListByUser(ctx, "user-id", schema.UserPrincipal, group.Filter{}).Return([]group.Group{
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.ListByUser(ctx, tt.args.principalID, tt.args.principalType, tt.args.flt)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
				repo.EXPECT().GetByName(ctx, "project-id").Return(project.Project{
					ID: "project-id",
				}, nil)

				policyService.EXPECT().List(ctx, policy.Filter{
					ProjectID:     "project-id",
					PrincipalType: schema.GroupPrincipal,
				}).Return([]policy.Policy{}, nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
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
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService := mockService(t)
				relationService.EXPECT().Delete(ctx, relation.Relation{
					Object: relation.Object{
						ID:        "project-id",
						Namespace: schema.ProjectNamespace,
					},
				}).Return(nil)
				repo.EXPECT().Delete(ctx, "project-id").Return(nil)
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService)
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
