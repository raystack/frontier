package postgres_test

import (
	"context"
	"testing"

	"github.com/raystack/frontier/core/serviceuser"
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

func TestServiceUserRepository(t *testing.T) {
	suite.Run(t, new(ServiceUserRepositoryTestSuite))
}
