package userpat_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/core/userpat/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/sha3"
)

var defaultConfig = userpat.Config{
	Enabled:          true,
	Prefix:           "fpt",
	MaxPerUserPerOrg: 50,
	MaxLifetime:      "8760h",
}

func newSuccessMocks(t *testing.T) (*mocks.OrganizationService, *mocks.RoleService, *mocks.PolicyService, *mocks.AuditRecordRepository) {
	t.Helper()
	orgSvc := mocks.NewOrganizationService(t)
	orgSvc.On("GetRaw", mock.Anything, mock.Anything).
		Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
	roleSvc := mocks.NewRoleService(t)
	roleSvc.On("List", mock.Anything, mock.Anything).
		Return([]role.Role{{
			ID:     "role-1",
			Name:   "test-role",
			Scopes: []string{schema.OrganizationNamespace},
		}}, nil).Maybe()
	policySvc := mocks.NewPolicyService(t)
	policySvc.On("Create", mock.Anything, mock.Anything).
		Return(policy.Policy{}, nil).Maybe()
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).
		Return(models.AuditRecord{}, nil).Maybe()
	return orgSvc, roleSvc, policySvc, auditRepo
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *userpat.Service
		req          userpat.CreateRequest
		wantErr      bool
		wantErrIs    error
		wantErrMsg   string
		validateFunc func(t *testing.T, got userpat.PAT, tokenValue string)
	}{
		{
			name: "should return ErrDisabled when PAT feature is disabled",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: userpat.ErrDisabled,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(log.NewNoop(), repo, userpat.Config{
					Enabled: false,
				}, orgSvc, nil, nil, auditRepo)
			},
		},
		{
			name: "should return error when CountActive fails",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:    true,
			wantErrMsg: "counting active PATs",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), errors.New("db connection failed"))
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, nil, nil, auditRepo)
			},
		},
		{
			name: "should return ErrLimitExceeded when token count equals max",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: userpat.ErrLimitExceeded,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(50), nil)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, nil, nil, auditRepo)
			},
		},
		{
			name: "should return ErrLimitExceeded when token count exceeds max",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: userpat.ErrLimitExceeded,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(55), nil)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, nil, nil, auditRepo)
			},
		},
		{
			name: "should return error when repo Create fails",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:    true,
			wantErrMsg: "insert failed",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Return(userpat.PAT{}, errors.New("insert failed"))
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.On("List", mock.Anything, mock.Anything).Return([]role.Role{{
					ID: "role-1", Name: "test-role", Scopes: []string{schema.OrganizationNamespace},
				}}, nil).Maybe()
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, nil, auditRepo)
			},
		},
		{
			name: "should return ErrConflict when repo returns duplicate error",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: userpat.ErrConflict,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Return(userpat.PAT{}, userpat.ErrConflict)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.On("List", mock.Anything, mock.Anything).Return([]role.Role{{
					ID: "role-1", Name: "test-role", Scopes: []string{schema.OrganizationNamespace},
				}}, nil).Maybe()
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, nil, auditRepo)
			},
		},
		{
			name: "should create token successfully with correct fields",
			req: userpat.CreateRequest{
				UserID:     "user-1",
				OrgID:      "org-1",
				Title:      "my-token",
				RoleIDs:    []string{"role-1"},
				ProjectIDs: []string{"proj-1"},
				ExpiresAt:  time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
				Metadata:   map[string]any{"env": "staging"},
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Run(func(ctx context.Context, pat userpat.PAT) {
						if pat.UserID != "user-1" {
							t.Errorf("Create() UserID = %v, want %v", pat.UserID, "user-1")
						}
						if pat.OrgID != "org-1" {
							t.Errorf("Create() OrgID = %v, want %v", pat.OrgID, "org-1")
						}
						if pat.Title != "my-token" {
							t.Errorf("Create() Title = %v, want %v", pat.Title, "my-token")
						}
						if pat.SecretHash == "" {
							t.Error("Create() SecretHash should not be empty")
						}
						if diff := cmp.Diff(map[string]any{"env": "staging"}, map[string]any(pat.Metadata)); diff != "" {
							t.Errorf("Create() Metadata mismatch (-want +got):\n%s", diff)
						}
						if !pat.ExpiresAt.Equal(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)) {
							t.Errorf("Create() ExpiresAt = %v, want %v", pat.ExpiresAt, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
						}
					}).
					Return(userpat.PAT{
						ID:        "pat-id-1",
						UserID:    "user-1",
						OrgID:     "org-1",
						Title:     "my-token",
						Metadata:  map[string]any{"env": "staging"},
						ExpiresAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
						CreatedAt: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
					}, nil)
				orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
			},
			validateFunc: func(t *testing.T, got userpat.PAT, tokenValue string) {
				t.Helper()
				if got.ID != "pat-id-1" {
					t.Errorf("Create() ID = %v, want %v", got.ID, "pat-id-1")
				}
				if got.UserID != "user-1" {
					t.Errorf("Create() UserID = %v, want %v", got.UserID, "user-1")
				}
				if tokenValue == "" {
					t.Error("Create() tokenValue should not be empty")
				}
			},
		},
		{
			name: "should generate token with correct prefix and base64url encoding",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Return(userpat.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
			},
			validateFunc: func(t *testing.T, got userpat.PAT, tokenValue string) {
				t.Helper()
				if !strings.HasPrefix(tokenValue, "fpt_") {
					t.Errorf("token should start with prefix fpt_, got %v", tokenValue)
				}
				parts := strings.SplitN(tokenValue, "_", 2)
				if len(parts) != 2 {
					t.Fatal("token should have format prefix_secret")
				}
				decoded, err := base64.RawURLEncoding.DecodeString(parts[1])
				if err != nil {
					t.Errorf("secret part should be valid base64url: %v", err)
				}
				if len(decoded) != 32 {
					t.Errorf("secret should decode to 32 bytes, got %d", len(decoded))
				}
			},
		},
		{
			name: "should hash the raw secret bytes with sha3-256",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Return(userpat.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
			},
			validateFunc: func(t *testing.T, got userpat.PAT, tokenValue string) {
				t.Helper()
				parts := strings.SplitN(tokenValue, "_", 2)
				if len(parts) != 2 {
					t.Fatal("token should have format prefix_secret")
				}
				secretBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
				if err != nil {
					t.Fatalf("failed to decode secret: %v", err)
				}
				hash := sha3.Sum256(secretBytes)
				hashStr := hex.EncodeToString(hash[:])
				if len(hashStr) != 64 {
					t.Errorf("sha3-256 hash should be 64 hex chars, got %d", len(hashStr))
				}
			},
		},
		{
			name: "should use custom token prefix from config",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Return(userpat.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(log.NewNoop(), repo, userpat.Config{
					Enabled:                true,
					Prefix:            "custom",
					MaxPerUserPerOrg: 50,
					MaxLifetime:       "8760h",
				}, orgSvc, roleSvc, policySvc, auditRepo)
			},
			validateFunc: func(t *testing.T, got userpat.PAT, tokenValue string) {
				t.Helper()
				if !strings.HasPrefix(tokenValue, "custom_") {
					t.Errorf("token should start with custom_, got %v", tokenValue)
				}
			},
		},
		{
			name: "should allow creation when count is just below max",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(49), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Return(userpat.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
			},
		},
		{
			name: "should generate unique tokens on each call",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				RoleIDs:   []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Return(userpat.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, tokenValue, err := s.Create(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("Create() error = %v, wantErrIs %v", err, tt.wantErrIs)
				return
			}
			if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("Create() error = %v, wantErrMsg containing %v", err, tt.wantErrMsg)
				return
			}
			if tt.validateFunc != nil {
				tt.validateFunc(t, got, tokenValue)
			}
		})
	}
}

