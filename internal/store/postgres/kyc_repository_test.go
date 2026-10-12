package postgres_test

import (
	"context"
	"database/sql"
	"fmt"
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

type OrgKycRepositoryTestSuite struct {
	suite.Suite
	ctx           context.Context
	client        *db.Client
	repository    *postgres.OrgKycRepository
	orgRepository *postgres.OrganizationRepository
	kycs          []kyc.KYC
	orgs          []organization.Organization
}

func (s *OrgKycRepositoryTestSuite) SetupSuite() {
	var err error

	s.client, err = newTestClient()
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
	s.kycs, err = bootstrapOrganizationKYC(s.ctx, s.client, s.orgs)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *OrgKycRepositoryTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
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

func (s *OrgKycRepositoryTestSuite) TestUpsert() {
	newOrg := organization.Organization{
		Name: "test-organization",
	}

	type testCase struct {
		Description             string
		SelectedKycInput        kyc.KYC
		ExpectedOrganizationKYC kyc.KYC
		ErrString               string
		CreateNewOrg            bool
	}

	var testCases = []testCase{
		{
			Description: "should update an organization kyc if exist",
			SelectedKycInput: kyc.KYC{
				OrgID:  s.orgs[0].ID,
				Status: true,
				Link:   "abcd",
			},
			ExpectedOrganizationKYC: kyc.KYC{
				OrgID:  s.orgs[0].ID,
				Status: true,
				Link:   "abcd",
			},
		},
		{
			Description: "should create an organization kyc if not exist",
			SelectedKycInput: kyc.KYC{
				OrgID:  newOrg.ID,
				Status: true,
				Link:   "link1",
			},
			ExpectedOrganizationKYC: kyc.KYC{
				OrgID:  newOrg.ID,
				Status: true,
				Link:   "link1",
			},
			CreateNewOrg: true,
		},
		{
			Description: "should return error if link is not given while marking kyc status true",
			SelectedKycInput: kyc.KYC{
				OrgID:  s.orgs[0].ID,
				Status: true,
				Link:   "",
			},
			ErrString: kyc.ErrKycLinkNotSet.Error(),
		},
		{
			Description: "should return error if org can't be found",
			SelectedKycInput: kyc.KYC{
				OrgID:  uuid.NewString(),
				Status: false,
			},
			ErrString: kyc.ErrOrgDoesntExist.Error(),
		},
		{
			Description: "should return error if org id empty",
			SelectedKycInput: kyc.KYC{
				OrgID:  "",
				Status: false,
			},
			ErrString: kyc.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id is not uuid",
			SelectedKycInput: kyc.KYC{
				OrgID:  "10000",
				Status: false,
			},
			ErrString: kyc.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			if tc.CreateNewOrg {
				createdOrg, err := s.orgRepository.Create(s.ctx, newOrg)
				if err != nil {
					s.T().Fatalf("failed to create an org before testing org kyc upsert, err:, %s", err.Error())
				}
				tc.SelectedKycInput.OrgID = createdOrg.ID
				tc.ExpectedOrganizationKYC.OrgID = createdOrg.ID
			}
			got, err := s.repository.Upsert(s.ctx, tc.SelectedKycInput)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedOrganizationKYC, cmpopts.IgnoreFields(kyc.KYC{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedOrganizationKYC)
			}
		})
	}
}

func (s *OrgKycRepositoryTestSuite) TestList() {
	type testCase struct {
		Description  string
		SetupFunc    func() error
		ExpectedKYCs []kyc.KYC
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description: "should return mix of organizations with and without kyc",
			SetupFunc: func() error {
				if err := s.cleanup(); err != nil {
					return err
				}

				// Create org with KYC
				org1, err := s.orgRepository.Create(s.ctx, organization.Organization{
					Name: "org-with-kyc",
				})
				if err != nil {
					return err
				}

				_, err = s.repository.Upsert(s.ctx, kyc.KYC{
					OrgID:  org1.ID,
					Status: true,
					Link:   "test-link",
				})
				if err != nil {
					return err
				}

				// Create org without KYC
				_, err = s.orgRepository.Create(s.ctx, organization.Organization{
					Name: "org-without-kyc",
				})
				return err
			},
			ExpectedKYCs: []kyc.KYC{
				{Status: false, Link: ""},         // Non-KYC org first
				{Status: true, Link: "test-link"}, // KYC org second
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			if tc.SetupFunc != nil {
				err := tc.SetupFunc()
				if err != nil {
					s.T().Fatalf("failed to setup test: %v", err)
				}
			}

			got, err := s.repository.List(s.ctx)
			if tc.ErrString != "" {
				s.Assert().EqualError(err, tc.ErrString)
				return
			}

			s.Assert().NoError(err)
			s.Assert().Equal(len(tc.ExpectedKYCs), len(got), "expected %d KYCs, got %d", len(tc.ExpectedKYCs), len(got))

			// Sort both slices by Status for consistent comparison
			sort.Slice(got, func(i, j int) bool {
				if got[i].Status != got[j].Status {
					return !got[i].Status // false comes before true
				}
				return got[i].OrgID < got[j].OrgID
			})

			// Compare each KYC record
			for i := range got {
				// Compare Status and Link only
				s.Assert().Equal(tc.ExpectedKYCs[i].Status, got[i].Status, "Status mismatch at index %d", i)
				s.Assert().Equal(tc.ExpectedKYCs[i].Link, got[i].Link, "Link mismatch at index %d", i)
			}
		})
	}
}

