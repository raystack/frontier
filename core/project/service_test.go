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

func TestService_List_WithPrincipal(t *testing.T) {
	ctx := context.Background()
	userPrincipal := authenticate.Principal{ID: "user-id", Type: schema.UserPrincipal}

	tests := []struct {
		name    string
		setup   func(*testing.T) *project.Service
		filter  project.Filter
		want    []project.Project
		wantErr bool
	}{
		{
			name:   "errors when membership service is not wired",
			filter: project.Filter{Principal: &userPrincipal},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				// Intentionally skip SetMembershipService.
				return project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
			},
			wantErr: true,
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "", false).
					Return([]string{"p1", "p2"}, nil)
				repo.EXPECT().
					List(ctx, project.Filter{Principal: &userPrincipal, ProjectIDs: []string{"p1", "p2"}}).
					Return([]project.Project{{ID: "p1", Name: "p1"}, {ID: "p2", Name: "p2"}}, nil)
				svc := project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "org-1", true).
					Return([]string{"p1"}, nil)
				repo.EXPECT().
					List(ctx, project.Filter{Principal: &userPrincipal, OrgID: "org-1", NonInherited: true, ProjectIDs: []string{"p1"}}).
					Return([]project.Project{{ID: "p1", Organization: organization.Organization{ID: "org-1"}}}, nil)
				svc := project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "", false).
					Return([]string{"p1", "p2", "p3"}, nil)
				repo.EXPECT().
					List(ctx, project.Filter{Principal: &userPrincipal, ProjectIDs: []string{"p2", "p3"}}).
					Return([]project.Project{{ID: "p2"}, {ID: "p3"}}, nil)
				svc := project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "", false).
					Return(nil, nil)
				// repo.List must NOT be called.
				svc := project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
				svc.SetMembershipService(membershipService)
				return svc
			},
		},
		{
			name:   "propagates membership shim error",
			filter: project.Filter{Principal: &userPrincipal},
			setup: func(t *testing.T) *project.Service {
				t.Helper()
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
				membershipService := mocks.NewMembershipService(t)
				membershipService.EXPECT().
					ListProjectsByPrincipal(ctx, userPrincipal, "", false).
					Return(nil, errors.New("membership boom"))
				svc := project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
				repo, userService, suserService, relationService, policyService, authnService, groupService, roleService := mockService(t)
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
				svc := project.NewService(repo, relationService, userService, policyService, authnService, suserService, groupService, roleService)
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
		{
			name:          "should skip delete+create when role is unchanged",
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
				// user already has the same role on this project
				policySvc.EXPECT().List(ctx, policy.Filter{
					ProjectID: projectID, PrincipalID: userID, PrincipalType: schema.UserPrincipal,
				}).Return([]policy.Policy{{ID: "existing-p1", RoleID: roleID}}, nil)
				// no Delete or Create should be called — early return
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