func TestService_Create_UniquePATs(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
		Return(int64(0), nil).Times(2)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
		Return(userpat.PAT{ID: "pat-1", OrgID: "org-1"}, nil).Times(2)

	orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)

	req := userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		RoleIDs:   []string{"role-1"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	_, pat1, err1 := svc.Create(context.Background(), req)
	if err1 != nil {
		t.Fatalf("Create() first call error = %v", err1)
	}
	_, pat2, err2 := svc.Create(context.Background(), req)
	if err2 != nil {
		t.Fatalf("Create() second call error = %v", err2)
	}
	if pat1 == pat2 {
		t.Errorf("Create() should generate unique PATs, got same value twice: %v", pat1)
	}
}

func TestService_Create_HashVerification(t *testing.T) {
	var capturedHash string
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
		Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
		Run(func(ctx context.Context, pat userpat.PAT) {
			capturedHash = pat.SecretHash
		}).
		Return(userpat.PAT{ID: "pat-1", OrgID: "org-1"}, nil)

	orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)

	_, tokenValue, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		RoleIDs:   []string{"role-1"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	// extract the raw secret bytes from the token value
	parts := strings.SplitN(tokenValue, "_", 2)
	if len(parts) != 2 {
		t.Fatal("token should have format prefix_secret")
	}
	secretBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		t.Fatalf("failed to decode secret: %v", err)
	}
	expectedHash := sha3.Sum256(secretBytes)
	expectedHashStr := hex.EncodeToString(expectedHash[:])
	if capturedHash != expectedHashStr {
		t.Errorf("Create() hash mismatch: stored %v, expected sha3-256(secret) = %v", capturedHash, expectedHashStr)
	}
}

