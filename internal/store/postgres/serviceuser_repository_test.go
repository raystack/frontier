package postgres_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

type ServiceUserRepositoryTestSuite struct {
	suite.Suite
	ctx                  context.Context
	client               *db.Client
	repository           *postgres.ServiceUserRepository
	credentialRepository *postgres.ServiceUserCredentialRepository
	orgID                string
}

func (s *ServiceUserRepositoryTestSuite) SetupSuite() {
	var err error

	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewServiceUserRepository(s.client)
	s.credentialRepository = postgres.NewServiceUserCredentialRepository(s.client)

	orgs, err := bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.orgID = orgs[0].ID
}

func (s *ServiceUserRepositoryTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ServiceUserRepositoryTestSuite) TestSkipsSoftDeletedServiceUsers() {
	live, err := s.repository.Create(s.ctx, serviceuser.ServiceUser{OrgID: s.orgID, Title: "live"})
	s.Require().NoError(err)
	deleted, err := s.repository.Create(s.ctx, serviceuser.ServiceUser{OrgID: s.orgID, Title: "gone"})
	s.Require().NoError(err)

	_, err = s.client.ExecContext(s.ctx, "UPDATE serviceusers SET deleted_at = now() WHERE id = $1", deleted.ID)
	s.Require().NoError(err)

	_, err = s.repository.GetByID(s.ctx, deleted.ID)
	s.Assert().ErrorIs(err, serviceuser.ErrNotExist)

	byIDs, err := s.repository.GetByIDs(s.ctx, []string{live.ID, deleted.ID})
	s.Assert().NoError(err)
	s.Require().Len(byIDs, 1)
	s.Assert().Equal(live.ID, byIDs[0].ID)

	got, err := s.repository.List(s.ctx, serviceuser.Filter{OrgID: s.orgID})
	s.Assert().NoError(err)
	ids := make([]string, 0, len(got))
	for _, su := range got {
		ids = append(ids, su.ID)
	}
	s.Assert().Contains(ids, live.ID)
	s.Assert().NotContains(ids, deleted.ID)
}

func (s *ServiceUserRepositoryTestSuite) TestSkipsSoftDeletedCredentials() {
	owner, err := s.repository.Create(s.ctx, serviceuser.ServiceUser{OrgID: s.orgID, Title: "owner"})
	s.Require().NoError(err)

	live, err := s.credentialRepository.Create(s.ctx, serviceuser.Credential{
		ServiceUserID: owner.ID, Type: serviceuser.ClientSecretCredentialType, SecretHash: "h1", Title: "live",
	})
	s.Require().NoError(err)
	deleted, err := s.credentialRepository.Create(s.ctx, serviceuser.Credential{
		ServiceUserID: owner.ID, Type: serviceuser.ClientSecretCredentialType, SecretHash: "h2", Title: "gone",
	})
	s.Require().NoError(err)

	_, err = s.client.ExecContext(s.ctx, "UPDATE serviceuser_credentials SET deleted_at = now() WHERE id = $1", deleted.ID)
	s.Require().NoError(err)

	_, err = s.credentialRepository.Get(s.ctx, deleted.ID)
	s.Assert().ErrorIs(err, serviceuser.ErrCredNotExist)

	got, err := s.credentialRepository.List(s.ctx, serviceuser.Filter{ServiceUserID: owner.ID})
	s.Assert().NoError(err)
	s.Require().Len(got, 1)
	s.Assert().Equal(live.ID, got[0].ID)
}

func (s *ServiceUserRepositoryTestSuite) TestUpdateBootstrapSecretHash() {
	bootstrapSU, err := s.repository.Create(s.ctx, serviceuser.ServiceUser{
		ID: schema.BootstrapServiceUserID, OrgID: schema.PlatformOrgID.String(), Title: "bootstrap",
	})
	s.Require().NoError(err)
	newCred := func(ownerID, title string) serviceuser.Credential {
		cred, err := s.credentialRepository.Create(s.ctx, serviceuser.Credential{
			ServiceUserID: ownerID, Type: serviceuser.ClientSecretCredentialType, SecretHash: "old", Title: title,
		})
		s.Require().NoError(err)
		return cred
	}
	storedHash := func(id string) string {
		var hash string
		s.Require().NoError(s.client.GetContext(s.ctx, &hash, "SELECT secret_hash FROM serviceuser_credentials WHERE id = $1", id))
		return hash
	}

	s.Run("replaces the hash and keeps the row", func() {
		cred := newCred(bootstrapSU.ID, "rotated")
		s.Require().NoError(s.credentialRepository.UpdateBootstrapSecretHash(s.ctx, cred.ID, "new"))

		got, err := s.credentialRepository.Get(s.ctx, cred.ID)
		s.Require().NoError(err)
		s.Assert().Equal("new", got.SecretHash)
		s.Assert().Equal(bootstrapSU.ID, got.ServiceUserID)
		s.Assert().Equal("rotated", got.Title)
		s.Assert().True(got.UpdatedAt.After(cred.UpdatedAt))
	})

	s.Run("reports a missing credential", func() {
		err := s.credentialRepository.UpdateBootstrapSecretHash(s.ctx, uuid.New().String(), "new")
		s.Assert().ErrorIs(err, serviceuser.ErrCredNotExist)
	})

	s.Run("leaves a credential of any other service user alone", func() {
		other, err := s.repository.Create(s.ctx, serviceuser.ServiceUser{OrgID: s.orgID, Title: "other"})
		s.Require().NoError(err)
		cred := newCred(other.ID, "not bootstrap")

		err = s.credentialRepository.UpdateBootstrapSecretHash(s.ctx, cred.ID, "new")
		s.Assert().ErrorIs(err, serviceuser.ErrCredNotExist)
		s.Assert().Equal("old", storedHash(cred.ID))
	})

	s.Run("leaves a soft-deleted credential alone", func() {
		cred := newCred(bootstrapSU.ID, "gone")
		_, err := s.client.ExecContext(s.ctx, "UPDATE serviceuser_credentials SET deleted_at = now() WHERE id = $1", cred.ID)
		s.Require().NoError(err)

		err = s.credentialRepository.UpdateBootstrapSecretHash(s.ctx, cred.ID, "new")
		s.Assert().ErrorIs(err, serviceuser.ErrCredNotExist)
		s.Assert().Equal("old", storedHash(cred.ID))
	})
}

func TestServiceUserRepository(t *testing.T) {
	suite.Run(t, new(ServiceUserRepositoryTestSuite))
}
