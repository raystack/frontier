package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

type OrgBillingRepositoryPGTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.OrgBillingRepository
}

func (s *OrgBillingRepositoryPGTestSuite) SetupSuite() {
	var err error
	s.client, err = newTestClient()
	if err != nil {
		s.T().Fatal(err)
	}
	s.ctx = context.TODO()
	s.repository = postgres.NewOrgBillingRepository(s.client)
}

func (s *OrgBillingRepositoryPGTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgBillingRepositoryPGTestSuite) TearDownTest() {
	queries := []string{fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_ORGANIZATIONS)}
	if err := execQueries(s.ctx, s.client, queries); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgBillingRepositoryPGTestSuite) TestSearchSkipsDeletedOrgs() {
	execSQL(s.T(), s.ctx, s.client, `INSERT INTO organizations (name, title, avatar) VALUES ('ob-live', 'Live Org', ''), ('ob-gone', 'Gone Org', '')`)
	execSQL(s.T(), s.ctx, s.client, `UPDATE organizations SET deleted_at = now() WHERE name = 'ob-gone'`)

	got, err := s.repository.Search(s.ctx, &rql.Query{Limit: 10})
	s.Require().NoError(err)

	s.Require().Len(got.Organizations, 1)
	s.Assert().Equal("ob-live", got.Organizations[0].Name)
}

func TestOrgBillingRepositoryPG(t *testing.T) {
	suite.Run(t, new(OrgBillingRepositoryPGTestSuite))
}
