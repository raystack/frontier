package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/prospect"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/log"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

const (
	defaultOffset = 0
	defaultLimit  = 50
)

type ProspectRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.ProspectRepository
	prospects  []prospect.Prospect
}

func (s *ProspectRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewProspectRepository(s.client)
}

func (s *ProspectRepositoryTestSuite) SetupTest() {
	var err error

	s.prospects, err = bootstrapProspect(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func bootstrapProspect(client *db.Client) ([]prospect.Prospect, error) {
	prospectRepository := postgres.NewProspectRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-prospect.json")
	if err != nil {
		return nil, err
	}

	var fixtureData []prospect.Prospect
	if err = json.Unmarshal(testFixtureJSON, &fixtureData); err != nil {
		return nil, err
	}

	var insertedData []prospect.Prospect
	for _, d := range fixtureData {
		domain, err := prospectRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func (s *ProspectRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ProspectRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ProspectRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_PROSPECTS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *ProspectRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description      string
		ProspectToCreate prospect.Prospect
		Expected         prospect.Prospect
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description: "should create a prospect successfully",
			ProspectToCreate: prospect.Prospect{
				Email:    "test@example.com",
				Name:     "Test User",
				Phone:    "+1234567890",
				Activity: "signup",
				Status:   prospect.Subscribed,
				Source:   "website",
				Verified: true,
				Metadata: metadata.Metadata{},
			},
			Expected: prospect.Prospect{
				Email:    "test@example.com",
				Name:     "Test User",
				Phone:    "+1234567890",
				Activity: "signup",
				Status:   prospect.Subscribed,
				Source:   "website",
				Verified: true,
				Metadata: metadata.Metadata{},
			},
		},
		{
			Description: "should return error when creating prospect with duplicate email and activity",
			ProspectToCreate: prospect.Prospect{
				Email:    "test@example.com",
				Activity: "signup",
				Status:   prospect.Subscribed,
				Metadata: metadata.Metadata{},
			},
			ErrString: prospect.ErrEmailActivityAlreadyExists.Error(),
		},
		{
			Description: "should create prospect with different activity for same email",
			ProspectToCreate: prospect.Prospect{
				Email:    "test@example.com",
				Activity: "newsletter",
				Status:   prospect.Subscribed,
				Metadata: metadata.Metadata{},
			},
			Expected: prospect.Prospect{
				Email:    "test@example.com",
				Activity: "newsletter",
				Status:   prospect.Subscribed,
				Metadata: metadata.Metadata{},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.ProspectToCreate)
			if tc.ErrString != "" {
				s.Assert().Error(err)
				s.Assert().Equal(tc.ErrString, err.Error())
				return
			} else {
				s.Assert().NoError(err)
				if diff := cmp.Diff(tc.Expected, got, cmpopts.IgnoreFields(prospect.Prospect{}, "ID", "CreatedAt", "UpdatedAt", "ChangedAt")); diff != "" {
					s.T().Fatalf("mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}

func (s *ProspectRepositoryTestSuite) TestGet() {
	type testCase struct {
		Description string
		SelectedID  string
		Expected    prospect.Prospect
		ErrString   string
	}

	var testCases = []testCase{
		{
			Description: "should get a prospect",
			SelectedID:  s.prospects[0].ID,
			Expected: prospect.Prospect{
				ID:        s.prospects[0].ID,
				Name:      s.prospects[0].Name,
				Email:     s.prospects[0].Email,
				Phone:     s.prospects[0].Phone,
				Activity:  s.prospects[0].Activity,
				Status:    s.prospects[0].Status,
				ChangedAt: s.prospects[0].ChangedAt,
				Source:    s.prospects[0].Source,
				Verified:  s.prospects[0].Verified,
				CreatedAt: s.prospects[0].CreatedAt,
				UpdatedAt: s.prospects[0].UpdatedAt,
				Metadata:  s.prospects[0].Metadata,
			},
		},
		{
			Description: "should return error no exist if can't found prospect",
			SelectedID:  uuid.NewString(),
			ErrString:   prospect.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			SelectedID:  "not-uuid",
			ErrString:   prospect.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Get(s.ctx, tc.SelectedID)
			if tc.ErrString != "" {
				s.Assert().Error(err)
				s.Assert().Equal(tc.ErrString, err.Error())
				return
			}
			if !cmp.Equal(got, tc.Expected) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.Expected)
			}
		})
	}
}

func (s *ProspectRepositoryTestSuite) TestList() {
	type testCase struct {
		Description string
		Setup       func(t *testing.T) *rql.Query
		Expected    prospect.ListProspects
		Err         error
	}

	testCases := []testCase{
		{
			Description: "should list all prospects with default pagination",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultOffset, defaultLimit, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			Expected: prospect.ListProspects{
				Prospects: s.prospects,
				Page: utils.Page{
					Limit:      defaultLimit,
					Offset:     defaultOffset,
					TotalCount: int64(len(s.prospects)),
				},
			},
		},
		{
			Description: "should return paginated results",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", 1, 1, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			Expected: prospect.ListProspects{
				Prospects: s.prospects[1:2],
				Page: utils.Page{
					Limit:      1,
					Offset:     1,
					TotalCount: int64(len(s.prospects)),
				},
			},
		},
		{
			Description: "should handle filtering",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultOffset, defaultLimit, []rql.Filter{{
					Name:     "email",
					Operator: "eq",
					Value:    "test1@example.com",
				},
				}, []rql.Sort{}, []string{})
			},
			Expected: prospect.ListProspects{
				Prospects: []prospect.Prospect{s.prospects[0]},
				Page: utils.Page{
					Limit:      defaultLimit,
					Offset:     defaultOffset,
					TotalCount: 1,
				},
			},
		},
		{
			Description: "should handle search",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("test", defaultOffset, defaultLimit, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			Expected: prospect.ListProspects{
				Prospects: []prospect.Prospect{s.prospects[0], s.prospects[1]},
				Page: utils.Page{
					Limit:      defaultLimit,
					Offset:     defaultOffset,
					TotalCount: 2,
				},
			},
		},
		{
			Description: "should handle sorting",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultOffset, defaultLimit, []rql.Filter{}, []rql.Sort{
					{Name: "created_at", Order: "desc"},
				}, []string{})
			},
			Expected: prospect.ListProspects{
				Prospects: []prospect.Prospect{s.prospects[1], s.prospects[0]},
				Page: utils.Page{
					Limit:      defaultLimit,
					Offset:     defaultOffset,
					TotalCount: 2,
				},
			},
		},
		{
			Description: "should handle groups",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultOffset, defaultLimit, []rql.Filter{}, []rql.Sort{}, []string{"activity"})
			},
			Expected: prospect.ListProspects{
				Prospects: []prospect.Prospect{s.prospects[0], s.prospects[1]},
				Page: utils.Page{
					Limit:      defaultLimit,
					Offset:     defaultOffset,
					TotalCount: 2,
				},
				Group: &utils.Group{
					Name: "activity",
					Data: []utils.GroupData{
						{
							Name:  "activity-1",
							Count: 1,
						},
						{
							Name:  "activity-2",
							Count: 1,
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			query := tc.Setup(s.T())
			got, err := s.repository.List(s.ctx, query)

			if tc.Err != nil {
				s.Assert().Error(err)
				s.Assert().Equal(tc.Err.Error(), err.Error())
				return
			}

			s.Assert().NoError(err)
			s.Assert().Equal(tc.Expected.Page.TotalCount, got.Page.TotalCount)
			s.Assert().Equal(tc.Expected.Page.Limit, got.Page.Limit)
			s.Assert().Equal(tc.Expected.Page.Offset, got.Page.Offset)

			if diff := cmp.Diff(tc.Expected.Prospects, got.Prospects, cmpopts.IgnoreFields(prospect.Prospect{},
				"ID", "CreatedAt", "UpdatedAt", "ChangedAt")); diff != "" {
				s.T().Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func (s *ProspectRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description   string
		ProspectID    string
		UpdatedFields prospect.Prospect
		Expected      prospect.Prospect
		ErrString     string
	}

	testCases := []testCase{
		{
			Description: "should update prospect successfully",
			ProspectID:  s.prospects[0].ID,
			UpdatedFields: prospect.Prospect{
				ID:       s.prospects[0].ID,
				Email:    "updated@example.com",
				Name:     "Updated Name",
				Phone:    "+9876543210",
				Activity: s.prospects[0].Activity,
				Status:   prospect.Unsubscribed,
				Source:   "mobile",
				Verified: false,
				Metadata: metadata.Metadata{"medium": "updated-value"},
			},
			Expected: prospect.Prospect{
				ID:       s.prospects[0].ID,
				Email:    "updated@example.com",
				Name:     "Updated Name",
				Phone:    "+9876543210",
				Activity: s.prospects[0].Activity,
				Status:   prospect.Unsubscribed,
				Source:   "mobile",
				Verified: false,
				Metadata: metadata.Metadata{"medium": "updated-value"},
			},
		},
		{
			Description: "should return error when updating to existing email and activity combination",
			ProspectID:  s.prospects[0].ID,
			UpdatedFields: prospect.Prospect{
				ID:       s.prospects[0].ID,
				Email:    s.prospects[1].Email,
				Activity: s.prospects[1].Activity,
				Status:   prospect.Subscribed,
				Source:   "",
				Verified: false,
				Metadata: metadata.Metadata{},
			},
			ErrString: prospect.ErrEmailActivityAlreadyExists.Error(),
		},
		{
			Description: "should allow updating email with same activity",
			ProspectID:  s.prospects[0].ID,
			UpdatedFields: prospect.Prospect{
				ID:       s.prospects[0].ID,
				Email:    "new-email@example.com",
				Activity: s.prospects[0].Activity,
				Status:   s.prospects[0].Status,
				Source:   s.prospects[0].Source,
				Verified: s.prospects[0].Verified,
				Metadata: s.prospects[0].Metadata,
			},
			Expected: prospect.Prospect{
				ID:       s.prospects[0].ID,
				Email:    "new-email@example.com",
				Activity: s.prospects[0].Activity,
				Status:   s.prospects[0].Status,
				Source:   s.prospects[0].Source,
				Verified: s.prospects[0].Verified,
				Metadata: s.prospects[0].Metadata,
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.UpdatedFields)
			if tc.ErrString != "" {
				s.Assert().Error(err)
				s.Assert().Equal(tc.ErrString, err.Error())
				return
			}

			s.Assert().NoError(err)
			if diff := cmp.Diff(tc.Expected, got, cmpopts.IgnoreFields(prospect.Prospect{},
				"CreatedAt", "UpdatedAt", "ChangedAt")); diff != "" {
				s.T().Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func (s *ProspectRepositoryTestSuite) TestDelete() {
	type testCase struct {
		Description string
		ProspectID  string
		SetupFn     func(t *testing.T)
		ValidateFn  func(t *testing.T, err error)
	}

	testCases := []testCase{
		{
			Description: "should delete prospect successfully",
			ProspectID:  s.prospects[0].ID,
			ValidateFn: func(t *testing.T, err error) {
				t.Helper()
				s.Assert().NoError(err)
				// Verify prospect is deleted
				_, err = s.repository.Get(s.ctx, "00000000-0000-0000-0000-000000000001") // hardcode the UUID which we deleted, don't use s.prospects[0].ID,
				s.Assert().Equal(prospect.ErrNotExist.Error(), err.Error())
			},
		},
		{
			Description: "should return error when deleting with invalid uuid",
			ProspectID:  "invalid-uuid",
			ValidateFn: func(t *testing.T, err error) {
				t.Helper()
				s.Assert().Error(err)
				s.Assert().Equal(prospect.ErrInvalidUUID.Error(), err.Error())
			},
		},
		{
			Description: "should be idempotent when deleting already deleted prospect",
			ProspectID:  s.prospects[0].ID,
			SetupFn: func(t *testing.T) {
				t.Helper()
				err := s.repository.Delete(s.ctx, "00000000-0000-0000-0000-000000000001") // hardcode the UUID which we deleted, don't use s.prospects[0].ID,
				s.Assert().NoError(err)
			},
			ValidateFn: func(t *testing.T, err error) {
				t.Helper()
				s.Assert().NoError(err)
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			// Reset the test data for each test case
			s.cleanup()
			s.SetupTest()

			if tc.SetupFn != nil {
				tc.SetupFn(s.T())
			}

			err := s.repository.Delete(s.ctx, tc.ProspectID)
			if tc.ValidateFn != nil {
				tc.ValidateFn(s.T(), err)
			}
		})
	}
}
func TestProspectRepository(t *testing.T) {
	suite.Run(t, new(ProspectRepositoryTestSuite))
}