func (s *OrgKycRepositoryTestSuite) softDeleteKycRow(orgID string) {
	_, err := s.client.ExecContext(s.ctx, "UPDATE organizations_kyc SET deleted_at = now() WHERE org_id = $1 AND deleted_at IS NULL", orgID)
	s.Require().NoError(err)
}

func (s *OrgKycRepositoryTestSuite) insertKycRow(orgID string) error {
	_, err := s.client.ExecContext(s.ctx, "INSERT INTO organizations_kyc (org_id, status, link) VALUES ($1, false, '')", orgID)
	return err
}

func (s *OrgKycRepositoryTestSuite) TestOrgIDIsUniqueOverLiveRowsOnly() {
	orgID := s.kycs[0].OrgID

	s.Run("a live row cannot be duplicated", func() {
		s.Assert().Error(s.insertKycRow(orgID))
	})

	s.Run("a soft deleted row and a live row coexist", func() {
		s.softDeleteKycRow(orgID)
		s.Require().NoError(s.insertKycRow(orgID))
		s.Assert().Error(s.insertKycRow(orgID))
	})
}

func (s *OrgKycRepositoryTestSuite) TestGetByOrgIDSkipsSoftDeletedRow() {
	orgID := s.kycs[0].OrgID
	s.softDeleteKycRow(orgID)

	_, err := s.repository.GetByOrgID(s.ctx, orgID)
	s.Assert().ErrorIs(err, kyc.ErrNotExist)
}

func (s *OrgKycRepositoryTestSuite) deletedAtOf(orgID string) []sql.NullTime {
	var deletedAts []sql.NullTime
	err := s.client.SelectContext(s.ctx, &deletedAts, "SELECT deleted_at FROM organizations_kyc WHERE org_id = $1 ORDER BY created_at", orgID)
	s.Require().NoError(err)
	return deletedAts
}

func (s *OrgKycRepositoryTestSuite) orgWithKyc(name string) string {
	org, err := s.orgRepository.Create(s.ctx, organization.Organization{Name: name})
	s.Require().NoError(err)
	_, err = s.repository.Upsert(s.ctx, kyc.KYC{OrgID: org.ID, Status: true, Link: "abcd"})
	s.Require().NoError(err)
	return org.ID
}

func (s *OrgKycRepositoryTestSuite) TestDelete() {
	s.Run("keeps the row with deleted_at set", func() {
		orgID := s.orgWithKyc("org-delete-keeps-row")

		s.Require().NoError(s.repository.Delete(s.ctx, orgID))

		deletedAts := s.deletedAtOf(orgID)
		s.Require().Len(deletedAts, 1)
		s.Assert().True(deletedAts[0].Valid)
	})

	s.Run("a second delete keeps the first deleted_at and returns nil", func() {
		orgID := s.orgWithKyc("org-second-delete")
		s.Require().NoError(s.repository.Delete(s.ctx, orgID))
		first := s.deletedAtOf(orgID)

		s.Require().NoError(s.repository.Delete(s.ctx, orgID))

		second := s.deletedAtOf(orgID)
		s.Require().Len(second, 1)
		s.Assert().True(first[0].Time.Equal(second[0].Time))
	})

	s.Run("an org without kyc returns nil", func() {
		org, err := s.orgRepository.Create(s.ctx, organization.Organization{Name: "org-without-kyc"})
		s.Require().NoError(err)

		s.Assert().NoError(s.repository.Delete(s.ctx, org.ID))
	})

	s.Run("upsert after delete inserts a new live row", func() {
		orgID := s.orgWithKyc("org-upsert-after-delete")
		s.Require().NoError(s.repository.Delete(s.ctx, orgID))

		got, err := s.repository.Upsert(s.ctx, kyc.KYC{OrgID: orgID, Status: true, Link: "fresh-link"})

		s.Require().NoError(err)
		s.Assert().Equal("fresh-link", got.Link)
		deletedAts := s.deletedAtOf(orgID)
		s.Require().Len(deletedAts, 2)
		s.Assert().True(deletedAts[0].Valid)
		s.Assert().False(deletedAts[1].Valid)
	})

	s.Run("updating a live row leaves the deleted row untouched", func() {
		orgID := s.orgWithKyc("org-update-after-delete")
		s.Require().NoError(s.repository.Delete(s.ctx, orgID))
		_, err := s.repository.Upsert(s.ctx, kyc.KYC{OrgID: orgID, Status: true, Link: "fresh-link"})
		s.Require().NoError(err)

		_, err = s.repository.Upsert(s.ctx, kyc.KYC{OrgID: orgID, Status: true, Link: "newer-link"})
		s.Require().NoError(err)

		var links []string
		s.Require().NoError(s.client.SelectContext(s.ctx, &links, "SELECT link FROM organizations_kyc WHERE org_id = $1 AND deleted_at IS NOT NULL", orgID))
		s.Assert().Equal([]string{"abcd"}, links)
	})
}

