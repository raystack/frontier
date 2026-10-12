package serviceuser_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"slices"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/frontier/core/auditrecord/models"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/serviceuser/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	pkgAuditRecord "github.com/raystack/frontier/pkg/auditrecord"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/stretchr/testify/mock"
)

func newTestService(t *testing.T) (*serviceuser.Service, *mocks.Repository, *mocks.CredentialRepository, *mocks.RelationService, *mocks.MembershipService, *mocks.AuditRecordRepository) {
	t.Helper()
	repo := mocks.NewRepository(t)
	credRepo := mocks.NewCredentialRepository(t)
	relSvc := mocks.NewRelationService(t)
	memSvc := mocks.NewMembershipService(t)
	auditRepo := mocks.NewAuditRecordRepository(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := serviceuser.NewService(logger, repo, credRepo, relSvc, auditRepo)
	svc.SetMembershipService(memSvc)
	return svc, repo, credRepo, relSvc, memSvc, auditRepo
}

func TestService_Sudo(t *testing.T) {
	ctx := context.Background()
	const suID = "550e8400-e29b-41d4-a716-446655440000"
	// platformRel builds the platform relation used by IsSudo's check and by
	// Create/Delete (field order in the literal doesn't matter for struct equality).
	platformRel := func(rel string) relation.Relation {
		return relation.Relation{
			Object:       relation.Object{ID: schema.PlatformID, Namespace: schema.PlatformNamespace},
			Subject:      relation.Subject{ID: suID, Namespace: schema.ServiceUserPrincipal},
			RelationName: rel,
		}
	}

	t.Run("grants admin relation and audits the grant", func(t *testing.T) {
		svc, repo, _, rel, _, audit := newTestService(t)
		repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, Title: "svc"}, nil)
		// Sudo checks the exact relation (admin), not the superuser permission, and
		// is safe to run again — the same way UnSudo works.
		rel.On("CheckPermission", ctx, platformRel(schema.AdminRelationName)).Return(false, nil)
		rel.On("Create", ctx, platformRel(schema.AdminRelationName)).Return(relation.Relation{}, nil)
		audit.On("Create", ctx, mock.MatchedBy(func(r models.AuditRecord) bool {
			return r.Event == pkgAuditRecord.PlatformAdminAddedEvent && r.Target != nil && r.Target.ID == suID
		})).Return(models.AuditRecord{}, nil)

		if err := svc.Sudo(ctx, suID, schema.AdminRelationName); err != nil {
			t.Fatalf("Sudo() error = %v", err)
		}
	})

	t.Run("grants member relation and audits the grant", func(t *testing.T) {
		svc, repo, _, rel, _, audit := newTestService(t)
		repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, Title: "svc"}, nil)
		// member is checked at the relation level, so an existing admin still gets
		// the member relation created (the downgrade path).
		rel.On("CheckPermission", ctx, platformRel(schema.MemberRelationName)).Return(false, nil)
		rel.On("Create", ctx, platformRel(schema.MemberRelationName)).Return(relation.Relation{}, nil)
		audit.On("Create", ctx, mock.MatchedBy(func(r models.AuditRecord) bool {
			return r.Event == pkgAuditRecord.PlatformMemberAddedEvent && r.Target != nil && r.Target.ID == suID
		})).Return(models.AuditRecord{}, nil)

		if err := svc.Sudo(ctx, suID, schema.MemberRelationName); err != nil {
			t.Fatalf("Sudo() error = %v", err)
		}
	})
}

