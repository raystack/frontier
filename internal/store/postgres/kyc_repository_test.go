package postgres_test

import (
	"context"
	"fmt"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/organization"

	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
	"testing"
)

type OrgKycRepositoryTestSuite struct {
	suite.Suite
	ctx           context.Context
	client        *db.Client
	pool          *dockertest.Pool
	resource      *dockertest.Resource
	repository    *postgres.OrgKycRepository
	orgRepository *postgres.OrganizationRepository
	kycs          []kyc.KYC
	orgs          []organization.Organization
}

func (s *OrgKycRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewOrgKycRepository(s.client)
	s.orgRepository = postgres.NewOrganizationRepository(s.client)

}

func (s *OrgKycRepositoryTestSuite) SetupTest() {
	var err error
	s.orgs, err = bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.kycs, err = bootstrapOrganizationKYC(s.client, s.orgs)
	if err != nil {
		s.T().Fatal(err)
	}

}

func (s *OrgKycRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgKycRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgKycRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_ORGANIZATIONS_KYC),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_ORGANIZATIONS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *OrgKycRepositoryTestSuite) TestGetByID() {
	type testCase struct {
		Description             string
		SelectedID              string
		ExpectedOrganizationKYC kyc.KYC
		ErrString               string
	}

	var testCases = []testCase{
		{
			Description: "should get an organization kyc",
			SelectedID:  s.kycs[0].OrgID,
			ExpectedOrganizationKYC: kyc.KYC{
				Status: true,
				Link:   "abcd",
			},
		},
		{
			Description: "should return error no exist if can't found organization kyc",
			SelectedID:  uuid.NewString(),
			ErrString:   kyc.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id empty",
			ErrString:   kyc.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			SelectedID:  "10000",
			ErrString:   kyc.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByOrgID(s.ctx, tc.SelectedID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedOrganizationKYC, cmpopts.IgnoreFields(kyc.KYC{}, "OrgID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedOrganizationKYC)
			}
		})
	}
}

func TestOrganizationKYCRepository(t *testing.T) {
	suite.Run(t, new(OrgKycRepositoryTestSuite))
}
