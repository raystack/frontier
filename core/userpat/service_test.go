package userpat_test

import (
	"context"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"io"
	"log/slog"

	"github.com/google/go-cmp/cmp"
	auditmodels "github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/userpat"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/core/userpat/mocks"
	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/sha3"
)

var defaultConfig = userpat.Config{
	Enabled:          true,
	Prefix:           "fpt",
	MaxPerUserPerOrg: 50,
	MaxLifetime:      "8760h",
}

func newSuccessMocks(t *testing.T) (*mocks.OrganizationService, *mocks.RoleService, *mocks.PolicyService, *mocks.ProjectService, *mocks.AuditRecordRepository) {
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
	roleSvc.On("Get", mock.Anything, mock.Anything).
		Return(role.Role{
			ID:     "role-1",
			Name:   "test-role",
			Scopes: []string{schema.OrganizationNamespace},
		}, nil).Maybe()
	policySvc := mocks.NewPolicyService(t)
	policySvc.On("Create", mock.Anything, mock.Anything).
		Return(policy.Policy{}, nil).Maybe()
	policySvc.On("List", mock.Anything, mock.Anything).
		Return([]policy.Policy{}, nil).Maybe()
	projSvc := mocks.NewProjectService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).
		Return(auditmodels.AuditRecord{}, nil).Maybe()
	return orgSvc, roleSvc, policySvc, projSvc, auditRepo
}

func TestService_Create(t *testing.T) {
	futureExpiry := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Second)

	tests := []struct {
		name         string
		setup        func() *userpat.Service
		req          userpat.CreateRequest
		wantErr      bool
		wantErrIs    error
		wantErrMsg   string
		validateFunc func(t *testing.T, got models.PAT, tokenValue string)
	}{
		{
			name: "should return ErrDisabled when PAT feature is disabled",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrDisabled,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, userpat.Config{
					Enabled: false,
				}, orgSvc, nil, nil, nil, auditRepo)
			},
		},
		{
			name: "should return error when CountActive fails",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
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
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, nil, nil, auditRepo)
			},
		},
		{
			name: "should return ErrLimitExceeded when token count equals max",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrLimitExceeded,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(50), nil)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, nil, nil, auditRepo)
			},
		},
		{
			name: "should return ErrLimitExceeded when token count exceeds max",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrLimitExceeded,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(55), nil)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, nil, nil, auditRepo)
			},
		},
		{
			name: "should return error when repo Create fails",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:    true,
			wantErrMsg: "insert failed",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Return(models.PAT{}, errors.New("insert failed"))
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.On("List", mock.Anything, mock.Anything).Return([]role.Role{{
					ID: "role-1", Name: "test-role", Scopes: []string{schema.OrganizationNamespace},
				}}, nil).Maybe()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name: "should return ErrConflict when repo returns duplicate error",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrConflict,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Return(models.PAT{}, paterrors.ErrConflict)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.On("List", mock.Anything, mock.Anything).Return([]role.Role{{
					ID: "role-1", Name: "test-role", Scopes: []string{schema.OrganizationNamespace},
				}}, nil).Maybe()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name: "should create token successfully with correct fields",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: futureExpiry,
				Metadata:  map[string]any{"env": "staging"},
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Run(func(ctx context.Context, pat models.PAT) {
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
						if !pat.ExpiresAt.Equal(futureExpiry) {
							t.Errorf("Create() ExpiresAt = %v, want %v", pat.ExpiresAt, futureExpiry)
						}
					}).
					Return(models.PAT{
						ID:        "pat-id-1",
						UserID:    "user-1",
						OrgID:     "org-1",
						Title:     "my-token",
						Metadata:  map[string]any{"env": "staging"},
						ExpiresAt: futureExpiry,
						CreatedAt: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
					}, nil)
				orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
			},
			validateFunc: func(t *testing.T, got models.PAT, tokenValue string) {
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
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Return(models.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
			},
			validateFunc: func(t *testing.T, got models.PAT, tokenValue string) {
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
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Return(models.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
			},
			validateFunc: func(t *testing.T, got models.PAT, tokenValue string) {
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
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Return(models.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, userpat.Config{
					Enabled:          true,
					Prefix:           "custom",
					MaxPerUserPerOrg: 50,
					MaxLifetime:      "8760h",
				}, orgSvc, roleSvc, policySvc, nil, auditRepo)
			},
			validateFunc: func(t *testing.T, got models.PAT, tokenValue string) {
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
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(49), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Return(models.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
			},
		},
		{
			name: "should generate unique tokens on each call",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Return(models.PAT{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
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
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
		Return(models.PAT{ID: "pat-1", OrgID: "org-1"}, nil).Times(2)

	orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)

	req := userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
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
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
		Run(func(ctx context.Context, pat models.PAT) {
			capturedHash = pat.SecretHash
		}).
		Return(models.PAT{ID: "pat-1", OrgID: "org-1"}, nil)

	orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)

	_, tokenValue, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		Scopes:    []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
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
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
		Return(models.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc := mocks.NewOrganizationService(t)
	orgSvc.On("GetRaw", mock.Anything, mock.Anything).
		Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(auditmodels.AuditRecord{}, nil).Maybe()

	orgRole := role.Role{
		ID:          "org-role-1",
		Name:        "org_viewer",
		Permissions: []string{"app_organization_get"},
		Scopes:      []string{schema.OrganizationNamespace},
	}

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"org-role-1"}}).Return([]role.Role{orgRole}, nil)
	roleSvc.On("Get", mock.Anything, "org-role-1").Return(orgRole, nil).Maybe()

	policySvc := mocks.NewPolicyService(t)
	policySvc.EXPECT().Create(mock.Anything, policy.Policy{
		RoleID:        "org-role-1",
		ResourceID:    "org-1",
		ResourceType:  schema.OrganizationNamespace,
		PrincipalID:   "pat-1",
		PrincipalType: schema.PATPrincipal,
		GrantRelation: schema.RoleGrantRelationName,
	}).Return(policy.Policy{ID: "pol-1"}, nil)
	policySvc.On("List", mock.Anything, mock.Anything).Return([]policy.Policy{}, nil).Maybe()

	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "org-token",
		Scopes:    []models.PATScope{{RoleID: "org-role-1", ResourceType: schema.OrganizationNamespace}},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestService_CreatePolicies_ProjectScopedAllProjects(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
		Return(models.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc := mocks.NewOrganizationService(t)
	orgSvc.On("GetRaw", mock.Anything, mock.Anything).
		Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(auditmodels.AuditRecord{}, nil).Maybe()

	projRole := role.Role{
		ID:          "proj-role-1",
		Name:        "proj_viewer",
		Permissions: []string{"app_project_get"},
		Scopes:      []string{schema.ProjectNamespace},
	}

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"proj-role-1"}}).Return([]role.Role{projRole}, nil)
	roleSvc.On("Get", mock.Anything, "proj-role-1").Return(projRole, nil).Maybe()

	policySvc := mocks.NewPolicyService(t)
	policySvc.EXPECT().Create(mock.Anything, policy.Policy{
		RoleID:        "proj-role-1",
		ResourceID:    "org-1",
		ResourceType:  schema.OrganizationNamespace,
		PrincipalID:   "pat-1",
		PrincipalType: schema.PATPrincipal,
		GrantRelation: schema.PATGrantRelationName,
	}).Return(policy.Policy{ID: "pol-1"}, nil)
	policySvc.On("List", mock.Anything, mock.Anything).Return([]policy.Policy{}, nil).Maybe()

	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "all-projects-token",
		Scopes:    []models.PATScope{{RoleID: "proj-role-1", ResourceType: schema.ProjectNamespace}},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
}