func TestService_UnSudo(t *testing.T) {
	ctx := context.Background()
	const suID = "550e8400-e29b-41d4-a716-446655440000"
	platformRel := func(rel string) relation.Relation {
		return relation.Relation{
			Object:       relation.Object{ID: schema.PlatformID, Namespace: schema.PlatformNamespace},
			Subject:      relation.Subject{ID: suID, Namespace: schema.ServiceUserPrincipal},
			RelationName: rel,
		}
	}

	t.Run("removes admin relation and audits the revoke", func(t *testing.T) {
		svc, repo, _, rel, _, audit := newTestService(t)
		repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, Title: "svc"}, nil)
		// UnSudo checks the relation directly (admin), not the superuser permission
		rel.On("CheckPermission", ctx, platformRel(schema.AdminRelationName)).Return(true, nil)
		rel.On("Delete", ctx, platformRel(schema.AdminRelationName)).Return(nil)
		audit.On("Create", ctx, mock.MatchedBy(func(r models.AuditRecord) bool {
			return r.Event == pkgAuditRecord.PlatformAdminRemovedEvent && r.Target != nil && r.Target.ID == suID
		})).Return(models.AuditRecord{}, nil)

		if err := svc.UnSudo(ctx, suID, schema.AdminRelationName); err != nil {
			t.Fatalf("UnSudo() error = %v", err)
		}
	})

	t.Run("admin removal makes no change when the relation is absent", func(t *testing.T) {
		svc, repo, _, rel, _, _ := newTestService(t)
		repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, Title: "svc"}, nil)
		rel.On("CheckPermission", ctx, platformRel(schema.AdminRelationName)).Return(false, nil)

		if err := svc.UnSudo(ctx, suID, schema.AdminRelationName); err != nil {
			t.Fatalf("UnSudo() error = %v", err)
		}
	})

	t.Run("removes member relation and audits the revoke", func(t *testing.T) {
		svc, repo, _, rel, _, audit := newTestService(t)
		repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, Title: "svc"}, nil)
		rel.On("CheckPermission", ctx, platformRel(schema.MemberRelationName)).Return(true, nil)
		rel.On("Delete", ctx, platformRel(schema.MemberRelationName)).Return(nil)
		audit.On("Create", ctx, mock.MatchedBy(func(r models.AuditRecord) bool {
			return r.Event == pkgAuditRecord.PlatformMemberRemovedEvent && r.Target != nil && r.Target.ID == suID
		})).Return(models.AuditRecord{}, nil)

		if err := svc.UnSudo(ctx, suID, schema.MemberRelationName); err != nil {
			t.Fatalf("UnSudo() error = %v", err)
		}
	})

	t.Run("member removal makes no change when the relation is absent", func(t *testing.T) {
		svc, repo, _, rel, _, _ := newTestService(t)
		repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, Title: "svc"}, nil)
		rel.On("CheckPermission", ctx, platformRel(schema.MemberRelationName)).Return(false, nil)

		if err := svc.UnSudo(ctx, suID, schema.MemberRelationName); err != nil {
			t.Fatalf("UnSudo() error = %v", err)
		}
	})

	t.Run("rejects an invalid relation name", func(t *testing.T) {
		svc, _, _, _, _, _ := newTestService(t)
		if err := svc.UnSudo(ctx, suID, "owner"); err == nil {
			t.Fatal("UnSudo() expected error for invalid relation, got nil")
		}
	})
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	const suID = "su-id"
	su := serviceuser.ServiceUser{ID: suID, OrgID: "org-id"}
	credFilter := serviceuser.Filter{ServiceUserID: suID}

	subjectFilter := relation.Relation{
		Subject: relation.Subject{ID: suID, Namespace: schema.ServiceUserPrincipal},
	}
	objectFilter := relation.Relation{
		Object: relation.Object{ID: suID, Namespace: schema.ServiceUserPrincipal},
	}

	t.Run("removes policies, then credentials, then relations, then the row", func(t *testing.T) {
		svc, repo, cred, rel, mem, _ := newTestService(t)
		var steps []string
		step := func(name string) func(mock.Arguments) {
			return func(mock.Arguments) { steps = append(steps, name) }
		}
		repo.On("GetByID", ctx, suID).Return(su, nil)
		mem.On("RemovePrincipalPolicies", ctx, suID, schema.ServiceUserPrincipal).Run(step("policies")).Return(nil)
		cred.On("List", ctx, credFilter).Return([]serviceuser.Credential{{ID: "c1"}, {ID: "c2"}}, nil)
		cred.On("Delete", ctx, "c1").Run(step("credential c1")).Return(nil)
		cred.On("Delete", ctx, "c2").Run(step("credential c2")).Return(nil)
		rel.On("Delete", ctx, subjectFilter).Run(step("relations as subject")).Return(nil)
		rel.On("Delete", ctx, objectFilter).Run(step("relations as object")).Return(nil)
		repo.On("Delete", ctx, suID).Run(step("row")).Return(nil)

		if err := svc.Delete(ctx, suID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		want := []string{"policies", "credential c1", "credential c2", "relations as subject", "relations as object", "row"}
		if !slices.Equal(steps, want) {
			t.Fatalf("order = %v, want %v", steps, want)
		}
	})

	tests := []struct {
		name      string
		setup     func(*mocks.Repository, *mocks.CredentialRepository, *mocks.RelationService, *mocks.MembershipService)
		wantErr   bool
		wantErrIs error
	}{
		{
			name: "a missing service user is not found before anything runs",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{}, serviceuser.ErrNotExist)
			},
			wantErr:   true,
			wantErrIs: serviceuser.ErrNotExist,
		},
		{
			name: "a policy cleanup failure stops the delete before credentials and the row",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(su, nil)
				mem.On("RemovePrincipalPolicies", ctx, suID, schema.ServiceUserPrincipal).Return(errors.New("spicedb unavailable"))
			},
			wantErr: true,
		},
		{
			name: "a credential that is already deleted is skipped",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(su, nil)
				mem.On("RemovePrincipalPolicies", ctx, suID, schema.ServiceUserPrincipal).Return(nil)
				cred.On("List", ctx, credFilter).Return([]serviceuser.Credential{{ID: "c1"}}, nil)
				cred.On("Delete", ctx, "c1").Return(serviceuser.ErrCredNotExist)
				rel.On("Delete", ctx, subjectFilter).Return(nil)
				rel.On("Delete", ctx, objectFilter).Return(nil)
				repo.On("Delete", ctx, suID).Return(nil)
			},
		},
		{
			name: "a credential delete failure stops the delete before the row",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(su, nil)
				mem.On("RemovePrincipalPolicies", ctx, suID, schema.ServiceUserPrincipal).Return(nil)
				cred.On("List", ctx, credFilter).Return([]serviceuser.Credential{{ID: "c1"}}, nil)
				cred.On("Delete", ctx, "c1").Return(errors.New("db down"))
			},
			wantErr: true,
		},
		{
			name: "tolerates ErrNotExist from the object-side sweep",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(su, nil)
				mem.On("RemovePrincipalPolicies", ctx, suID, schema.ServiceUserPrincipal).Return(nil)
				cred.On("List", ctx, credFilter).Return([]serviceuser.Credential{}, nil)
				rel.On("Delete", ctx, subjectFilter).Return(nil)
				rel.On("Delete", ctx, objectFilter).Return(relation.ErrNotExist)
				repo.On("Delete", ctx, suID).Return(nil)
			},
		},
		{
			name: "an object-side failure stops the delete before the row",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(su, nil)
				mem.On("RemovePrincipalPolicies", ctx, suID, schema.ServiceUserPrincipal).Return(nil)
				cred.On("List", ctx, credFilter).Return([]serviceuser.Credential{}, nil)
				rel.On("Delete", ctx, subjectFilter).Return(nil)
				rel.On("Delete", ctx, objectFilter).Return(errors.New("spicedb unavailable"))
			},
			wantErr: true,
		},
		{
			name: "a subject-side failure stops the delete before the object sweep",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(su, nil)
				mem.On("RemovePrincipalPolicies", ctx, suID, schema.ServiceUserPrincipal).Return(nil)
				cred.On("List", ctx, credFilter).Return([]serviceuser.Credential{}, nil)
				rel.On("Delete", ctx, subjectFilter).Return(errors.New("spicedb unavailable"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, cred, rel, mem, _ := newTestService(t)
			tt.setup(repo, cred, rel, mem)

			err := svc.Delete(ctx, suID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErrIs != nil && !errors.Is(err, tt.wantErrIs) {
				t.Errorf("Delete() error = %v, want %v", err, tt.wantErrIs)
			}
		})
	}
}