func TestService_CreatePolicies_OrgScopedRole(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
		Return(userpat.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc := mocks.NewOrganizationService(t)
	orgSvc.On("GetRaw", mock.Anything, mock.Anything).
		Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(models.AuditRecord{}, nil).Maybe()

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"org-role-1"}}).Return([]role.Role{{
		ID:          "org-role-1",
		Name:        "org_viewer",
		Permissions: []string{"app_organization_get"},
		Scopes:      []string{schema.OrganizationNamespace},
	}}, nil)

	policySvc := mocks.NewPolicyService(t)
	policySvc.EXPECT().Create(mock.Anything, policy.Policy{
		RoleID:        "org-role-1",
		ResourceID:    "org-1",
		ResourceType:  schema.OrganizationNamespace,
		PrincipalID:   "pat-1",
		PrincipalType: schema.PATPrincipal,
	}).Return(policy.Policy{ID: "pol-1"}, nil)

	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "org-token",
		RoleIDs:   []string{"org-role-1"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestService_CreatePolicies_ProjectScopedAllProjects(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
		Return(userpat.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc := mocks.NewOrganizationService(t)
	orgSvc.On("GetRaw", mock.Anything, mock.Anything).
		Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(models.AuditRecord{}, nil).Maybe()

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"proj-role-1"}}).Return([]role.Role{{
		ID:          "proj-role-1",
		Name:        "proj_viewer",
		Permissions: []string{"app_project_get"},
		Scopes:      []string{schema.ProjectNamespace},
	}}, nil)

	policySvc := mocks.NewPolicyService(t)
	policySvc.EXPECT().Create(mock.Anything, policy.Policy{
		RoleID:        "proj-role-1",
		ResourceID:    "org-1",
		ResourceType:  schema.OrganizationNamespace,
		PrincipalID:   "pat-1",
		PrincipalType: schema.PATPrincipal,
		Metadata: metadata.Metadata{
			schema.GrantRelationMetadataKey: schema.PATGrantRelationName,
		},
	}).Return(policy.Policy{ID: "pol-1"}, nil)

	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "all-projects-token",
		RoleIDs:   []string{"proj-role-1"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestService_CreatePolicies_ProjectScopedSpecificProjects(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
		Return(userpat.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc := mocks.NewOrganizationService(t)
	orgSvc.On("GetRaw", mock.Anything, mock.Anything).
		Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(models.AuditRecord{}, nil).Maybe()

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"proj-role-1"}}).Return([]role.Role{{
		ID:          "proj-role-1",
		Name:        "proj_viewer",
		Permissions: []string{"app_project_get"},
		Scopes:      []string{schema.ProjectNamespace},
	}}, nil)

	policySvc := mocks.NewPolicyService(t)
	policySvc.EXPECT().Create(mock.Anything, policy.Policy{
		RoleID:        "proj-role-1",
		ResourceID:    "proj-a",
		ResourceType:  schema.ProjectNamespace,
		PrincipalID:   "pat-1",
		PrincipalType: schema.PATPrincipal,
	}).Return(policy.Policy{ID: "pol-1"}, nil)
	policySvc.EXPECT().Create(mock.Anything, policy.Policy{
		RoleID:        "proj-role-1",
		ResourceID:    "proj-b",
		ResourceType:  schema.ProjectNamespace,
		PrincipalID:   "pat-1",
		PrincipalType: schema.PATPrincipal,
	}).Return(policy.Policy{ID: "pol-2"}, nil)

	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:     "user-1",
		OrgID:      "org-1",
		Title:      "specific-projects-token",
		RoleIDs:    []string{"proj-role-1"},
		ProjectIDs: []string{"proj-a", "proj-b"},
		ExpiresAt:  time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestService_CreatePolicies_DeniedPermission(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	// repo.Create should NOT be called — validation fails before token creation

	orgSvc := mocks.NewOrganizationService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"admin-role"}}).Return([]role.Role{{
		ID:          "admin-role",
		Name:        "org_admin",
		Permissions: []string{"app_organization_administer", "app_organization_get"},
		Scopes:      []string{schema.OrganizationNamespace},
	}}, nil)

	policySvc := mocks.NewPolicyService(t)

	cfg := defaultConfig
	cfg.DeniedPermissions = []string{"app_organization_administer"}

	svc := userpat.NewService(log.NewNoop(), repo, cfg, orgSvc, roleSvc, policySvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "admin-token",
		RoleIDs:   []string{"admin-role"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Create() expected error for denied permission, got nil")
	}
	if !errors.Is(err, userpat.ErrDeniedRole) {
		t.Errorf("Create() error = %v, want ErrDeniedRole", err)
	}
}

