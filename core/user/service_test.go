package user_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/core/user/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/str"
	"github.com/stretchr/testify/mock"
)

func mockService(t *testing.T) (*mocks.Repository, *mocks.RelationService, *mocks.PolicyService, *mocks.RoleService) {
	t.Helper()

	repo := mocks.NewRepository(t)
	relationService := mocks.NewRelationService(t)
	policyService := mocks.NewPolicyService(t)
	roleService := mocks.NewRoleService(t)
	return repo, relationService, policyService, roleService
}

func TestService_GetByID(t *testing.T) {
	testID := uuid.New()
	tests := []struct {
		name    string
		setup   func() *user.Service
		id      string
		want    user.User
		wantErr bool
	}{
		{
			name: "get user by id",
			id:   testID.String(),
			want: user.User{
				ID:   testID.String(),
				Name: "test",
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByID(mock.Anything, testID.String()).Return(user.User{
					ID:   testID.String(),
					Name: "test",
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name: "get by email",
			id:   "Test@Test.com",
			want: user.User{
				ID:    testID.String(),
				Name:  "test",
				Email: "test@test.com",
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByEmail(mock.Anything, "test@test.com").Return(user.User{
					ID:    testID.String(),
					Name:  "test",
					Email: "test@test.com",
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name: "get by name",
			id:   "Test",
			want: user.User{
				ID:   testID.String(),
				Name: "test",
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByName(mock.Anything, "test").Return(user.User{
					ID:   testID.String(),
					Name: "test",
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name:    "invalid id should fail",
			id:      "invalid",
			want:    user.User{},
			wantErr: true,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByName(mock.Anything, "invalid").Return(user.User{}, errors.New("not found"))
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.GetByID(context.Background(), tt.id)
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
	tests := []struct {
		name    string
		user    user.User
		want    user.User
		wantErr bool
		setup   func() *user.Service
	}{
		{
			name: "create user",
			user: user.User{
				ID:     "test-id",
				Name:   "test",
				Email:  "test@email.com",
				State:  "enable",
				Avatar: "abc",
				Title:  "tesT",
				Metadata: map[string]any{
					"key": "value",
				},
			},
			want: user.User{
				ID:     "test-id",
				Name:   "test",
				Email:  "test@email.com",
				Avatar: "abc",
				Title:  "tesT",
				State:  user.Enabled,
				Metadata: map[string]any{
					"key": "value",
				},
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().Create(mock.Anything, user.User{
					Name:   "test",
					Email:  "test@email.com",
					State:  user.Enabled,
					Avatar: "abc",
					Title:  "tesT",
					Metadata: map[string]any{
						"key": "value",
					},
				}).Return(user.User{
					ID:     "test-id",
					Name:   "test",
					Email:  "test@email.com",
					State:  user.Enabled,
					Avatar: "abc",
					Title:  "tesT",
					Metadata: map[string]any{
						"key": "value",
					},
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name: "fail on create error",
			user: user.User{
				Name:  "test ",
				Email: "test",
			},
			want:    user.User{},
			wantErr: true,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().Create(mock.Anything, user.User{
					Name:  "test ",
					Email: "test",
					State: user.Enabled,
				}).Return(user.User{}, errors.New("failed to create"))
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Create(context.Background(), tt.user)
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
	tests := []struct {
		name    string
		flt     user.Filter
		want    []user.User
		wantErr bool
		setup   func() *user.Service
	}{
		{
			name: "list users by org id",
			flt: user.Filter{
				OrgID: "org-id",
			},
			want: []user.User{
				{
					ID:   "test-id",
					Name: "test",
				},
				{
					ID:   "test-id-2",
					Name: "test-2",
				},
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				policyService.EXPECT().List(mock.Anything, policy.Filter{
					OrgID: "org-id",
				}).Return([]policy.Policy{
					{
						RoleID:        "role-id",
						PrincipalID:   "test-id",
						PrincipalType: schema.UserPrincipal,
					},
					{
						RoleID:        "role-id",
						PrincipalID:   "test-id-2",
						PrincipalType: schema.UserPrincipal,
					},
				}, nil)

				repo.EXPECT().GetByIDs(mock.Anything, []string{"test-id", "test-id-2"}).Return([]user.User{
					{
						ID:   "test-id",
						Name: "test",
					},
					{
						ID:   "test-id-2",
						Name: "test-2",
					},
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name: "list users by group id",
			flt: user.Filter{
				GroupID: "group-id",
			},
			want: []user.User{
				{
					ID:   "test-id",
					Name: "test",
				},
				{
					ID:   "test-id-2",
					Name: "test-2",
				},
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				policyService.EXPECT().List(mock.Anything, policy.Filter{
					GroupID: "group-id",
				}).Return([]policy.Policy{
					{
						RoleID:        "role-id",
						PrincipalID:   "test-id",
						PrincipalType: schema.UserPrincipal,
					},
					{
						RoleID:        "role-id",
						PrincipalID:   "test-id-2",
						PrincipalType: schema.UserPrincipal,
					},
				}, nil)

				repo.EXPECT().GetByIDs(mock.Anything, []string{"test-id", "test-id-2"}).Return([]user.User{
					{
						ID:   "test-id",
						Name: "test",
					},
					{
						ID:   "test-id-2",
						Name: "test-2",
					},
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name: "list users by state",
			flt: user.Filter{
				State: user.Enabled,
			},
			want: []user.User{
				{
					ID:   "test-id",
					Name: "test",
				},
				{
					ID:   "test-id-2",
					Name: "test-2",
				},
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().List(mock.Anything, user.Filter{
					State: user.Enabled,
				}).Return([]user.User{
					{
						ID:   "test-id",
						Name: "test",
					},
					{
						ID:   "test-id-2",
						Name: "test-2",
					},
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.List(context.Background(), tt.flt)
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

func TestService_Update(t *testing.T) {
	testID := uuid.New()
	tests := []struct {
		name    string
		user    user.User
		want    user.User
		wantErr bool
		setup   func() *user.Service
	}{
		{
			name: "update user by email",
			user: user.User{
				ID:     "test@email.com",
				Name:   "test",
				Avatar: "abc",
				Title:  "tesT",
				Metadata: map[string]any{
					"key": "value",
				},
			},
			want: user.User{
				ID:     "test-id",
				Name:   "test",
				Email:  "test@email.com",
				Avatar: "abc",
				Title:  "tesT",
				State:  user.Enabled,
				Metadata: map[string]any{
					"key": "value",
				},
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().UpdateByEmail(mock.Anything, user.User{
					ID:     "test@email.com",
					Name:   "test",
					Avatar: "abc",
					Title:  "tesT",
					Metadata: map[string]any{
						"key": "value",
					},
				}).Return(user.User{
					ID:     "test-id",
					Name:   "test",
					Email:  "test@email.com",
					State:  user.Enabled,
					Avatar: "abc",
					Title:  "tesT",
					Metadata: map[string]any{
						"key": "value",
					},
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name: "update user by name",
			user: user.User{
				ID:     "test",
				Name:   "test",
				Avatar: "abc",
				Title:  "tesT",
				Metadata: map[string]any{
					"key": "value",
				},
			},
			want: user.User{
				ID:     "test-id",
				Name:   "test",
				Email:  "test@email.com",
				Avatar: "abc",
				Title:  "tesT",
				State:  user.Enabled,
				Metadata: map[string]any{
					"key": "value",
				},
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().UpdateByName(mock.Anything, user.User{
					ID:     "test",
					Name:   "test",
					Avatar: "abc",
					Title:  "tesT",
					Metadata: map[string]any{
						"key": "value",
					},
				}).Return(user.User{
					ID:     "test-id",
					Name:   "test",
					Email:  "test@email.com",
					State:  user.Enabled,
					Avatar: "abc",
					Title:  "tesT",
					Metadata: map[string]any{
						"key": "value",
					},
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name: "update user by id",
			user: user.User{
				ID:     testID.String(),
				Name:   "test",
				Avatar: "abc",
				Title:  "tesT",
				Metadata: map[string]any{
					"key": "value",
				},
			},
			want: user.User{
				ID:     testID.String(),
				Name:   "test",
				Email:  "test@email.com",
				Avatar: "abc",
				Title:  "tesT",
				State:  user.Enabled,
				Metadata: map[string]any{
					"key": "value",
				},
				CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
			},
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().UpdateByID(mock.Anything, user.User{
					ID:     testID.String(),
					Name:   "test",
					Avatar: "abc",
					Title:  "tesT",
					Metadata: map[string]any{
						"key": "value",
					},
				}).Return(user.User{
					ID:     testID.String(),
					Name:   "test",
					Email:  "test@email.com",
					State:  user.Enabled,
					Avatar: "abc",
					Title:  "tesT",
					Metadata: map[string]any{
						"key": "value",
					},
					CreatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
					UpdatedAt: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name: "fail on update error",
			user: user.User{
				Name:  "test ",
				Email: "test",
			},
			want:    user.User{},
			wantErr: true,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().UpdateByName(mock.Anything, user.User{
					Name:  "test ",
					Email: "test",
				}).Return(user.User{}, errors.New("failed to update"))
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Update(context.Background(), tt.user)
			if (err != nil) != tt.wantErr {
				t.Errorf("Update() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(got, tt.want); diff != "" {
				t.Errorf("Update() mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	tests := []struct {
		name    string
		id      string
		setup   func() *user.Service
		wantErr bool
	}{
		{
			name:    "while deleting user, delete it's relations",
			id:      "test-id",
			wantErr: false,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				relationService.EXPECT().Delete(mock.Anything, relation.Relation{Subject: relation.Subject{
					ID:        "test-id",
					Namespace: schema.UserPrincipal,
				}}).Return(nil)
				repo.EXPECT().Delete(mock.Anything, "test-id").Return(nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name:    "fail on delete relations shouldn't remove user",
			id:      "test-id",
			wantErr: true,
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				relationService.EXPECT().Delete(mock.Anything, relation.Relation{Subject: relation.Subject{
					ID:        "test-id",
					Namespace: schema.UserPrincipal,
				}}).Return(errors.New("failed to delete relation"))
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if err := s.Delete(context.Background(), tt.id); (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Sudo(t *testing.T) {
	type args struct {
		id           string
		relationName string
	}
	tests := []struct {
		name    string
		setup   func() *user.Service
		args    args
		wantErr bool
	}{
		{
			name:    "create user admin of platform",
			wantErr: false,
			args: args{
				id:           "test-id",
				relationName: schema.AdminRelationName,
			},
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByName(mock.Anything, "test-id").Return(user.User{
					ID:   "test-id",
					Name: "test",
				}, nil)

				relationService.EXPECT().BatchCheckPermission(mock.Anything, []relation.Relation{
					{
						Subject: relation.Subject{
							ID:        "test-id",
							Namespace: schema.UserPrincipal,
						},
						Object: relation.Object{
							ID:        schema.PlatformID,
							Namespace: schema.PlatformNamespace,
						},
						RelationName: schema.PlatformSudoPermission,
					},
				}).Return([]relation.CheckPair{
					{
						Relation: relation.Relation{
							Subject: relation.Subject{
								ID:        "test-id",
								Namespace: schema.UserPrincipal,
							},
							Object: relation.Object{
								ID:        schema.PlatformID,
								Namespace: schema.PlatformNamespace,
							},
							RelationName: schema.PlatformSudoPermission,
						},
						Status: false,
					},
				}, nil)

				relationService.EXPECT().Create(mock.Anything, relation.Relation{
					Object: relation.Object{
						ID:        schema.PlatformID,
						Namespace: schema.PlatformNamespace,
					},
					Subject: relation.Subject{
						ID:        "test-id",
						Namespace: schema.UserPrincipal,
					},
					RelationName: schema.AdminRelationName,
				}).Return(relation.Relation{}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name:    "don't create user admin if already created",
			wantErr: false,
			args: args{
				id:           "test-id",
				relationName: schema.AdminRelationName,
			},
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByName(mock.Anything, "test-id").Return(user.User{
					ID:   "test-id",
					Name: "test",
				}, nil)

				relationService.EXPECT().BatchCheckPermission(mock.Anything, []relation.Relation{
					{
						Subject: relation.Subject{
							ID:        "test-id",
							Namespace: schema.UserPrincipal,
						},
						Object: relation.Object{
							ID:        schema.PlatformID,
							Namespace: schema.PlatformNamespace,
						},
						RelationName: schema.PlatformSudoPermission,
					},
				}).Return([]relation.CheckPair{
					{
						Relation: relation.Relation{
							Subject: relation.Subject{
								ID:        "test-id",
								Namespace: schema.UserPrincipal,
							},
							Object: relation.Object{
								ID:        schema.PlatformID,
								Namespace: schema.PlatformNamespace,
							},
							RelationName: schema.PlatformSudoPermission,
						},
						Status: true,
					},
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name:    "create user and make it member",
			wantErr: false,
			args: args{
				id:           "test@test.com",
				relationName: schema.MemberRelationName,
			},
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByEmail(mock.Anything, "test@test.com").Return(user.User{}, user.ErrNotExist)

				repo.EXPECT().Create(mock.Anything, user.User{
					Email: "test@test.com",
					Name:  str.GenerateUserSlug("test@test.com"),
					State: user.Enabled,
				}).Return(user.User{
					ID:    "test-id",
					Email: "test@test.com",
					Name:  str.GenerateUserSlug("test@test.com"),
					State: user.Enabled,
				}, nil)

				relationService.EXPECT().BatchCheckPermission(mock.Anything, []relation.Relation{
					{
						Subject: relation.Subject{
							ID:        "test-id",
							Namespace: schema.UserPrincipal,
						},
						Object: relation.Object{
							ID:        schema.PlatformID,
							Namespace: schema.PlatformNamespace,
						},
						RelationName: schema.PlatformCheckPermission,
					},
				}).Return([]relation.CheckPair{
					{
						Relation: relation.Relation{
							Subject: relation.Subject{
								ID:        "test-id",
								Namespace: schema.UserPrincipal,
							},
							Object: relation.Object{
								ID:        schema.PlatformID,
								Namespace: schema.PlatformNamespace,
							},
							RelationName: schema.PlatformCheckPermission,
						},
						Status: false,
					},
				}, nil)

				relationService.EXPECT().Create(mock.Anything, relation.Relation{
					Object: relation.Object{
						ID:        schema.PlatformID,
						Namespace: schema.PlatformNamespace,
					},
					Subject: relation.Subject{
						ID:        "test-id",
						Namespace: schema.UserPrincipal,
					},
					RelationName: schema.MemberRelationName,
				}).Return(relation.Relation{}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if err := s.Sudo(context.Background(), tt.args.id, tt.args.relationName); (err != nil) != tt.wantErr {
				t.Errorf("Sudo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_UnSudo(t *testing.T) {
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		setup   func() *user.Service
		args    args
		wantErr bool
	}{
		{
			name:    "remove user member permission of platform",
			wantErr: false,
			args: args{
				id: "test-id",
			},
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByName(mock.Anything, "test-id").Return(user.User{
					ID:   "test-id",
					Name: "test",
				}, nil)

				relationService.EXPECT().BatchCheckPermission(mock.Anything, []relation.Relation{
					{
						Subject: relation.Subject{
							ID:        "test-id",
							Namespace: schema.UserPrincipal,
						},
						Object: relation.Object{
							ID:        schema.PlatformID,
							Namespace: schema.PlatformNamespace,
						},
						RelationName: schema.PlatformCheckPermission,
					},
				}).Return([]relation.CheckPair{
					{
						Relation: relation.Relation{
							Subject: relation.Subject{
								ID:        "test-id",
								Namespace: schema.UserPrincipal,
							},
							Object: relation.Object{
								ID:        schema.PlatformID,
								Namespace: schema.PlatformNamespace,
							},
							RelationName: schema.PlatformCheckPermission,
						},
						Status: true,
					},
				}, nil)

				relationService.EXPECT().Delete(mock.Anything, relation.Relation{
					Object: relation.Object{
						ID:        schema.PlatformID,
						Namespace: schema.PlatformNamespace,
					},
					Subject: relation.Subject{
						ID:        "test-id",
						Namespace: schema.UserPrincipal,
					},
					RelationName: schema.MemberRelationName,
				}).Return(nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
		{
			name:    "don't remove user member if already removed",
			wantErr: false,
			args: args{
				id: "test-id",
			},
			setup: func() *user.Service {
				repo, relationService, policyService, roleService := mockService(t)
				repo.EXPECT().GetByName(mock.Anything, "test-id").Return(user.User{
					ID:   "test-id",
					Name: "test",
				}, nil)

				relationService.EXPECT().BatchCheckPermission(mock.Anything, []relation.Relation{
					{
						Subject: relation.Subject{
							ID:        "test-id",
							Namespace: schema.UserPrincipal,
						},
						Object: relation.Object{
							ID:        schema.PlatformID,
							Namespace: schema.PlatformNamespace,
						},
						RelationName: schema.PlatformCheckPermission,
					},
				}).Return([]relation.CheckPair{
					{
						Relation: relation.Relation{
							Subject: relation.Subject{
								ID:        "test-id",
								Namespace: schema.UserPrincipal,
							},
							Object: relation.Object{
								ID:        schema.PlatformID,
								Namespace: schema.PlatformNamespace,
							},
							RelationName: schema.PlatformCheckPermission,
						},
						Status: false,
					},
				}, nil)
				return user.NewService(repo, relationService, policyService, roleService)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			if err := s.UnSudo(context.Background(), tt.args.id); (err != nil) != tt.wantErr {
				t.Errorf("UnSudo() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