func TestService_CredentialsNeedALiveServiceUser(t *testing.T) {
	ctx := context.Background()
	const suID = "0b6a0c4e-5f3d-4a2b-9c8e-1d2f3a4b5c6d"

	calls := []struct {
		name string
		call func(*serviceuser.Service) error
	}{
		{"CreateKey", func(svc *serviceuser.Service) error {
			_, err := svc.CreateKey(ctx, serviceuser.Credential{ServiceUserID: suID})
			return err
		}},
		{"CreateSecret", func(svc *serviceuser.Service) error {
			_, err := svc.CreateSecret(ctx, serviceuser.Credential{ServiceUserID: suID})
			return err
		}},
		{"CreateToken", func(svc *serviceuser.Service) error {
			_, err := svc.CreateToken(ctx, serviceuser.Credential{ServiceUserID: suID})
			return err
		}},
		{"ListKeys", func(svc *serviceuser.Service) error {
			_, err := svc.ListKeys(ctx, suID)
			return err
		}},
		{"ListSecret", func(svc *serviceuser.Service) error {
			_, err := svc.ListSecret(ctx, suID)
			return err
		}},
	}
	for _, tc := range calls {
		t.Run(tc.name+" on a deleted service user is not found and touches no credential", func(t *testing.T) {
			svc, repo, _, _, _, _ := newTestService(t)
			repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{}, serviceuser.ErrNotExist)

			if err := tc.call(svc); !errors.Is(err, serviceuser.ErrNotExist) {
				t.Fatalf("want ErrNotExist, got %v", err)
			}
		})
	}

	t.Run("ListKeys on a live service user lists its keys", func(t *testing.T) {
		svc, repo, credRepo, _, _, _ := newTestService(t)
		repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID}, nil)
		credRepo.On("List", ctx, serviceuser.Filter{ServiceUserID: suID, IsKey: true}).Return([]serviceuser.Credential{{ID: "k1"}}, nil)

		keys, err := svc.ListKeys(ctx, suID)
		if err != nil || len(keys) != 1 || keys[0].ID != "k1" {
			t.Fatalf("ListKeys() = %v, %v", keys, err)
		}
	})

	t.Run("CreateSecret on a live service user stores a client secret", func(t *testing.T) {
		svc, repo, credRepo, _, _, _ := newTestService(t)
		repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID}, nil)
		credRepo.On("Create", ctx, mock.MatchedBy(func(c serviceuser.Credential) bool {
			return c.ServiceUserID == suID && c.Type == serviceuser.ClientSecretCredentialType && c.SecretHash != ""
		})).Return(serviceuser.Credential{ID: "s1"}, nil)

		secret, err := svc.CreateSecret(ctx, serviceuser.Credential{ServiceUserID: suID})
		if err != nil || secret.ID != "s1" || secret.Value == "" {
			t.Fatalf("CreateSecret() = %+v, %v", secret, err)
		}
	})

	t.Run("an id that is not a uuid is rejected without a lookup", func(t *testing.T) {
		svc, _, _, _, _, _ := newTestService(t)

		if _, err := svc.ListSecret(ctx, "not-a-uuid"); !errors.Is(err, serviceuser.ErrInvalidID) {
			t.Fatalf("want ErrInvalidID, got %v", err)
		}
	})
}