func (s *OrgKycRepositoryTestSuite) TestListIgnoresSoftDeletedKyc() {
	s.Require().NoError(s.cleanup())
	deletedOrgID := s.orgWithKyc("org-with-deleted-kyc")
	liveOrgID := s.orgWithKyc("org-with-live-kyc")
	s.Require().NoError(s.repository.Delete(s.ctx, deletedOrgID))

	got, err := s.repository.List(s.ctx)

	s.Require().NoError(err)
	byOrg := map[string]kyc.KYC{}
	for _, k := range got {
		byOrg[k.OrgID] = k
	}
	s.Require().Len(got, 2)
	s.Assert().False(byOrg[deletedOrgID].Status)
	s.Assert().Equal("", byOrg[deletedOrgID].Link)
	s.Assert().True(byOrg[liveOrgID].Status)
	s.Assert().Equal("abcd", byOrg[liveOrgID].Link)
}

func TestOrganizationKYCRepository(t *testing.T) {
	suite.Run(t, new(OrgKycRepositoryTestSuite))
}

func (s *OrgKycRepositoryTestSuite) softDeleteOrg(orgID string) {
	_, err := s.client.ExecContext(s.ctx, "UPDATE organizations SET deleted_at = now() WHERE id = $1", orgID)
	s.Require().NoError(err)
}

func (s *OrgKycRepositoryTestSuite) kycAuditCount(orgID string) int {
	var n int
	err := s.client.QueryRowxContext(s.ctx, fmt.Sprintf(
		"SELECT count(*) FROM %s WHERE target_id = $1 AND event IN ('kyc.verified', 'kyc.unverified')", postgres.TABLE_AUDITRECORDS),
		orgID).Scan(&n)
	s.Require().NoError(err)
	return n
}

func (s *OrgKycRepositoryTestSuite) TestUpsertRefusesSoftDeletedOrg() {
	s.Run("first time insert writes nothing", func() {
		org, err := s.orgRepository.Create(s.ctx, organization.Organization{Name: "deleted-org-no-kyc"})
		s.Require().NoError(err)
		s.softDeleteOrg(org.ID)

		_, err = s.repository.Upsert(s.ctx, kyc.KYC{OrgID: org.ID, Status: true, Link: "abcd"})

		s.Assert().ErrorIs(err, kyc.ErrOrgDoesntExist)
		s.Assert().Empty(s.deletedAtOf(org.ID))
		s.Assert().Equal(0, s.kycAuditCount(org.ID))
	})

	s.Run("existing kyc row stays unchanged", func() {
		orgID := s.orgWithKyc("deleted-org-with-kyc")
		auditsBefore := s.kycAuditCount(orgID)
		s.softDeleteOrg(orgID)

		_, err := s.repository.Upsert(s.ctx, kyc.KYC{OrgID: orgID, Status: false, Link: "changed"})

		s.Assert().ErrorIs(err, kyc.ErrOrgDoesntExist)
		var status bool
		var link string
		s.Require().NoError(s.client.QueryRowxContext(s.ctx, "SELECT status, link FROM organizations_kyc WHERE org_id = $1", orgID).Scan(&status, &link))
		s.Assert().True(status)
		s.Assert().Equal("abcd", link)
		s.Assert().Equal(auditsBefore, s.kycAuditCount(orgID))
	})

	s.Run("live org still works", func() {
		orgID := s.orgWithKyc("live-org")
		got, err := s.repository.Upsert(s.ctx, kyc.KYC{OrgID: orgID, Status: true, Link: "new-link"})
		s.Require().NoError(err)
		s.Assert().Equal("new-link", got.Link)
	})
}