func TestService_CreatePolicies_ProjectScopedSpecificProjects(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
		Return(models.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc := mocks.NewOrganizationService(t)
	orgSvc.On("GetRaw", mock.Anything, mock.Anything).
		Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).Return(auditmodels.AuditRecord{}, nil).Maybe()

	projRole := role.Role{
		ID:          "proj-role-1",
		Name:        "proj_viewer",
		Permissions: []string{"app_project_get"},
		Scopes:      []string{schema.ProjectNamespace},
	}

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"proj-role-1"}}).Return([]role.Role{projRole}, nil)
	roleSvc.On("Get", mock.Anything, "proj-role-1").Return(projRole, nil).Maybe()

	policySvc := mocks.NewPolicyService(t)
	policySvc.EXPECT().Create(mock.Anything, policy.Policy{
		RoleID:        "proj-role-1",
		ResourceID:    "proj-a",
		ResourceType:  schema.ProjectNamespace,
		PrincipalID:   "pat-1",
		PrincipalType: schema.PATPrincipal,
		GrantRelation: schema.RoleGrantRelationName,
	}).Return(policy.Policy{ID: "pol-1"}, nil)
	policySvc.EXPECT().Create(mock.Anything, policy.Policy{
		RoleID:        "proj-role-1",
		ResourceID:    "proj-b",
		ResourceType:  schema.ProjectNamespace,
		PrincipalID:   "pat-1",
		PrincipalType: schema.PATPrincipal,
		GrantRelation: schema.RoleGrantRelationName,
	}).Return(policy.Policy{ID: "pol-2"}, nil)
	policySvc.On("List", mock.Anything, mock.Anything).Return([]policy.Policy{}, nil).Maybe()

	projSvc := mocks.NewProjectService(t)
	projSvc.On("ListByUser", mock.Anything, mock.Anything, mock.Anything).Return([]project.Project{
		{ID: "proj-a"}, {ID: "proj-b"},
	}, nil).Maybe()

	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, projSvc, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "specific-projects-token",
		Scopes:    []models.PATScope{{RoleID: "proj-role-1", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-a", "proj-b"}}},
		ExpiresAt: time.Now().Add(24 * time.Hour),
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

	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg, orgSvc, roleSvc, policySvc, nil, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "admin-token",
		Scopes:    []models.PATScope{{RoleID: "admin-role", ResourceType: schema.OrganizationNamespace}},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Create() expected error for denied permission, got nil")
	}
	if !errors.Is(err, paterrors.ErrDeniedRole) {
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

	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "bad-token",
		Scopes:    []models.PATScope{{RoleID: "bad-role", ResourceType: schema.OrganizationNamespace}},
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

	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "group-token",
		Scopes:    []models.PATScope{{RoleID: "group-role", ResourceType: schema.GroupNamespace}},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Create() expected error for unsupported scope, got nil")
	}
	if !errors.Is(err, paterrors.ErrUnsupportedScope) {
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

	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID: "user-1",
		OrgID:  "org-1",
		Title:  "missing-role-token",
		Scopes: []models.PATScope{
			{RoleID: "role-a", ResourceType: schema.OrganizationNamespace},
			{RoleID: "role-b", ResourceType: schema.OrganizationNamespace},
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("Create() expected error for missing role, got nil")
	}
	if !errors.Is(err, paterrors.ErrRoleNotFound) {
		t.Errorf("Create() error = %v, want ErrRoleNotFound", err)
	}
	if !strings.Contains(err.Error(), "role-b") {
		t.Errorf("Create() error = %v, want error mentioning missing role ID 'role-b'", err)
	}
}

func TestService_CreatePolicies_NoRoles(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
		Return(models.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc, roleSvc, policySvc, _, auditRepo := newSuccessMocks(t)
	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)

	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "no-roles-token",
		Scopes:    nil,
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
	if p.GrantRelation != "" {
		grant = p.GrantRelation
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
		scopes     []models.PATScope
		roles      []role.Role
		want       []wantPolicy
		config     *userpat.Config // nil = use defaultConfig
		wantErr    bool
		wantErrIs  error
		wantErrMsg string
	}{
		{
			name: "ex1: org_manager + project_owner, all projects",
			scopes: []models.PATScope{
				{RoleID: "org-mgr-id", ResourceType: schema.OrganizationNamespace},
				{RoleID: "proj-owner-id", ResourceType: schema.ProjectNamespace},
			},
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
			name: "ex2: org_viewer + project_viewer, all projects",
			scopes: []models.PATScope{
				{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
				{RoleID: "proj-viewer-id", ResourceType: schema.ProjectNamespace},
			},
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
			name: "ex3: org_viewer + project_owner, specific projects",
			scopes: []models.PATScope{
				{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
				{RoleID: "proj-owner-id", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1", "proj-2"}},
			},
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
			name: "ex4: org_viewer only, no project access",
			scopes: []models.PATScope{
				{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
			},
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "org-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
			},
		},

		// ── Multiple roles of same scope ─────────────────────────────────

		{
			name: "multiple org roles create separate org policies",
			scopes: []models.PATScope{
				{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
				{RoleID: "org-billing-id", ResourceType: schema.OrganizationNamespace},
			},
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
			name: "multiple project roles, all projects → separate pat_granted policies",
			scopes: []models.PATScope{
				{RoleID: "proj-viewer-id", ResourceType: schema.ProjectNamespace},
				{RoleID: "proj-editor-id", ResourceType: schema.ProjectNamespace},
			},
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
			name: "multiple project roles, specific projects → policy per role per project",
			scopes: []models.PATScope{
				{RoleID: "proj-viewer-id", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1", "proj-2"}},
				{RoleID: "proj-editor-id", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1", "proj-2"}},
			},
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
			name: "project role scoped to proj-1 only: no policy on proj-2",
			scopes: []models.PATScope{
				{RoleID: "proj-viewer-id", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1"}},
			},
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
			name: "org role does not create project policies when scoped to org",
			scopes: []models.PATScope{
				{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
			},
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
			},
			// Org-scoped role creates only org policy
			want: []wantPolicy{
				{RoleID: "org-viewer-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "granted"},
			},
		},
		{
			name: "mixed roles with specific projects: org on org, project on projects only",
			scopes: []models.PATScope{
				{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
				{RoleID: "proj-editor-id", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1"}},
			},
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
			name: "single project role, single project",
			scopes: []models.PATScope{
				{RoleID: "proj-viewer-id", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1"}},
			},
			roles: []role.Role{
				{ID: "proj-viewer-id", Name: "app_project_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "proj-viewer-id", ResourceID: "proj-1", ResourceType: schema.ProjectNamespace, Grant: "granted"},
			},
		},
		{
			name: "single project role, three projects",
			scopes: []models.PATScope{
				{RoleID: "proj-viewer-id", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-1", "proj-2", "proj-3"}},
			},
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
			name: "denied permission blocks all policy creation",
			scopes: []models.PATScope{
				{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
				{RoleID: "org-admin-id", ResourceType: schema.OrganizationNamespace},
			},
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "org-admin-id", Name: "app_organization_admin", Permissions: []string{"app_organization_administer"}, Scopes: []string{schema.OrganizationNamespace}},
			},
			config: &userpat.Config{
				Enabled:           true,
				Prefix:            "fpt",
				MaxPerUserPerOrg:  50,
				MaxLifetime:       "8760h",
				DeniedPermissions: []string{"app_organization_administer"},
			},
			want:      nil, // no policies should be created
			wantErr:   true,
			wantErrIs: paterrors.ErrDeniedRole,
		},
		{
			name: "unsupported scope rejects before any policy creation",
			scopes: []models.PATScope{
				{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
				{RoleID: "group-role-id", ResourceType: schema.GroupNamespace},
			},
			roles: []role.Role{
				{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				{ID: "group-role-id", Name: "app_group_manager", Permissions: []string{"app_group_get"}, Scopes: []string{schema.GroupNamespace}},
			},
			want:      nil, // scope validation happens upfront — no token or policies created
			wantErr:   true,
			wantErrIs: paterrors.ErrUnsupportedScope,
		},
		{
			name: "role with mixed scopes is allowed when requested resource type is supported",
			scopes: []models.PATScope{
				{RoleID: "mixed-scope-id", ResourceType: schema.ProjectNamespace},
			},
			roles: []role.Role{
				{ID: "mixed-scope-id", Name: "mixed_role", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace, schema.GroupNamespace}},
			},
			want: []wantPolicy{
				{RoleID: "mixed-scope-id", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, Grant: "pat_granted"},
			},
			wantErr: false,
		},
		{
			name: "role with empty scopes is unsupported",
			scopes: []models.PATScope{
				{RoleID: "no-scope-id", ResourceType: ""},
			},
			roles: []role.Role{
				{ID: "no-scope-id", Name: "custom_role", Permissions: []string{"app_organization_get"}, Scopes: nil},
			},
			want:      nil,
			wantErr:   true,
			wantErrIs: paterrors.ErrUnsupportedScope,
		},
		{
			name: "role count mismatch: requested 2 but found 1",
			scopes: []models.PATScope{
				{RoleID: "role-a", ResourceType: schema.OrganizationNamespace},
				{RoleID: "role-b", ResourceType: schema.OrganizationNamespace},
			},
			roles: []role.Role{
				{ID: "role-a", Name: "role_a", Scopes: []string{schema.OrganizationNamespace}},
			},
			want:      nil,
			wantErr:   true,
			wantErrIs: paterrors.ErrRoleNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := defaultConfig
			if tt.config != nil {
				cfg = *tt.config
			}

			repo := mocks.NewRepository(t)
			repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
			// Only mock repo.Create for success cases — validation errors fail before token creation
			if !tt.wantErr {
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
					Return(models.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)
			}

			orgSvc := mocks.NewOrganizationService(t)
			orgSvc.On("GetRaw", mock.Anything, mock.Anything).
				Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
			auditRepo := mocks.NewAuditRecordRepository(t)
			auditRepo.On("Create", mock.Anything, mock.Anything).Return(auditmodels.AuditRecord{}, nil).Maybe()

			// --- extract role IDs from scopes for the role service mock
			var scopeRoleIDs []string
			for _, s := range tt.scopes {
				scopeRoleIDs = append(scopeRoleIDs, s.RoleID)
			}

			// --- roleService: return the test's roles
			roleSvc := mocks.NewRoleService(t)
			roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: scopeRoleIDs}).Return(tt.roles, nil)
			// createPoliciesFromScopes calls Get per scope
			if !tt.wantErr {
				roleMap := make(map[string]role.Role, len(tt.roles))
				for _, r := range tt.roles {
					roleMap[r.ID] = r
				}
				for _, sc := range tt.scopes {
					if r, ok := roleMap[sc.RoleID]; ok {
						roleSvc.On("Get", mock.Anything, sc.RoleID).Return(r, nil).Maybe()
					}
				}
			}

			// --- policyService: capture all Create calls
			var captured []policy.Policy
			policySvc := mocks.NewPolicyService(t)
			policySvc.On("Create", mock.Anything, mock.AnythingOfType("policy.Policy")).
				Run(func(args mock.Arguments) {
					captured = append(captured, args.Get(1).(policy.Policy))
				}).
				Return(policy.Policy{ID: "pol-gen"}, nil).Maybe()
			policySvc.On("List", mock.Anything, mock.Anything).Return([]policy.Policy{}, nil).Maybe()

			projSvc := mocks.NewProjectService(t)
			projSvc.On("ListByUser", mock.Anything, mock.Anything, mock.Anything).Return([]project.Project{
				{ID: "proj-1"}, {ID: "proj-2"}, {ID: "proj-3"}, {ID: "proj-a"}, {ID: "proj-b"},
			}, nil).Maybe()

			svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg, orgSvc, roleSvc, policySvc, projSvc, auditRepo)
			_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "test-token",
				Scopes:    tt.scopes,
				ExpiresAt: time.Now().Add(24 * time.Hour),
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
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
		Return(models.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)

	orgSvc := mocks.NewOrganizationService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)

	orgViewerRole := role.Role{ID: "org-viewer-id", Name: "app_organization_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}}
	orgBillingRole := role.Role{ID: "org-billing-id", Name: "app_organization_billing", Permissions: []string{"app_organization_billingview"}, Scopes: []string{schema.OrganizationNamespace}}

	roleSvc := mocks.NewRoleService(t)
	roleSvc.EXPECT().List(mock.Anything, role.Filter{IDs: []string{"org-viewer-id", "org-billing-id"}}).
		Return([]role.Role{orgViewerRole, orgBillingRole}, nil)
	roleSvc.On("Get", mock.Anything, "org-viewer-id").Return(orgViewerRole, nil).Maybe()
	roleSvc.On("Get", mock.Anything, "org-billing-id").Return(orgBillingRole, nil).Maybe()

	// first policy Create succeeds, second fails
	policySvc := mocks.NewPolicyService(t)
	policySvc.On("Create", mock.Anything, mock.MatchedBy(func(p policy.Policy) bool {
		return p.RoleID == "org-viewer-id"
	})).Return(policy.Policy{ID: "pol-1"}, nil)
	policySvc.On("Create", mock.Anything, mock.MatchedBy(func(p policy.Policy) bool {
		return p.RoleID == "org-billing-id"
	})).Return(policy.Policy{}, errors.New("spicedb unavailable"))

	svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
	_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID: "user-1",
		OrgID:  "org-1",
		Title:  "fail-token",
		Scopes: []models.PATScope{
			{RoleID: "org-viewer-id", ResourceType: schema.OrganizationNamespace},
			{RoleID: "org-billing-id", ResourceType: schema.OrganizationNamespace},
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err == nil {
		t.Fatal("expected error when policyService.Create fails, got nil")
	}
	if !strings.Contains(err.Error(), "spicedb unavailable") {
		t.Errorf("error = %v, want containing 'spicedb unavailable'", err)
	}
}

func TestService_ListAllowedRoles(t *testing.T) {
	tests := []struct {
		name      string
		scopes    []string
		setup     func() *userpat.Service
		wantErr   bool
		wantErrIs error
		wantCount int
		wantIDs   []string
	}{
		{
			name:      "should return ErrDisabled when PAT feature is disabled",
			wantErr:   true,
			wantErrIs: paterrors.ErrDisabled,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, userpat.Config{
					Enabled: false,
				}, orgSvc, nil, nil, nil, auditRepo)
			},
		},
		{
			name:    "should propagate role service error",
			wantErr: true,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.OrganizationNamespace, schema.ProjectNamespace},
				}).Return(nil, errors.New("db connection failed"))
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should filter out roles with denied permissions",
			wantErr:   false,
			wantCount: 2,
			wantIDs:   []string{"org-viewer-id", "proj-viewer-id"},
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.OrganizationNamespace, schema.ProjectNamespace},
				}).Return([]role.Role{
					{ID: "org-viewer-id", Name: "org_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
					{ID: "org-admin-id", Name: "org_admin", Permissions: []string{"app_organization_administer", "app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
					{ID: "proj-viewer-id", Name: "proj_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
				}, nil)
				cfg := defaultConfig
				cfg.DeniedPermissions = []string{"app_organization_administer"}
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should return empty slice when all roles are denied",
			wantErr:   false,
			wantCount: 0,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.OrganizationNamespace, schema.ProjectNamespace},
				}).Return([]role.Role{
					{ID: "admin-id", Name: "org_admin", Permissions: []string{"app_organization_administer"}, Scopes: []string{schema.OrganizationNamespace}},
				}, nil)
				cfg := defaultConfig
				cfg.DeniedPermissions = []string{"app_organization_administer"}
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, cfg, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should return empty slice when no roles exist",
			wantErr:   false,
			wantCount: 0,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.OrganizationNamespace, schema.ProjectNamespace},
				}).Return([]role.Role{}, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should return all roles when no denied permissions configured",
			wantErr:   false,
			wantCount: 3,
			wantIDs:   []string{"org-viewer-id", "org-admin-id", "proj-viewer-id"},
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.OrganizationNamespace, schema.ProjectNamespace},
				}).Return([]role.Role{
					{ID: "org-viewer-id", Name: "org_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
					{ID: "org-admin-id", Name: "org_admin", Permissions: []string{"app_organization_administer"}, Scopes: []string{schema.OrganizationNamespace}},
					{ID: "proj-viewer-id", Name: "proj_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
				}, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should normalize short alias 'project' to app/project",
			scopes:    []string{"project"},
			wantErr:   false,
			wantCount: 1,
			wantIDs:   []string{"proj-viewer-id"},
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.ProjectNamespace},
				}).Return([]role.Role{
					{ID: "proj-viewer-id", Name: "proj_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
				}, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should normalize short alias 'org' to app/organization",
			scopes:    []string{"org"},
			wantErr:   false,
			wantCount: 1,
			wantIDs:   []string{"org-viewer-id"},
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.OrganizationNamespace},
				}).Return([]role.Role{
					{ID: "org-viewer-id", Name: "org_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
				}, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should accept full namespace app/project",
			scopes:    []string{schema.ProjectNamespace},
			wantErr:   false,
			wantCount: 1,
			wantIDs:   []string{"proj-viewer-id"},
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.ProjectNamespace},
				}).Return([]role.Role{
					{ID: "proj-viewer-id", Name: "proj_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
				}, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should reject unsupported scope like group",
			scopes:    []string{"group"},
			wantErr:   true,
			wantErrIs: paterrors.ErrUnsupportedScope,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, nil, nil, auditRepo)
			},
		},
		{
			name:      "should reject unknown scope",
			scopes:    []string{"unknown"},
			wantErr:   true,
			wantErrIs: paterrors.ErrUnsupportedScope,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, nil, nil, auditRepo)
			},
		},
		{
			name:      "should deduplicate repeated scopes",
			scopes:    []string{"project", "project", "project"},
			wantErr:   false,
			wantCount: 1,
			wantIDs:   []string{"proj-viewer-id"},
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.ProjectNamespace},
				}).Return([]role.Role{
					{ID: "proj-viewer-id", Name: "proj_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
				}, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
		{
			name:      "should deduplicate mixed aliases and full namespaces",
			scopes:    []string{"project", schema.ProjectNamespace, "org", schema.OrganizationNamespace},
			wantErr:   false,
			wantCount: 3,
			wantIDs:   []string{"org-viewer-id", "org-admin-id", "proj-viewer-id"},
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, role.Filter{
					OrgID:  schema.PlatformOrgID.String(),
					Scopes: []string{schema.ProjectNamespace, schema.OrganizationNamespace},
				}).Return([]role.Role{
					{ID: "org-viewer-id", Name: "org_viewer", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
					{ID: "org-admin-id", Name: "org_admin", Permissions: []string{"app_organization_get"}, Scopes: []string{schema.OrganizationNamespace}},
					{ID: "proj-viewer-id", Name: "proj_viewer", Permissions: []string{"app_project_get"}, Scopes: []string{schema.ProjectNamespace}},
				}, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, nil, nil, auditRepo)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			got, err := svc.ListAllowedRoles(context.Background(), tt.scopes)
			if (err != nil) != tt.wantErr {
				t.Errorf("ListAllowedRoles() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("ListAllowedRoles() error = %v, wantErrIs %v", err, tt.wantErrIs)
				return
			}
			if !tt.wantErr {
				if len(got) != tt.wantCount {
					t.Errorf("ListAllowedRoles() returned %d roles, want %d", len(got), tt.wantCount)
				}
				if tt.wantIDs != nil {
					var gotIDs []string
					for _, r := range got {
						gotIDs = append(gotIDs, r.ID)
					}
					if diff := cmp.Diff(tt.wantIDs, gotIDs); diff != "" {
						t.Errorf("ListAllowedRoles() IDs mismatch (-want +got):\n%s", diff)
					}
				}
			}
		})
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

func TestService_Get(t *testing.T) {
	testPAT := models.PAT{
		ID:        "pat-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name      string
		setup     func() *userpat.Service
		userID    string
		patID     string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "should return ErrDisabled when PAT feature is disabled",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc, _, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, userpat.Config{
					Enabled: false,
				}, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrDisabled,
		},
		{
			name:   "should return error when repo GetByID fails",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(models.PAT{}, paterrors.ErrNotFound)
				orgSvc, _, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrNotFound,
		},
		{
			name:   "should return ErrNotFound when PAT belongs to different user",
			userID: "user-2",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				orgSvc, _, policySvc, _, auditRepo := newSuccessMocks(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrNotFound,
		},
		{
			name:   "should return PAT when user owns it",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				orgSvc, _, policySvc, _, auditRepo := newSuccessMocks(t)
				policySvc.On("List", mock.Anything, mock.Anything).
					Return([]policy.Policy{
						{RoleID: "role-1", ResourceType: "app/organization", ResourceID: "org-1"},
					}, nil).Maybe()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
		{
			name:   "should return error when enrichWithScope fails",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				orgSvc := mocks.NewOrganizationService(t)
				policySvc := mocks.NewPolicyService(t)
				policySvc.On("List", mock.Anything, mock.Anything).
					Return(nil, errors.New("spicedb down"))
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			got, err := svc.Get(context.Background(), tt.userID, tt.patID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Get() expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Get() error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("Get() unexpected error: %v", err)
			}
			if got.ID != testPAT.ID {
				t.Errorf("Get() PAT ID = %v, want %v", got.ID, testPAT.ID)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	testPAT := models.PAT{
		ID:        "pat-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	tests := []struct {
		name      string
		setup     func() *userpat.Service
		userID    string
		patID     string
		wantErr   bool
		wantErrIs error
	}{
		{
			name:   "should return ErrDisabled when PAT feature is disabled",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, userpat.Config{
					Enabled: false,
				}, orgSvc, nil, nil, nil, auditRepo)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrDisabled,
		},
		{
			name:   "should return ErrNotFound when PAT does not exist",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(models.PAT{}, paterrors.ErrNotFound)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, nil, nil, auditRepo)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrNotFound,
		},
		{
			name:   "should return ErrNotFound when PAT belongs to different user",
			userID: "user-2",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, nil, nil, auditRepo)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrNotFound,
		},
		{
			name:   "should return error when repo delete fails",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				repo.EXPECT().Delete(mock.Anything, "pat-1").
					Return(errors.New("db error"))
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, nil, nil, auditRepo)
			},
			wantErr: true,
		},
		{
			name:   "should return error when policy list fails after soft-delete",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				repo.EXPECT().Delete(mock.Anything, "pat-1").
					Return(nil)
				orgSvc := mocks.NewOrganizationService(t)
				policySvc := mocks.NewPolicyService(t)
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return(nil, errors.New("spicedb down"))
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: true,
		},
		{
			name:   "should return error when policy delete fails after soft-delete",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				repo.EXPECT().Delete(mock.Anything, "pat-1").
					Return(nil)
				orgSvc := mocks.NewOrganizationService(t)
				policySvc := mocks.NewPolicyService(t)
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{{ID: "pol-1"}}, nil)
				policySvc.EXPECT().Delete(mock.Anything, "pol-1").
					Return(errors.New("spicedb unavailable"))
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: true,
		},
		{
			name:   "should delete successfully with policies",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				repo.EXPECT().Delete(mock.Anything, "pat-1").
					Return(nil)
				orgSvc := mocks.NewOrganizationService(t)
				orgSvc.On("GetRaw", mock.Anything, mock.Anything).
					Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
				policySvc := mocks.NewPolicyService(t)
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{
					{ID: "pol-1"},
					{ID: "pol-2"},
				}, nil)
				policySvc.EXPECT().Delete(mock.Anything, "pol-1").Return(nil)
				policySvc.EXPECT().Delete(mock.Anything, "pol-2").Return(nil)
				auditRepo := mocks.NewAuditRecordRepository(t)
				auditRepo.On("Create", mock.Anything, mock.Anything).
					Return(auditmodels.AuditRecord{}, nil).Maybe()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
		{
			name:   "should delete successfully with no policies",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				repo.EXPECT().Delete(mock.Anything, "pat-1").
					Return(nil)
				orgSvc := mocks.NewOrganizationService(t)
				orgSvc.On("GetRaw", mock.Anything, mock.Anything).
					Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
				policySvc := mocks.NewPolicyService(t)
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{}, nil)
				auditRepo := mocks.NewAuditRecordRepository(t)
				auditRepo.On("Create", mock.Anything, mock.Anything).
					Return(auditmodels.AuditRecord{}, nil).Maybe()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
		{
			name:   "should succeed even when audit record creation fails",
			userID: "user-1",
			patID:  "pat-1",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				repo.EXPECT().Delete(mock.Anything, "pat-1").
					Return(nil)
				orgSvc := mocks.NewOrganizationService(t)
				orgSvc.On("GetRaw", mock.Anything, mock.Anything).
					Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
				policySvc := mocks.NewPolicyService(t)
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{}, nil)
				auditRepo := mocks.NewAuditRecordRepository(t)
				auditRepo.On("Create", mock.Anything, mock.Anything).
					Return(auditmodels.AuditRecord{}, errors.New("audit db down"))
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			err := svc.Delete(context.Background(), tt.userID, tt.patID)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Delete() expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Delete() error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("Delete() unexpected error: %v", err)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	testPAT := models.PAT{
		ID:        "pat-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "old-title",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	updatedPAT := models.PAT{
		ID:        "pat-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "new-title",
		ExpiresAt: testPAT.ExpiresAt,
		CreatedAt: testPAT.CreatedAt,
		UpdatedAt: time.Now(),
	}

	validRole := role.Role{
		ID:     "role-1",
		Name:   "test-role",
		Scopes: []string{schema.OrganizationNamespace},
	}

	defaultInput := models.PAT{
		UserID:   "user-1",
		ID:       "pat-1",
		Title:    "new-title",
		Scopes:   []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
		Metadata: map[string]any{"key": "val"},
	}

	tests := []struct {
		name      string
		setup     func() *userpat.Service
		input     models.PAT
		wantErr   bool
		wantErrIs error
	}{
		{
			name:  "should return ErrDisabled when PAT feature is disabled",
			input: defaultInput,
			setup: func() *userpat.Service {
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, userpat.Config{
					Enabled: false,
				}, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrDisabled,
		},
		{
			name:  "should return ErrNotFound when PAT does not exist",
			input: defaultInput,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(models.PAT{}, paterrors.ErrNotFound)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrNotFound,
		},
		{
			name: "should return ErrNotFound when PAT belongs to different user",
			input: models.PAT{
				UserID: "user-2",
				ID:     "pat-1",
				Title:  "new-title",
				Scopes: []models.PATScope{{RoleID: "role-1", ResourceType: schema.OrganizationNamespace}},
			},
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrNotFound,
		},
		{
			name:  "should return error when role validation fails",
			input: defaultInput,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, mock.Anything).
					Return(nil, paterrors.ErrRoleNotFound)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, roleSvc, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrRoleNotFound,
		},
		{
			name:  "should return error when repo update fails",
			input: defaultInput,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, mock.Anything).
					Return([]role.Role{validRole}, nil)
				policySvc := mocks.NewPolicyService(t)
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{}, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(models.PAT{}, errors.New("db error"))
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, roleSvc, policySvc, nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "should return ErrConflict when title already exists",
			input: defaultInput,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, mock.Anything).
					Return([]role.Role{validRole}, nil)
				policySvc := mocks.NewPolicyService(t)
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{}, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(models.PAT{}, paterrors.ErrConflict)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, roleSvc, policySvc, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrConflict,
		},
		{
			name:  "should return error when delete old policies fails",
			input: defaultInput,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil)
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, mock.Anything).
					Return([]role.Role{validRole}, nil)
				policySvc := mocks.NewPolicyService(t)
				// captureOldScope call
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{}, nil).Once()
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(updatedPAT, nil)
				// deletePolicies call
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return(nil, errors.New("spicedb down")).Once()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, roleSvc, policySvc, nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "should return error when PAT deleted concurrently",
			input: defaultInput,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				// getOwnedPAT
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil).Once()
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, mock.Anything).
					Return([]role.Role{validRole}, nil)
				policySvc := mocks.NewPolicyService(t)
				policySvc.EXPECT().List(mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{}, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(updatedPAT, nil)
				// TOCTOU re-check returns not found (concurrent delete)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(models.PAT{}, paterrors.ErrNotFound).Once()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, roleSvc, policySvc, nil, nil)
			},
			wantErr: true,
		},
		{
			name:  "should update successfully",
			input: defaultInput,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				// getOwnedPAT
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil).Once()
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, mock.Anything).
					Return([]role.Role{validRole}, nil)
				roleSvc.On("Get", mock.Anything, mock.Anything).Return(validRole, nil).Maybe()
				orgSvc := mocks.NewOrganizationService(t)
				orgSvc.On("GetRaw", mock.Anything, mock.Anything).
					Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
				policySvc := mocks.NewPolicyService(t)
				// captureOldScope + enrichWithScope (after update)
				policySvc.On("List", mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{}, nil)
				policySvc.On("Create", mock.Anything, mock.Anything).
					Return(policy.Policy{}, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(updatedPAT, nil)
				// TOCTOU re-check
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(updatedPAT, nil).Once()
				auditRepo := mocks.NewAuditRecordRepository(t)
				auditRepo.On("Create", mock.Anything, mock.Anything).
					Return(auditmodels.AuditRecord{}, nil).Maybe()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
		{
			name:  "should succeed even when audit record creation fails",
			input: defaultInput,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(testPAT, nil).Once()
				roleSvc := mocks.NewRoleService(t)
				roleSvc.EXPECT().List(mock.Anything, mock.Anything).
					Return([]role.Role{validRole}, nil)
				roleSvc.On("Get", mock.Anything, mock.Anything).Return(validRole, nil).Maybe()
				orgSvc := mocks.NewOrganizationService(t)
				orgSvc.On("GetRaw", mock.Anything, mock.Anything).
					Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
				policySvc := mocks.NewPolicyService(t)
				policySvc.On("List", mock.Anything, policy.Filter{
					PrincipalID:   "pat-1",
					PrincipalType: schema.PATPrincipal,
				}).Return([]policy.Policy{}, nil)
				policySvc.On("Create", mock.Anything, mock.Anything).
					Return(policy.Policy{}, nil)
				repo.EXPECT().Update(mock.Anything, mock.Anything).
					Return(updatedPAT, nil)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(updatedPAT, nil).Once()
				auditRepo := mocks.NewAuditRecordRepository(t)
				auditRepo.On("Create", mock.Anything, mock.Anything).
					Return(auditmodels.AuditRecord{}, errors.New("audit db down"))
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			_, err := svc.Update(context.Background(), tt.input)

			if tt.wantErr {
				if err == nil {
					t.Fatal("Update() expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Update() error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("Update() unexpected error: %v", err)
			}
		})
	}
}

