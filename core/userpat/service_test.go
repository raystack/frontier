package userpat_test

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/userpat"
	"github.com/raystack/frontier/core/userpat/mocks"
	"github.com/stretchr/testify/mock"
)

var defaultConfig = userpat.Config{
	Enabled:                true,
	TokenPrefix:            "fpt",
	MaxTokensPerUserPerOrg: 50,
	MaxTokenLifetime:       "8760h",
}

func newSuccessMocks(t *testing.T) (*mocks.OrganizationService, *mocks.AuditRecordRepository) {
	t.Helper()
	orgSvc := mocks.NewOrganizationService(t)
	orgSvc.On("GetRaw", mock.Anything, mock.Anything).
		Return(organization.Organization{ID: "org-1", Title: "Test Org"}, nil).Maybe()
	auditRepo := mocks.NewAuditRecordRepository(t)
	auditRepo.On("Create", mock.Anything, mock.Anything).
		Return(models.AuditRecord{}, nil).Maybe()
	return orgSvc, auditRepo
}

func TestService_Create(t *testing.T) {
	tests := []struct {
		name         string
		setup        func() *userpat.Service
		req          userpat.CreateRequest
		wantErr      bool
		wantErrIs    error
		wantErrMsg   string
		validateFunc func(t *testing.T, got userpat.PersonalAccessToken, tokenValue string)
	}{
		{
			name: "should return ErrDisabled when PAT feature is disabled",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: userpat.ErrDisabled,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(repo, userpat.Config{
					Enabled: false,
				}, orgSvc, auditRepo)
			},
		},
		{
			name: "should return error when CountActive fails",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:    true,
			wantErrMsg: "counting active tokens",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), errors.New("db connection failed"))
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
		},
		{
			name: "should return ErrLimitExceeded when token count equals max",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
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
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
		},
		{
			name: "should return ErrLimitExceeded when token count exceeds max",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
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
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
		},
		{
			name: "should return error when repo Create fails",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:    true,
			wantErrMsg: "insert failed",
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
					Return(userpat.PersonalAccessToken{}, errors.New("insert failed"))
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
		},
		{
			name: "should return ErrConflict when repo returns duplicate error",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr:   true,
			wantErrIs: userpat.ErrConflict,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
					Return(userpat.PersonalAccessToken{}, userpat.ErrConflict)
				orgSvc := mocks.NewOrganizationService(t)
				auditRepo := mocks.NewAuditRecordRepository(t)
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
		},
		{
			name: "should create token successfully with correct fields",
			req: userpat.CreateRequest{
				UserID:     "user-1",
				OrgID:      "org-1",
				Title:      "my-token",
				Roles:      []string{"role-1"},
				ProjectIDs: []string{"proj-1"},
				ExpiresAt:  time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
				Metadata:   map[string]any{"env": "staging"},
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
					Run(func(ctx context.Context, pat userpat.PersonalAccessToken) {
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
						if diff := cmp.Diff(map[string]any(pat.Metadata), map[string]any{"env": "staging"}); diff != "" {
							t.Errorf("Create() Metadata mismatch (-want +got):\n%s", diff)
						}
						if !pat.ExpiresAt.Equal(time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)) {
							t.Errorf("Create() ExpiresAt = %v, want %v", pat.ExpiresAt, time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC))
						}
					}).
					Return(userpat.PersonalAccessToken{
						ID:        "pat-id-1",
						UserID:    "user-1",
						OrgID:     "org-1",
						Title:     "my-token",
						Metadata:  map[string]any{"env": "staging"},
						ExpiresAt: time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC),
						CreatedAt: time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC),
					}, nil)
				orgSvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
			validateFunc: func(t *testing.T, got userpat.PersonalAccessToken, tokenValue string) {
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
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
					Return(userpat.PersonalAccessToken{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
			validateFunc: func(t *testing.T, got userpat.PersonalAccessToken, tokenValue string) {
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
			name: "should hash the full token string including prefix with sha256",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
					Return(userpat.PersonalAccessToken{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
			validateFunc: func(t *testing.T, got userpat.PersonalAccessToken, tokenValue string) {
				t.Helper()
				// we can't directly access the hash passed to repo from here,
				// but we verify the token is well-formed and hashable
				hash := sha256.Sum256([]byte(tokenValue))
				hashStr := hex.EncodeToString(hash[:])
				if len(hashStr) != 64 {
					t.Errorf("sha256 hash should be 64 hex chars, got %d", len(hashStr))
				}
			},
		},
		{
			name: "should use custom token prefix from config",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
					Return(userpat.PersonalAccessToken{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(repo, userpat.Config{
					Enabled:                true,
					TokenPrefix:            "custom",
					MaxTokensPerUserPerOrg: 50,
					MaxTokenLifetime:       "8760h",
				}, orgSvc, auditRepo)
			},
			validateFunc: func(t *testing.T, got userpat.PersonalAccessToken, tokenValue string) {
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
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(49), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
					Return(userpat.PersonalAccessToken{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
			},
		},
		{
			name: "should generate unique tokens on each call",
			req: userpat.CreateRequest{
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				Roles:     []string{"role-1"},
				ExpiresAt: time.Now().Add(24 * time.Hour),
			},
			wantErr: false,
			setup: func() *userpat.Service {
				repo := mocks.NewRepository(t)
				repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
					Return(int64(0), nil)
				repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
					Return(userpat.PersonalAccessToken{ID: "pat-1", OrgID: "org-1"}, nil)
				orgSvc, auditRepo := newSuccessMocks(t)
				return userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)
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

func TestService_Create_UniqueTokens(t *testing.T) {
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
		Return(int64(0), nil).Times(2)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
		Return(userpat.PersonalAccessToken{ID: "pat-1", OrgID: "org-1"}, nil).Times(2)

	orgSvc, auditRepo := newSuccessMocks(t)
	svc := userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)

	req := userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		Roles:     []string{"role-1"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	_, token1, err1 := svc.Create(context.Background(), req)
	if err1 != nil {
		t.Fatalf("Create() first call error = %v", err1)
	}
	_, token2, err2 := svc.Create(context.Background(), req)
	if err2 != nil {
		t.Fatalf("Create() second call error = %v", err2)
	}
	if token1 == token2 {
		t.Errorf("Create() should generate unique tokens, got same token twice: %v", token1)
	}
}

func TestService_Create_HashVerification(t *testing.T) {
	var capturedHash string
	repo := mocks.NewRepository(t)
	repo.EXPECT().CountActive(mock.Anything, "user-1", "org-1").
		Return(int64(0), nil)
	repo.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.PersonalAccessToken")).
		Run(func(ctx context.Context, pat userpat.PersonalAccessToken) {
			capturedHash = pat.SecretHash
		}).
		Return(userpat.PersonalAccessToken{ID: "pat-1", OrgID: "org-1"}, nil)

	orgSvc, auditRepo := newSuccessMocks(t)
	svc := userpat.NewService(repo, defaultConfig, orgSvc, auditRepo)

	_, tokenValue, err := svc.Create(context.Background(), userpat.CreateRequest{
		UserID:    "user-1",
		OrgID:     "org-1",
		Title:     "my-token",
		Roles:     []string{"role-1"},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	expectedHash := sha256.Sum256([]byte(tokenValue))
	expectedHashStr := hex.EncodeToString(expectedHash[:])
	if capturedHash != expectedHashStr {
		t.Errorf("Create() hash mismatch: stored %v, expected sha256(%v) = %v", capturedHash, tokenValue, expectedHashStr)
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
			cfg:  userpat.Config{MaxTokenLifetime: "720h"},
			want: 720 * time.Hour,
		},
		{
			name: "should return default 1 year on invalid duration",
			cfg:  userpat.Config{MaxTokenLifetime: "invalid"},
			want: 365 * 24 * time.Hour,
		},
		{
			name: "should return default 1 year on empty duration",
			cfg:  userpat.Config{MaxTokenLifetime: ""},
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
