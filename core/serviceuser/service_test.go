package serviceuser_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"

	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/serviceuser/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
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
