package serviceuser_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/lestrrat-go/jwx/v2/jwa"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/lestrrat-go/jwx/v2/jwt"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/serviceuser/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/utils"
)

func newTestService(t *testing.T) (*serviceuser.Service, *mocks.Repository, *mocks.CredentialRepository, *mocks.RelationService, *mocks.MembershipService) {
	t.Helper()
	repo := mocks.NewRepository(t)
	credRepo := mocks.NewCredentialRepository(t)
	relSvc := mocks.NewRelationService(t)
	memSvc := mocks.NewMembershipService(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := serviceuser.NewService(logger, repo, credRepo, relSvc)
	svc.SetMembershipService(memSvc)
	return svc, repo, credRepo, relSvc, memSvc
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	const suID = "su-id"
	const orgID = "org-id"

	subjectFilter := relation.Relation{
		Subject: relation.Subject{ID: suID, Namespace: schema.ServiceUserPrincipal},
	}
	objectFilter := relation.Relation{
		Object: relation.Object{ID: suID, Namespace: schema.ServiceUserPrincipal},
	}

	tests := []struct {
		name    string
		setup   func(*mocks.Repository, *mocks.CredentialRepository, *mocks.RelationService, *mocks.MembershipService)
		wantErr bool
	}{
		{
			name: "sweeps SU as subject and as object",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID}, nil)
				cred.On("List", ctx, serviceuser.Filter{ServiceUserID: suID}).Return([]serviceuser.Credential{}, nil)
				mem.On("RemoveOrganizationMember", ctx, orgID, suID, schema.ServiceUserPrincipal).Return(nil)
				rel.On("Delete", ctx, subjectFilter).Return(nil)
				rel.On("Delete", ctx, objectFilter).Return(nil)
				repo.On("Delete", ctx, suID).Return(nil)
			},
		},
		{
			name: "membership failure is swallowed and both sweeps still run",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID}, nil)
				cred.On("List", ctx, serviceuser.Filter{ServiceUserID: suID}).Return([]serviceuser.Credential{}, nil)
				// covers the path where membership returns early without reaching its
				// cascade cleanup (e.g. SU has no remaining org policies)
				mem.On("RemoveOrganizationMember", ctx, orgID, suID, schema.ServiceUserPrincipal).Return(errors.New("not a member"))
				rel.On("Delete", ctx, subjectFilter).Return(nil)
				rel.On("Delete", ctx, objectFilter).Return(nil)
				repo.On("Delete", ctx, suID).Return(nil)
			},
		},
		{
			name: "tolerates ErrNotExist from Object-side sweep",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID}, nil)
				cred.On("List", ctx, serviceuser.Filter{ServiceUserID: suID}).Return([]serviceuser.Credential{}, nil)
				mem.On("RemoveOrganizationMember", ctx, orgID, suID, schema.ServiceUserPrincipal).Return(nil)
				rel.On("Delete", ctx, subjectFilter).Return(nil)
				rel.On("Delete", ctx, objectFilter).Return(relation.ErrNotExist)
				repo.On("Delete", ctx, suID).Return(nil)
			},
		},
		{
			name: "Object-side non-ErrNotExist failure blocks repo delete",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID}, nil)
				cred.On("List", ctx, serviceuser.Filter{ServiceUserID: suID}).Return([]serviceuser.Credential{}, nil)
				mem.On("RemoveOrganizationMember", ctx, orgID, suID, schema.ServiceUserPrincipal).Return(nil)
				rel.On("Delete", ctx, subjectFilter).Return(nil)
				rel.On("Delete", ctx, objectFilter).Return(errors.New("spicedb unavailable"))
				// repo.Delete must NOT be called
			},
			wantErr: true,
		},
		{
			name: "Subject-side failure short-circuits before Object sweep",
			setup: func(repo *mocks.Repository, cred *mocks.CredentialRepository, rel *mocks.RelationService, mem *mocks.MembershipService) {
				repo.On("GetByID", ctx, suID).Return(serviceuser.ServiceUser{ID: suID, OrgID: orgID}, nil)
				cred.On("List", ctx, serviceuser.Filter{ServiceUserID: suID}).Return([]serviceuser.Credential{}, nil)
				mem.On("RemoveOrganizationMember", ctx, orgID, suID, schema.ServiceUserPrincipal).Return(nil)
				rel.On("Delete", ctx, subjectFilter).Return(errors.New("spicedb unavailable"))
				// Object-side Delete and repo.Delete must NOT be called
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, repo, cred, rel, mem := newTestService(t)
			tt.setup(repo, cred, rel, mem)

			err := svc.Delete(ctx, suID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Delete() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
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
			svc, repo, _, _, _ := newTestService(t)
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
			svc, repo, _, _, mem := newTestService(t)
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
		svc, _, _, _, _ := newTestService(t)
		if _, err := svc.GetByJWT(ctx, "fpt_not-a-jwt"); !errors.Is(err, serviceuser.ErrTokenNotJWT) {
			t.Errorf("GetByJWT() error = %v, want errors.Is(ErrTokenNotJWT)", err)
		}
	})

	t.Run("malformed (non-uuid) kid skips with ErrInvalidKeyID", func(t *testing.T) {
		svc, _, _, _, _ := newTestService(t)
		tok, _ := buildToken(t, "not-a-uuid")
		// credRepo.Get must not be called for a malformed kid
		if _, err := svc.GetByJWT(ctx, string(tok)); !errors.Is(err, serviceuser.ErrInvalidKeyID) {
			t.Errorf("GetByJWT() error = %v, want errors.Is(ErrInvalidKeyID)", err)
		}
	})

	t.Run("non-string kid skips with ErrInvalidKeyID", func(t *testing.T) {
		svc, _, _, _, _ := newTestService(t)
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
		svc, _, credRepo, _, _ := newTestService(t)
		kid := "11111111-1111-1111-1111-111111111111"
		tok, _ := buildToken(t, kid)
		credRepo.On("Get", ctx, kid).Return(serviceuser.Credential{}, serviceuser.ErrCredNotExist)
		if _, err := svc.GetByJWT(ctx, string(tok)); !errors.Is(err, serviceuser.ErrCredNotExist) {
			t.Errorf("GetByJWT() error = %v, want errors.Is(ErrCredNotExist)", err)
		}
	})

	t.Run("bad signature stops with ErrInvalidCred", func(t *testing.T) {
		svc, _, credRepo, _, _ := newTestService(t)
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