func TestService_Get(t *testing.T) {
	ctx := context.Background()
	const validID = "68f86fec-eb87-49f0-9be0-8d99b00a4a9c"

	tests := []struct {
		name      string
		id        string
		setup     func(*mocks.Repository)
		wantErrIs error
	}{
		{
			name:      "empty id returns ErrInvalidID without hitting the repo",
			id:        "",
			setup:     func(repo *mocks.Repository) {},
			wantErrIs: serviceuser.ErrInvalidID,
		},
		{
			name:      "non-uuid id returns ErrInvalidID without hitting the repo",
			id:        "not-a-uuid",
			setup:     func(repo *mocks.Repository) {},
			wantErrIs: serviceuser.ErrInvalidID,
		},
		{
			name: "valid uuid delegates to the repo",
			id:   validID,
			setup: func(repo *mocks.Repository) {
				repo.On("GetByID", ctx, validID).Return(serviceuser.ServiceUser{ID: validID}, nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _, _, _ := newTestService(t)
			tt.setup(repo)

			_, err := svc.Get(ctx, tt.id)
			if tt.wantErrIs != nil {
				if !errors.Is(err, tt.wantErrIs) {
					t.Errorf("Get() error = %v, want errors.Is(%v)", err, tt.wantErrIs)
				}
				return
			}
			if err != nil {
				t.Errorf("Get() unexpected error = %v", err)
			}
		})
	}
}

func TestService_ListByOrg(t *testing.T) {
	ctx := context.Background()
	const orgID = "org-id"

	tests := []struct {
		name    string
		setup   func(*mocks.Repository, *mocks.MembershipService)
		want    int
		wantErr bool
	}{
		{
			name: "members found are fetched from the repo",
			setup: func(repo *mocks.Repository, mem *mocks.MembershipService) {
				mem.On("ListPrincipalIDsByResource", ctx, orgID, schema.OrganizationNamespace, schema.ServiceUserPrincipal).
					Return([]string{"su-1", "su-2"}, nil)
				repo.On("GetByIDs", ctx, []string{"su-1", "su-2"}).
					Return([]serviceuser.ServiceUser{{ID: "su-1"}, {ID: "su-2"}}, nil)
			},
			want: 2,
		},
		{
			name: "no members returns empty list without hitting the repo",
			setup: func(repo *mocks.Repository, mem *mocks.MembershipService) {
				mem.On("ListPrincipalIDsByResource", ctx, orgID, schema.OrganizationNamespace, schema.ServiceUserPrincipal).
					Return([]string{}, nil)
				// repo.GetByIDs must NOT be called
			},
			want: 0,
		},
		{
			name: "membership error is propagated",
			setup: func(repo *mocks.Repository, mem *mocks.MembershipService) {
				mem.On("ListPrincipalIDsByResource", ctx, orgID, schema.OrganizationNamespace, schema.ServiceUserPrincipal).
					Return(nil, errors.New("policy store unavailable"))
			},
			wantErr: true,
		},
		{
			name: "repo error after successful ID lookup is propagated",
			setup: func(repo *mocks.Repository, mem *mocks.MembershipService) {
				mem.On("ListPrincipalIDsByResource", ctx, orgID, schema.OrganizationNamespace, schema.ServiceUserPrincipal).
					Return([]string{"su-1"}, nil)
				repo.On("GetByIDs", ctx, []string{"su-1"}).
					Return(nil, errors.New("db unavailable"))
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, _, _, mem, _ := newTestService(t)
			tt.setup(repo, mem)

			got, err := svc.ListByOrg(ctx, orgID)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ListByOrg() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("ListByOrg() returned %d service users, want %d", len(got), tt.want)
			}
		})
	}
}