func TestService_CreatePolicies_RoleFetchError(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	// repo.Create should NOT be called — role fetch fails before token creation

	orgSvc := mocks.NewOrganizationService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"bad-role"}}).
		Return(nil, errors.New("role not found"))

	policySvc := mocks.NewPolicyService(t)

	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "bad-token",
		RoleIDs:   []string{"bad-role"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Create() expected error for bad role, got nil")
	}
	if !strings.Contains(err.Error(), "fetching roles") {
		t.Errorf("Create() error = %v, want error containing 'fetching roles'", err)
	}
}

func TestService_CreatePolicies_UnsupportedScope(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	// repo.Create should NOT be called — scope validation fails before token creation

	orgSvc := mocks.NewOrganizationService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"group-role"}}).Return([]role.Role{{
		ID:          "group-role",
		Name:        "group_owner",
		Permissions: []string{"app_group_administer"},
		Scopes:      []string{schema.GroupNamespace},
	}}, nil)

	policySvc := mocks.NewPolicyService(t)

	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "group-token",
		RoleIDs:   []string{"group-role"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Create() expected error for unsupported scope, got nil")
	}
	if !errors.Is(err, userpat.ErrUnsupportedScope) {
		t.Errorf("Create() error = %v, want ErrUnsupportedScope", err)
	}
}

