package policy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/policy/mocks"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/bootstrap/schema"
)

func mockService(t *testing.T) (*mocks.Repository, *mocks.RoleService, *mocks.RelationService) {
	t.Helper()
	repo := mocks.NewRepository(t)
	roleService := mocks.NewRoleService(t)
	relationService := mocks.NewRelationService(t)
	return repo, roleService, relationService
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		id      string
		wantErr bool
		setup   func() *policy.Service
	}{
		{
			name:    "delete policy from relation before repository",
			id:      "test-id",
			wantErr: false,
			setup: func() *policy.Service {
				repo, roleService, relationService := mockService(t)
				relationService.On("Delete", ctx, relation.Relation{
					Object: relation.Object{
						ID:        "test-id",
						Namespace: schema.RoleBindingNamespace,
					},
				}).Return(nil)
				repo.On("Delete", ctx, "test-id").Return(nil)
				return policy.NewService(repo, relationService, roleService)
			},
		},
		{
			name:    "delete policy from relation fails shouldn't delete from repository",
			id:      "test-id",
			wantErr: true,
			setup: func() *policy.Service {
				repo, roleService, relationService := mockService(t)
				relationService.On("Delete", ctx, relation.Relation{
					Object: relation.Object{
						ID:        "test-id",
						Namespace: schema.RoleBindingNamespace,
					},
				}).Return(errors.New("relation delete failed"))
				return policy.NewService(repo, relationService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if err := s.Delete(ctx, tt.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		policy  policy.Policy
		want    policy.Policy
		setup   func() *policy.Service
		wantErr bool
	}{
		{
			name: "check role exists before assigning successfully",
			policy: policy.Policy{
				RoleID: "role-id",
			},
			wantErr: true,
			setup: func() *policy.Service {
				repo, roleService, relationService := mockService(t)
				roleService.On("Get", ctx, "role-id").Return(role.Role{}, errors.New("role not found"))
				return policy.NewService(repo, relationService, roleService)
			},
		},
		{
			name: "create policy successfully",
			policy: policy.Policy{
				ID:            "policy-id",
				RoleID:        "role-id",
				ResourceID:    "resource-id",
				ResourceType:  schema.ProjectNamespace,
				PrincipalID:   "user-id",
				PrincipalType: schema.UserPrincipal,
			},
			want: policy.Policy{
				ID:            "policy-id",
				RoleID:        "role-id",
				ResourceID:    "resource-id",
				ResourceType:  schema.ProjectNamespace,
				PrincipalID:   "user-id",
				PrincipalType: schema.UserPrincipal,
			},
			setup: func() *policy.Service {
				repo, roleService, relationService := mockService(t)
				roleService.On("Get", ctx, "role-id").Return(role.Role{ID: "role-id"}, nil)
				repo.On("Upsert", ctx, policy.Policy{
					ID:            "policy-id",
					RoleID:        "role-id",
					ResourceID:    "resource-id",
					ResourceType:  schema.ProjectNamespace,
					PrincipalID:   "user-id",
					PrincipalType: schema.UserPrincipal,
				}).Return(policy.Policy{
					ID:            "policy-id",
					RoleID:        "role-id",
					ResourceID:    "resource-id",
					ResourceType:  schema.ProjectNamespace,
					PrincipalID:   "user-id",
					PrincipalType: schema.UserPrincipal,
				}, nil)

				// assign role
				relationService.On("Create", ctx, relation.Relation{
					Object: relation.Object{
						ID:        "policy-id",
						Namespace: schema.RoleBindingNamespace,
					},
					Subject: relation.Subject{
						ID:              "user-id",
						Namespace:       schema.UserPrincipal,
						SubRelationName: "",
					},
					RelationName: schema.RoleBearerRelationName,
				}).Return(relation.Relation{}, nil)
				relationService.On("Create", ctx, relation.Relation{
					Object: relation.Object{
						ID:        "policy-id",
						Namespace: schema.RoleBindingNamespace,
					},
					Subject: relation.Subject{
						ID:        "role-id",
						Namespace: schema.RoleNamespace,
					},
					RelationName: schema.RoleRelationName,
				}).Return(relation.Relation{}, nil)
				relationService.On("Create", ctx, relation.Relation{
					Object: relation.Object{
						ID:        "resource-id",
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						ID:        "policy-id",
						Namespace: schema.RoleBindingNamespace,
					},
					RelationName: schema.RoleGrantRelationName,
				}).Return(relation.Relation{}, nil)
				return policy.NewService(repo, relationService, roleService)
			},
		},
		{
			name: "create policy using member role for groups",
			policy: policy.Policy{
				ID:            "policy-id",
				RoleID:        "role-id",
				ResourceID:    "resource-id",
				ResourceType:  schema.ProjectNamespace,
				PrincipalID:   "group-id",
				PrincipalType: schema.GroupPrincipal,
			},
			want: policy.Policy{
				ID:            "policy-id",
				RoleID:        "role-id",
				ResourceID:    "resource-id",
				ResourceType:  schema.ProjectNamespace,
				PrincipalID:   "group-id",
				PrincipalType: schema.GroupPrincipal,
			},
			setup: func() *policy.Service {
				repo, roleService, relationService := mockService(t)
				roleService.On("Get", ctx, "role-id").Return(role.Role{ID: "role-id"}, nil)
				repo.On("Upsert", ctx, policy.Policy{
					ID:            "policy-id",
					RoleID:        "role-id",
					ResourceID:    "resource-id",
					ResourceType:  schema.ProjectNamespace,
					PrincipalID:   "group-id",
					PrincipalType: schema.GroupPrincipal,
				}).Return(policy.Policy{
					ID:            "policy-id",
					RoleID:        "role-id",
					ResourceID:    "resource-id",
					ResourceType:  schema.ProjectNamespace,
					PrincipalID:   "group-id",
					PrincipalType: schema.GroupPrincipal,
				}, nil)

				// assign role
				relationService.On("Create", ctx, relation.Relation{
					Object: relation.Object{
						ID:        "policy-id",
						Namespace: schema.RoleBindingNamespace,
					},
					Subject: relation.Subject{
						ID:              "group-id",
						Namespace:       schema.GroupPrincipal,
						SubRelationName: schema.MemberRelationName,
					},
					RelationName: schema.RoleBearerRelationName,
				}).Return(relation.Relation{}, nil)
				relationService.On("Create", ctx, relation.Relation{
					Object: relation.Object{
						ID:        "policy-id",
						Namespace: schema.RoleBindingNamespace,
					},
					Subject: relation.Subject{
						ID:        "role-id",
						Namespace: schema.RoleNamespace,
					},
					RelationName: schema.RoleRelationName,
				}).Return(relation.Relation{}, nil)
				relationService.On("Create", ctx, relation.Relation{
					Object: relation.Object{
						ID:        "resource-id",
						Namespace: schema.ProjectNamespace,
					},
					Subject: relation.Subject{
						ID:        "policy-id",
						Namespace: schema.RoleBindingNamespace,
					},
					RelationName: schema.RoleGrantRelationName,
				}).Return(relation.Relation{}, nil)
				return policy.NewService(repo, relationService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Create(ctx, tt.policy)
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

func TestService_ListRoles(t *testing.T) {
	ctx := context.Background()
	type args struct {
		principalType   string
		principalID     string
		objectNamespace string
		objectID        string
	}
	tests := []struct {
		name    string
		setup   func() *policy.Service
		args    args
		want    []role.Role
		wantErr bool
	}{
		{
			name: "list roles assigned to user",
			args: args{
				principalType:   schema.UserPrincipal,
				principalID:     "user-id",
				objectNamespace: schema.OrganizationNamespace,
				objectID:        "org-id",
			},
			want: []role.Role{
				{
					ID:   "role-id",
					Name: "role-name",
				},
			},
			wantErr: false,
			setup: func() *policy.Service {
				repo, roleService, relationService := mockService(t)
				repo.On("List", ctx, policy.Filter{
					PrincipalType: schema.UserPrincipal,
					PrincipalID:   "user-id",
					OrgID:         "org-id",
				}).Return([]policy.Policy{
					{
						ID:            "policy-id",
						RoleID:        "role-id",
						ResourceID:    "resource-id",
						ResourceType:  schema.ProjectNamespace,
						PrincipalID:   "user-id",
						PrincipalType: schema.UserPrincipal,
					},
				}, nil)
				roleService.On("List", ctx, role.Filter{
					IDs: []string{"role-id"},
				}).Return([]role.Role{
					{
						ID:   "role-id",
						Name: "role-name",
					},
				}, nil)
				return policy.NewService(repo, relationService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.ListRoles(ctx, tt.args.principalType, tt.args.principalID, tt.args.objectNamespace, tt.args.objectID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("ListRoles() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