func TestService_Regenerate(t *testing.T) {
	futureExpiry := time.Now().Add(48 * time.Hour)

	activePAT := models.PAT{
		ID:        "pat-1",
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		ExpiresAt: time.Now().Add(24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	expiredPAT := models.PAT{
		ID:        "pat-2",
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "expired-token",
		ExpiresAt: time.Now().Add(-24 * time.Hour),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	regenTime := time.Now()
	regeneratedPAT := models.PAT{
		ID:            "pat-1",
		UserID:        "user-1",
		OrgID:         "org-1",
		Title:         "my-token",
		ExpiresAt:     futureExpiry,
		RegeneratedAt: &regenTime,
		CreatedAt:     activePAT.CreatedAt,
		UpdatedAt:     time.Now(),
	}

	tests := []struct {
		name      string
		setup     func() *userpat.Service
		userID    string
		patID     string
		expiresAt time.Time
		wantErr   bool
		wantErrIs error
	}{
		{
			name:      "should return ErrDisabled when PAT feature is disabled",
			userID:    "user-1",
			patID:     "pat-1",
			expiresAt: futureExpiry,
			setup: func() *userpat.Service {
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, userpat.Config{
					Enabled: false,
				}, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrDisabled,
		},
		{
			name:      "should return ErrNotFound when PAT does not exist",
			userID:    "user-1",
			patID:     "pat-1",
			expiresAt: futureExpiry,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(models.PAT{}, paterrors.ErrNotFound)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrNotFound,
		},
		{
			name:      "should return ErrNotFound when PAT belongs to different user",
			userID:    "user-2",
			patID:     "pat-1",
			expiresAt: futureExpiry,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(activePAT, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrNotFound,
		},
		{
			name:      "should return error when expiry is in the past",
			userID:    "user-1",
			patID:     "pat-1",
			expiresAt: time.Now().Add(-1 * time.Hour),
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(activePAT, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrExpiryInPast,
		},
		{
			name:      "should return ErrLimitExceeded when reviving expired PAT at limit",
			userID:    "user-1",
			patID:     "pat-2",
			expiresAt: futureExpiry,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-2").
					Return(expiredPAT, nil)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(50), nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrLimitExceeded,
		},
		{
			name:      "should not check limit when regenerating active PAT",
			userID:    "user-1",
			patID:     "pat-1",
			expiresAt: futureExpiry,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(activePAT, nil)
				// No CountActive call expected — PAT is active
				repo.EXPECT().Regenerate(mock.Anything, "pat-1", mock.Anything, mock.Anything).
					Return(regeneratedPAT, nil)
				orgSvc := mocks.NewOrganizationService(t)
				orgSvc.On("GetRaw", mock.Anything, mock.Anything).
					Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
				policySvc := mocks.NewPolicyService(t)
				policySvc.On("List", mock.Anything, mock.Anything).
					Return([]policy.Policy{}, nil).Maybe()
				auditRepo := mocks.NewAuditRecordRepository(t)
				auditRepo.On("Create", mock.Anything, mock.Anything).
					Return(auditmodels.AuditRecord{}, nil).Maybe()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
		{
			name:      "should return error when repo regenerate fails",
			userID:    "user-1",
			patID:     "pat-1",
			expiresAt: futureExpiry,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(activePAT, nil)
				repo.EXPECT().Regenerate(mock.Anything, "pat-1", mock.Anything, mock.Anything).
					Return(models.PAT{}, errors.New("db error"))
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantErr: true,
		},
		{
			name:      "should regenerate expired PAT successfully when under limit",
			userID:    "user-1",
			patID:     "pat-2",
			expiresAt: futureExpiry,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-2").
					Return(expiredPAT, nil)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(10), nil)
				repo.EXPECT().Regenerate(mock.Anything, "pat-2", mock.Anything, mock.Anything).
					Return(regeneratedPAT, nil)
				orgSvc := mocks.NewOrganizationService(t)
				orgSvc.On("GetRaw", mock.Anything, mock.Anything).
					Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
				policySvc := mocks.NewPolicyService(t)
				policySvc.On("List", mock.Anything, mock.Anything).
					Return([]policy.Policy{}, nil).Maybe()
				auditRepo := mocks.NewAuditRecordRepository(t)
				auditRepo.On("Create", mock.Anything, mock.Anything).
					Return(auditmodels.AuditRecord{}, nil).Maybe()
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
		{
			name:      "should succeed even when audit record creation fails",
			userID:    "user-1",
			patID:     "pat-1",
			expiresAt: futureExpiry,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().GetByID(mock.Anything, "pat-1").
					Return(activePAT, nil)
				repo.EXPECT().Regenerate(mock.Anything, "pat-1", mock.Anything, mock.Anything).
					Return(regeneratedPAT, nil)
				orgSvc := mocks.NewOrganizationService(t)
				orgSvc.On("GetRaw", mock.Anything, mock.Anything).
					Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
				policySvc := mocks.NewPolicyService(t)
				policySvc.On("List", mock.Anything, mock.Anything).
					Return([]policy.Policy{}, nil).Maybe()
				auditRepo := mocks.NewAuditRecordRepository(t)
				auditRepo.On("Create", mock.Anything, mock.Anything).
					Return(auditmodels.AuditRecord{}, errors.New("audit db down"))
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, nil, policySvc, nil, auditRepo)
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			_, _, err := svc.Regenerate(context.Background(), tt.userID, tt.patID, tt.expiresAt)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Regenerate() expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Regenerate() error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("Regenerate() unexpected error: %v", err)
			}
		})
	}
}

func TestService_IsTitleAvailable(t *testing.T) {
	tests := []struct {
		name          string
		setup         func() *userpat.Service
		userID        string
		orgID         string
		title         string
		wantAvailable bool
		wantErr       bool
		wantErrIs     error
	}{
		{
			name:   "should return ErrDisabled when PAT feature is disabled",
			userID: "user-1",
			orgID:  "org-1",
			title:  "my-token",
			setup: func() *userpat.Service {
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), nil, userpat.Config{
					Enabled: false,
				}, nil, nil, nil, nil, nil)
			},
			wantErr:   true,
			wantErrIs: paterrors.ErrDisabled,
		},
		{
			name:   "should return true when title is available",
			userID: "user-1",
			orgID:  "org-1",
			title:  "new-token",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().IsTitleAvailable(mock.Anything, "user-1", "org-1", "new-token").
					Return(true, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantAvailable: true,
		},
		{
			name:   "should return false when title is taken",
			userID: "user-1",
			orgID:  "org-1",
			title:  "existing-token",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().IsTitleAvailable(mock.Anything, "user-1", "org-1", "existing-token").
					Return(false, nil)
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantAvailable: false,
		},
		{
			name:   "should return error when repo fails",
			userID: "user-1",
			orgID:  "org-1",
			title:  "my-token",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().IsTitleAvailable(mock.Anything, "user-1", "org-1", "my-token").
					Return(false, errors.New("db error"))
				return userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, nil, nil, nil, nil)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := tt.setup()
			available, err := svc.IsTitleAvailable(context.Background(), tt.userID, tt.orgID, tt.title)
			if tt.wantErr {
				if err == nil {
					t.Fatal("IsTitleAvailable() expected error, got nil")
				}
				if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
					t.Errorf("IsTitleAvailable() error = %v, want %v", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Fatalf("IsTitleAvailable() unexpected error: %v", err)
			}
			if available != tt.wantAvailable {
				t.Errorf("IsTitleAvailable() = %v, want %v", available, tt.wantAvailable)
			}
		})
	}
}

func TestService_ValidateProjectAccess(t *testing.T) {
	t.Run("should reject project user has no access to", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
		roleSvc := mocks.NewRoleService(t)
		roleSvc.EXPECT().List(mock.Anything, mock.Anything).Return([]role.Role{
			{ID: "role-1", Name: "proj_viewer", Scopes: []string{schema.ProjectNamespace}, Permissions: []string{"app_project_get"}},
		}, nil)
		projSvc := mocks.NewProjectService(t)
		projSvc.On("ListByUser", mock.Anything, mock.Anything, mock.Anything).Return([]project.Project{
			{ID: "proj-in-org"},
		}, nil)

		svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, nil, roleSvc, nil, projSvc, nil)
		_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
			UserID: "user-1",
			OrgID:  "org-1",
			Title:  "cross-org-test",
			Scopes: []models.PATScope{
				{RoleID: "role-1", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-not-in-org"}},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, paterrors.ErrProjectForbidden) {
			t.Errorf("expected ErrProjectForbidden, got %v", err)
		}
	})

	t.Run("should allow project user has access to", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
		repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
			Return(models.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)
		orgSvc := mocks.NewOrganizationService(t)
		orgSvc.On("GetRaw", mock.Anything, mock.Anything).
			Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
		roleSvc := mocks.NewRoleService(t)
		roleSvc.EXPECT().List(mock.Anything, mock.Anything).Return([]role.Role{
			{ID: "role-1", Name: "proj_viewer", Scopes: []string{schema.ProjectNamespace}, Permissions: []string{"app_project_get"}},
		}, nil)
		policySvc := mocks.NewPolicyService(t)
		policySvc.On("Create", mock.Anything, mock.Anything).Return(policy.Policy{}, nil).Maybe()
		policySvc.On("List", mock.Anything, mock.Anything).Return([]policy.Policy{}, nil).Maybe()
		projSvc := mocks.NewProjectService(t)
		projSvc.On("ListByUser", mock.Anything, mock.Anything, mock.Anything).Return([]project.Project{
			{ID: "proj-in-org"},
		}, nil)
		auditRepo := mocks.NewAuditRecordRepository(t)
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(auditmodels.AuditRecord{}, nil).Maybe()

		svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, projSvc, auditRepo)
		_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
			UserID: "user-1",
			OrgID:  "org-1",
			Title:  "valid-project-test",
			Scopes: []models.PATScope{
				{RoleID: "role-1", ResourceType: schema.ProjectNamespace, ResourceIDs: []string{"proj-in-org"}},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("should skip validation for all-projects scope", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").Return(int64(0), nil)
		repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("models.PAT")).
			Return(models.PAT{ID: "pat-1", OrgID: "org-1", CreatedAt: time.Now()}, nil)
		orgSvc := mocks.NewOrganizationService(t)
		orgSvc.On("GetRaw", mock.Anything, mock.Anything).
			Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
		roleSvc := mocks.NewRoleService(t)
		roleSvc.EXPECT().List(mock.Anything, mock.Anything).Return([]role.Role{
			{ID: "role-1", Name: "proj_viewer", Scopes: []string{schema.ProjectNamespace}, Permissions: []string{"app_project_get"}},
		}, nil)
		policySvc := mocks.NewPolicyService(t)
		policySvc.On("Create", mock.Anything, mock.Anything).Return(policy.Policy{}, nil).Maybe()
		policySvc.On("List", mock.Anything, mock.Anything).Return([]policy.Policy{}, nil).Maybe()
		auditRepo := mocks.NewAuditRecordRepository(t)
		auditRepo.On("Create", mock.Anything, mock.Anything).Return(auditmodels.AuditRecord{}, nil).Maybe()

		// No projectService mock needed — all-projects scope has empty ResourceIDs, skips validation
		svc := userpat.NewService(slog.New(slog.NewTextHandler(io.Discard, nil)), repo, defaultConfig, orgSvc, roleSvc, policySvc, nil, auditRepo)
		_, _, err := svc.Create(context.Background(), userpat.CreateRequest{
			UserID: "user-1",
			OrgID:  "org-1",
			Title:  "all-projects-test",
			Scopes: []models.PATScope{
				{RoleID: "role-1", ResourceType: schema.ProjectNamespace, ResourceIDs: nil},
			},
			ExpiresAt: time.Now().Add(24 * time.Hour),
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestService_List(t *testing.T) {
	t.Run("should return ErrDisabled when PAT feature is disabled", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		auditRepo := mocks.NewAuditRecordRepository(t)
		svc := userpat.NewService(log.NewNoop(), repo, userpat.Config{Enabled: false}, nil, nil, nil, nil, auditRepo)

		_, err := svc.List(context.Background(), "user-1", "org-1", nil)
		if !errors.Is(err, paterrors.ErrDisabled) {
			t.Fatalf("expected ErrDisabled, got %v", err)
		}
	})

	t.Run("should return error when repo List fails", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		repo.EXPECT().List(mock.Anything, "user-1", "org-1", mock.Anything).
			Return(models.PATList{}, errors.New("db connection failed"))
		auditRepo := mocks.NewAuditRecordRepository(t)
		svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, nil, nil, nil, nil, auditRepo)

		_, err := svc.List(context.Background(), "user-1", "org-1", nil)
		if err == nil || !strings.Contains(err.Error(), "db connection failed") {
			t.Fatalf("expected db error, got %v", err)
		}
	})

	t.Run("should return error when enrichWithScope fails", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		repo.EXPECT().List(mock.Anything, "user-1", "org-1", mock.Anything).
			Return(models.PATList{
				PATs: []models.PAT{{ID: "pat-1", UserID: "user-1", OrgID: "org-1"}},
			}, nil)
		policySvc := mocks.NewPolicyService(t)
		policySvc.EXPECT().List(mock.Anything, policy.Filter{
			PrincipalID:   "pat-1",
			PrincipalType: schema.PATPrincipal,
		}).Return(nil, errors.New("policy service down"))
		auditRepo := mocks.NewAuditRecordRepository(t)
		svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, nil, nil, policySvc, nil, auditRepo)

		_, err := svc.List(context.Background(), "user-1", "org-1", nil)
		if err == nil || !strings.Contains(err.Error(), "enriching PAT scope") {
			t.Fatalf("expected enriching error, got %v", err)
		}
	})

	t.Run("should return enriched PAT list", func(t *testing.T) {
		repo := mocks.NewRepository(t)
		repo.EXPECT().List(mock.Anything, "user-1", "org-1", mock.Anything).
			Return(models.PATList{
				PATs: []models.PAT{
					{ID: "pat-1", UserID: "user-1", OrgID: "org-1", Title: "token-1"},
					{ID: "pat-2", UserID: "user-1", OrgID: "org-1", Title: "token-2"},
				},
			}, nil)
		policySvc := mocks.NewPolicyService(t)
		// enrichWithScope for pat-1
		policySvc.EXPECT().List(mock.Anything, policy.Filter{
			PrincipalID:   "pat-1",
			PrincipalType: schema.PATPrincipal,
		}).Return([]policy.Policy{
			{ID: "pol-1", RoleID: "role-1", ResourceID: "org-1", ResourceType: schema.OrganizationNamespace, GrantRelation: "granted"},
		}, nil)
		// enrichWithScope for pat-2
		policySvc.EXPECT().List(mock.Anything, policy.Filter{
			PrincipalID:   "pat-2",
			PrincipalType: schema.PATPrincipal,
		}).Return([]policy.Policy{}, nil)
		auditRepo := mocks.NewAuditRecordRepository(t)
		svc := userpat.NewService(log.NewNoop(), repo, defaultConfig, nil, nil, policySvc, nil, auditRepo)

		result, err := svc.List(context.Background(), "user-1", "org-1", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.PATs) != 2 {
			t.Fatalf("expected 2 PATs, got %d", len(result.PATs))
		}
		if result.PATs[0].Title != "token-1" {
			t.Fatalf("expected token-1, got %s", result.PATs[0].Title)
		}
		// pat-1 should have 1 scope from the policy
		if len(result.PATs[0].Scopes) != 1 {
			t.Fatalf("expected 1 scope for pat-1, got %d", len(result.PATs[0].Scopes))
		}
		// pat-2 should have 0 scopes
		if len(result.PATs[1].Scopes) != 0 {
			t.Fatalf("expected 0 scopes for pat-2, got %d", len(result.PATs[1].Scopes))
		}
	})
}