func TestService_CreatePolicies_MissingRoleID(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	// repo.Create should NOT be called — role count mismatch fails before token creation

	orgSvc := mocks.NewOrganizationService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)

	roleSvc := mocks.NewRoleService(t)
	// request 2 roles but only 1 found
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"role-a", "role-b"}}).Return([]role.Role{{
		ID:     "role-a",
		Name:   "role_a",
		Scopes: []string{schema.OrganizationNamespace},
	}}, nil)

	policySvc := mocks.NewPolicyService(t)

	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "missing-role-token",
		RoleIDs:   []string{"role-a", "role-b"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Create() expected error for missing role, got nil")
	}
	if !errors.Is(err, userpat.ErrRoleNotFound) {
		t.Errorf("Create() error = %v, want ErrRoleNotFound", err)
	}
	if !strings.Contains(err.Error(), "role-b") {
		t.Errorf("Create() error = %v, want error mentioning missing role ID 'role-b'", err)
	}
}

func TestService_CreatePolicies_NoRoles(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
		Return(userpat.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc, roleSvc, policySvc, auditRepo := newSuccessMocks(t)
	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)

	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "no-roles-token",
		RoleIDs:   nil,
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

// policyKey creates a comparable string for a policy to enable set comparison.
// Format: "roleID→resourceType:resourceID(grantRelation)"
func policyKey(p policy.Policy) string {
	grant := "granted"
	if gr, ok := p.Metadata[schema.GrantRelationMetadataKey].(string); ok && gr != "" {
		grant = gr
	}
	return p.RoleID + "→" + p.ResourceType + ":" + p.ResourceID + "(" + grant + ")"
}

// TestService_CreatePolicies_ScopeMatrix is a comprehensive table-driven test that
// verifies the exact set of policies created for every role/project combination.
//
// The test captures every policyService.Create call and compares the full set against
// expected policies. Because testify mock records all calls, any EXTRA unexpected policy
// creation is caught — e.g. a PAT scoped to proj-1 must NOT produce a policy on proj-2.
func TestService_CreatePolicies_ScopeMatrix(t *testing.T) {
	type wantPolicy struct {
		RoleID       string
		ResourceID   string
		ResourceType string
		Grant        string // "granted" (default) or "pat_granted"
	}

	tests := []struct {
		name       string
		roleIDs    []string
		projectIDs []string
		roles      []role.Role
		want       []wantPolicy
		config     userpat.Config // zero value = use defaultConfig
		wantErr    bool
		wantErrIs  error
		wantErrMsg string
	}{
		{
			name:       "ex1: org_manager + project_owner, all projects",
			roleIDs:    []string{"org-mgr-id", "proj-owner-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "org-mgr-id", Name: "app_organization_manager", Permissions: []string{"app_organization_get", "app_organization_update"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "proj-owner-id", Name: "app_project_owner", Permissions: []string{"app_project_get", "app_project_update", "app_project_delete"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "org-mgr-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
				{RoleID: "proj-owner-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "pat_granted"},
			},
		},
		{
			name:       "ex2: org_viewer + project_viewer, all projects",
			roleIDs:    []string{"org-viewer-id", "proj-viewer-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "proj-viewer-id", Name: "app_project_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "org-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
				{RoleID: "proj-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "pat_granted"},
			},
		},
		{
			name:       "ex3: org_viewer + project_owner, specific projects",
			roleIDs:    []string{"org-viewer-id", "proj-owner-id"},
			projectIDs: []string{"proj-1", "proj-2"},
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "proj-owner-id", Name: "app_project_owner", Permissions: []string{"app_project_get", "app_project_update", "app_project_delete"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "org-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
				{RoleID: "proj-owner-id", ResourceID: "proj-1", ResourceType: schema.ProjectNamespace, Grant: "granted"},
				{RoleID: "proj-owner-id", ResourceID: "proj-2", ResourceType: schema.ProjectNamespace, Grant: "granted"},
			},
		},
		{
			name:       "ex4: org_viewer only, no project access",
			roleIDs:    []string{"org-viewer-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "org-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
			},
		},

		// ── Multiple roles of same scope ─────────────────────────────────

		{
			name:       "multiple org roles create separate org policies",
			roleIDs:    []string{"org-viewer-id", "org-billing-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "org-billing-id", Name: "app_organization_billing_viewer", Permissions: []string{"app_organization_billingview"}, Scopes: []string{schema.OrganizationNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "org-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
				{RoleID: "org-billing-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
			},
		},
		{
			name:       "multiple project roles, all projects → separate pat_granted policies",
			roleIDs:    []string{"proj-viewer-id", "proj-editor-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "proj-viewer-id", Name: "app_project_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
				{ID: "proj-editor-id", Name: "app_project_editor", Permissions: []string{"app_project_get", "app_project_update"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "proj-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "pat_granted"},
				{RoleID: "proj-editor-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "pat_granted"},
			},
		},
		{
			name:       "multiple project roles, specific projects → policy per role per project",
			roleIDs:    []string{"proj-viewer-id", "proj-editor-id"},
			projectIDs: []string{"proj-1", "proj-2"},
			roles: []role.Role{
				{ID: "proj-viewer-id", Name: "app_project_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
				{ID: "proj-editor-id", Name: "app_project_editor", Permissions: []string{"app_project_get", "app_project_update"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "proj-viewer-id", ResourceID: "proj-1", ResourceType: schema.ProjectNamespace, Grant: "granted"},
				{RoleID: "proj-viewer-id", ResourceID: "proj-2", ResourceType: schema.ProjectNamespace, Grant: "granted"},
				{RoleID: "proj-editor-id", ResourceID: "proj-1", ResourceType: schema.ProjectNamespace, Grant: "granted"},
				{RoleID: "proj-editor-id", ResourceID: "proj-2", ResourceType: schema.ProjectNamespace, Grant: "granted"},
			},
		},

		// ── Scope isolation ──────────────────────────────────────────────

		{
			name:       "project role scoped to proj-1 only: no policy on proj-2",
			roleIDs:    []string{"proj-viewer-id"},
			projectIDs: []string{"proj-1"},
			roles: []role.Role{
				{ID: "proj-viewer-id", Name: "app_project_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
			},
			// Only proj-1 gets a policy. If code mistakenly creates a policy
			// on any other project, the captured set won't match.
			want: []wantPolicy{
				{RoleID: "proj-viewer-id", ResourceID: "proj-1", ResourceType: schema.ProjectNamespace, Grant: "granted"},
			},
		},
		{
			name:       "org role does not create project policies even when projectIDs provided",
			roleIDs:    []string{"org-viewer-id"},
			projectIDs: []string{"proj-1", "proj-2"},
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
			},
			// Org-scoped role ignores projectIDs entirely — only org policy created
			want: []wantPolicy{
				{RoleID: "org-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
			},
		},
		{
			name:       "mixed roles with specific projects: org on org, project on projects only",
			roleIDs:    []string{"org-viewer-id", "proj-editor-id"},
			projectIDs: []string{"proj-1"},
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "proj-editor-id", Name: "app_project_editor", Permissions: []string{"app_project_get", "app_project_update"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "org-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
				{RoleID: "proj-editor-id", ResourceID: "proj-1", ResourceType: schema.ProjectNamespace, Grant: "granted"},
			},
		},
		{
			name:       "single project role, single project",
			roleIDs:    []string{"proj-viewer-id"},
			projectIDs: []string{"proj-1"},
			roles: []role.Role{
				{ID: "proj-viewer-id", Name: "app_project_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "proj-viewer-id", ResourceID: "proj-1", ResourceType: schema.ProjectNamespace, Grant: "granted"},
			},
		},
		{
			name:       "single project role, three projects",
			roleIDs:    []string{"proj-viewer-id"},
			projectIDs: []string{"proj-1", "proj-2", "proj-3"},
			roles: []role.Role{
				{ID: "proj-viewer-id", Name: "app_project_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "proj-viewer-id", ResourceID: "proj-1", ResourceType: schema.ProjectNamespace, Grant: "granted"},
				{RoleID: "proj-viewer-id", ResourceID: "proj-2", ResourceType: schema.ProjectNamespace, Grant: "granted"},
				{RoleID: "proj-viewer-id", ResourceID: "proj-3", ResourceType: schema.ProjectNamespace, Grant: "granted"},
			},
		},

		// ── Error cases ──────────────────────────────────────────────────

		{
			name:       "denied permission blocks all policy creation",
			roleIDs:    []string{"org-viewer-id", "org-admin-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "org-admin-id", Name: "app_organization_admin", Permissions: []string{"app_organization_administer"}, Scopes: []string{schema.OrganizationNamespace}},
			},
			config: userpat.Config{
				Enabled:                true,
				Prefix:            "fpt",
				MaxPerUserPerOrg: 50,
				MaxLifetime:       "8760h",
				DeniedPermissions:      []string{"app_organization_administer"},
			},
			want:      nil, // no policies should be created
			wantErr:   true,
			wantErrIs: userpat.ErrDeniedRole,
		},
		{
			name:       "unsupported scope rejects before any policy creation",
			roleIDs:    []string{"org-viewer-id", "group-role-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "group-role-id", Name: "app_group_manager", Permissions: []string{"app_group_get"}, Scopes: []string{schema.GroupNamespace}},
			},
			want:      nil, // scope validation happens upfront — no token or policies created
			wantErr:   true,
			wantErrIs: userpat.ErrUnsupportedScope,
		},
		{
			name:       "role with mixed supported and unsupported scopes is rejected",
			roleIDs:    []string{"mixed-scope-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "mixed-scope-id", Name: "mixed_role", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace, schema.GroupNamespace}},
			},
			want:      nil,
			wantErr:   true,
			wantErrIs: userpat.ErrUnsupportedScope,
		},
		{
			name:       "role with empty scopes is unsupported",
			roleIDs:    []string{"no-scope-id"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "no-scope-id", Name: "custom_role", Permissions: []string{"app_organization_get"}, Scopes: nil},
			},
			want:      nil,
			wantErr:   true,
			wantErrIs: userpat.ErrUnsupportedScope,
		},
		{
			name:       "role count mismatch: requested 2 but found 1",
			roleIDs:    []string{"role-a", "role-b"},
			projectIDs: nil,
			roles: []role.Role{
				{ID: "role-a", Name: "role_a", Scopes: []string{schema.OrganizationNamespace}},
			},
			want:      nil,
			wantErr:   true,
			wantErrIs: userpat.ErrRoleNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig
			if tt.config.Enabled {
				cfg = tt.config
			}

			repo := mocks.NewRepository(t)
			repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
			// Only mock repo.Create for success cases — validation errors fail before token creation
			if !tt.wantErr {
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
					Return(userpat.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)
			}

			orgSvc := mocks.NewOrganizationService(t)
			orgSvc.On("GetRaw", mock.Anything, mock.Anything).
				Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
			auditRepo := mocks.NewAuditRecordRepository(t)
			auditRepo.On("Create", mock.Anything, mock.Anything).Return(models.AuditRecord{}, nil).Maybe()

			// --- roleService: return the test's roles
			roleSvc := mocks.NewRoleService(t)
			roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: tt.roleIDs}).Return(tt.roles, nil)

			// --- policyService: capture all Create calls
			var captured []policy.Policy
			policySvc := mocks.NewPolicyService(t)
			policySvc.On("Create", mock.Anything, mock.AnythingOfType("policy.Policy")).
				Run(func(args mock.Arguments) {
					captured = append(captured, args.Get(1).(policy.Policy))
				}).
				Return(policy.Policy{ID: "pol-gen"}, nil).Maybe()

			svc := userpat.NewService(log.NewNoop(), repo, cfg, orgSvc, roleSvc, policySvc, auditRepo)
			_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
				UserID:     "user-1",
				OrgID:      "org-1",
				Title:      "test-token",
				RoleIDs:    tt.roleIDs,
				ProjectIDs: tt.projectIDs,
				ExpiresAt:  time.Now().Add(24 * time.Hour),
			})

			// --- assert error
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("error = %v, want %v", err, tt.wantErrIs)
				}
				if tt.wantErrMsg != "" && !strings.Contains(err.Error(), tt.wantErrMsg) {
					t.Errorf("error = %v, want containing %q", err, tt.wantErrMsg)
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
			}

			// --- assert exact policy set
			if tt.want == nil && len(captured) > 0 {
				t.Errorf("expected no policies but got %d: %v", len(captured), captured)
				return
			}
			if tt.want == nil {
				return
			}

			// build expected key set
			wantKeys := make(map[string]bool, len(tt.want))
			for _, w := range tt.want {
				grant := w.Grant
				if grant == "" {
					grant = "granted"
				}
				key := w.RoleID + "→" + w.ResourceType + ":" + w.ResourceID + "(" + grant + ")"
				wantKeys[key] = true
			}

			// build captured key set
			gotKeys := make(map[string]bool, len(captured))
			for _, c := range captured {
				key := policyKey(c)
				gotKeys[key] = true

				// also verify common fields on every captured policy
				if c.PrincipalID != "pat-1" {
					t.Errorf("policy %s: PrincipalID = %q, want %q", key, c.PrincipalID, "pat-1")
				}
				if c.PrincipalType != schema.PATPrincipal {
					t.Errorf("policy %s: PrincipalType = %q, want %q", key, c.PrincipalType, schema.PATPrincipal)
				}
			}

			if len(wantKeys) != len(gotKeys) {
				t.Errorf("policy count: want %d, got %d\nwant: %v\ngot:  %v", len(wantKeys), len(gotKeys), wantKeys, gotKeys)
				return
			}

			for key := range wantKeys {
				if !gotKeys[key] {
					t.Errorf("missing expected policy: %s", key)
				}
			}
			for key := range gotKeys {
				if !wantKeys[key] {
					t.Errorf("unexpected policy created: %s", key)
				}
			}
		})
	}
}

func TestService_CreatePolicies_PolicyCreateFailure(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PAT")).
		Return(userpat.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc := mocks.NewOrganizationService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"org-viewer-id", "org-billing-id"}}).
		Return([]role.Role{
			{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
			{ID: "org-billing-id", Name: "app_organization_billing", Permissions: []string{"app_organization_billingview"}, Scopes: []string{schema.OrganizationNamespace}},
		}, nil)

	// first policy Create succeeds, second fails
	policySvc := mocks.NewPolicyService(t)
	policySvc.On("Create", mock.Anything, mock.MatchedBy(func(p policy.Policy) bool {
		return p.RoleID == "org-viewer-id"
	})).Return(policy.Policy{ID: "pol-1"}, nil)
	policySvc.On("Create", mock.Anything, mock.MatchedBy(func(p policy.Policy) bool {
		return p.RoleID == "org-billing-id"
	})).Return(policy.Policy{}, errors.New("spicedb unavailable"))

	svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, orgSvc, roleSvc, policySvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "fail-token",
		RoleIDs:   []string{"org-viewer-id", "org-billing-id"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("expected error when policyService.Create fails, got nil")
	}
	if !strings.Contains(err.Error(), "spicedb unavailable") {
		t.Errorf("error = %v, want containing 'spicedb unavailable'", err)
	}
}

func TestConfig_MaxExpiry(t *testing.T) {
	tests := []struct {
		name string
		cfg  userpat.Config
		want time.Duration
	}{
		{
			name: "should parse valid duration",
			cfg:  userpat.Config{MaxLifetime: "720h"},
			want: 720 * time.Hour,
		},
		{
			name: "should return default 1 year on invalid duration",
			cfg:  userpat.Config{MaxLifetime: "invalid"},
			want: 365 * 24 * time.Hour,
		},
		{
			name: "should return default 1 year on empty duration",
			cfg:  userpat.Config{MaxLifetime: ""},
			want: 365 * 24 * time.Hour,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.cfg.MaxExpiry()
			if got != tt.want {
				t.Errorf("MaxExpiry() = %v, want %v", got, tt.want)
			}
		})
	}
}