func TestService_DeleteKey_BootstrapProtected(t *testing.T) {
	ctx := context.Background()
	const credID = "cred-1"

	t.Run("refuses deleting a credential owned by the bootstrap SA", func(t *testing.T) {
		svc, _, credRepo, _, _, _ := newTestService(t)
		credRepo.On("Get", ctx, credID).Return(
			serviceuser.Credential{ID: credID, ServiceUserID: schema.BootstrapServiceUserID}, nil)

		err := svc.DeleteSecret(ctx, credID) // funnels through DeleteKey
		if !errors.Is(err, serviceuser.ErrProtected) {
			t.Fatalf("want ErrProtected, got %v", err)
		}
		credRepo.AssertNotCalled(t, "Delete", mock.Anything, mock.Anything)
	})

	t.Run("deletes a credential of a normal service user", func(t *testing.T) {
		svc, _, credRepo, _, _, _ := newTestService(t)
		credRepo.On("Get", ctx, credID).Return(
			serviceuser.Credential{ID: credID, ServiceUserID: "11111111-2222-3333-4444-555555555555"}, nil)
		credRepo.On("Delete", ctx, credID).Return(nil)

		if err := svc.DeleteToken(ctx, credID); err != nil { // also funnels through DeleteKey
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestService_GetByJWT_Classification(t *testing.T) {
	ctx := context.Background()

	// buildToken signs a jwt carrying kid in its claims and returns the matching public set.
	buildToken := func(t *testing.T, kid string) ([]byte, jwk.Set) {
		t.Helper()
		key, err := utils.CreateJWKWithKID(kid)
		if err != nil {
			t.Fatalf("CreateJWKWithKID: %v", err)
		}
		tok, err := utils.BuildToken(key, "issuer", "subject", time.Hour, nil)
		if err != nil {
			t.Fatalf("BuildToken: %v", err)
		}
		set := jwk.NewSet()
		if err := set.AddKey(key); err != nil {
			t.Fatalf("AddKey: %v", err)
		}
		pub, err := utils.GetPublicKeySet(ctx, set)
		if err != nil {
			t.Fatalf("GetPublicKeySet: %v", err)
		}
		return tok, pub
	}

	t.Run("not a jwt skips with ErrTokenNotJWT", func(t *testing.T) {
		svc, _, _, _, _, _ := newTestService(t)
		if _, err := svc.GetByJWT(ctx, "fpt_not-a-jwt"); !errors.Is(err, serviceuser.ErrTokenNotJWT) {
			t.Errorf("GetByJWT() error = %v, want errors.Is(ErrTokenNotJWT)", err)
		}
	})

	t.Run("malformed (non-uuid) kid skips with ErrInvalidKeyID", func(t *testing.T) {
		svc, _, _, _, _, _ := newTestService(t)
		tok, _ := buildToken(t, "not-a-uuid")
		// credRepo.Get must not be called for a malformed kid
		if _, err := svc.GetByJWT(ctx, string(tok)); !errors.Is(err, serviceuser.ErrInvalidKeyID) {
			t.Errorf("GetByJWT() error = %v, want errors.Is(ErrInvalidKeyID)", err)
		}
	})

	t.Run("non-string kid skips with ErrInvalidKeyID", func(t *testing.T) {
		svc, _, _, _, _, _ := newTestService(t)
		key, err := utils.CreateJWKWithKID("33333333-3333-3333-3333-333333333333")
		if err != nil {
			t.Fatalf("CreateJWKWithKID: %v", err)
		}
		tok, err := jwt.NewBuilder().Claim(jwk.KeyIDKey, 12345).Build()
		if err != nil {
			t.Fatalf("Build: %v", err)
		}
		signed, err := jwt.Sign(tok, jwt.WithKey(jwa.RS256, key))
		if err != nil {
			t.Fatalf("Sign: %v", err)
		}
		if _, err := svc.GetByJWT(ctx, string(signed)); !errors.Is(err, serviceuser.ErrInvalidKeyID) {
			t.Errorf("GetByJWT() error = %v, want errors.Is(ErrInvalidKeyID)", err)
		}
	})

	t.Run("unknown kid skips with ErrCredNotExist", func(t *testing.T) {
		svc, _, credRepo, _, _, _ := newTestService(t)
		kid := "11111111-1111-1111-1111-111111111111"
		tok, _ := buildToken(t, kid)
		credRepo.On("Get", ctx, kid).Return(serviceuser.Credential{}, serviceuser.ErrCredNotExist)
		if _, err := svc.GetByJWT(ctx, string(tok)); !errors.Is(err, serviceuser.ErrCredNotExist) {
			t.Errorf("GetByJWT() error = %v, want errors.Is(ErrCredNotExist)", err)
		}
	})

	t.Run("bad signature stops with ErrInvalidCred", func(t *testing.T) {
		svc, _, credRepo, _, _, _ := newTestService(t)
		kid := "22222222-2222-2222-2222-222222222222"
		tok, _ := buildToken(t, kid)      // signed by one key
		_, otherPub := buildToken(t, kid) // verified against a different key with the same kid
		credRepo.On("Get", ctx, kid).Return(
			serviceuser.Credential{ID: kid, ServiceUserID: "su-1", PublicKey: otherPub}, nil)
		if _, err := svc.GetByJWT(ctx, string(tok)); !errors.Is(err, serviceuser.ErrInvalidCred) {
			t.Errorf("GetByJWT() error = %v, want errors.Is(ErrInvalidCred)", err)
		}
	})
}
