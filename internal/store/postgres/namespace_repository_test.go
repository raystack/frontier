package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ory/dockertest"
	"github.com/raystack/salt/log"
	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/internal/store/postgres"
	"github.com/raystack/shield/pkg/db"
	"github.com/stretchr/testify/suite"
)

type NamespaceRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.NamespaceRepository
}

func (s *NamespaceRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewNamespaceRepository(s.client)
}

func (s *NamespaceRepositoryTestSuite) SetupTest() {
	var err error
	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *NamespaceRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *NamespaceRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *NamespaceRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_NAMESPACES),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *NamespaceRepositoryTestSuite) TestGet() {
	type testCase struct {
		Description       string
		SelectedID        string
		ExpectedNamespace namespace.Namespace
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description: "should get a namespace",
			SelectedID:  "ns2",
			ExpectedNamespace: namespace.Namespace{
				ID:   "ns2",
				Name: "ns2",
			},
		},
		{
			Description: "should return error no exist if can't found namespace",
			SelectedID:  "10000",
			ErrString:   namespace.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id empty",
			ErrString:   namespace.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Get(s.ctx, tc.SelectedID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got, tc.ExpectedNamespace, cmpopts.IgnoreFields(namespace.Namespace{}, "ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedNamespace)
			}
		})
	}
}

func (s *NamespaceRepositoryTestSuite) TestUpsert() {
	type testCase struct {
		Description       string
		NamespaceToCreate namespace.Namespace
		ExpectedNamespace namespace.Namespace
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description: "should create a namespace",
			NamespaceToCreate: namespace.Namespace{
				Name: "ns3",
			},
			ExpectedNamespace: namespace.Namespace{
				Name: "ns3",
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Upsert(s.ctx, tc.NamespaceToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got, tc.ExpectedNamespace, cmpopts.IgnoreFields(namespace.Namespace{}, "ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedNamespace)
			}
		})
	}
}

func (s *NamespaceRepositoryTestSuite) TestList() {
	type testCase struct {
		Description        string
		ExpectedNamespaces []namespace.Namespace
		ErrString          string
	}

	var testCases = []testCase{
		{
			Description: "should get all namespaces",
			ExpectedNamespaces: []namespace.Namespace{
				{
					ID:   "ns1",
					Name: "ns1",
				},
				{
					ID:   "ns2",
					Name: "ns2",
				},
				{
					ID:   "back1_r1",
					Name: "Back1 R1",
				},
				{
					ID:   "back1_r2",
					Name: "Back1 R2",
				},
				{
					ID:   "back2_r1",
					Name: "Back2 R1",
				},
				{
					ID:   "app/group",
					Name: "app/group",
				},
				{
					ID:   "app/project",
					Name: "app/project",
				},
				{
					ID:   "app/organization",
					Name: "app/organization",
				},
				{
					ID:   "app/user",
					Name: "app/user",
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got, tc.ExpectedNamespaces, cmpopts.IgnoreFields(namespace.Namespace{}, "ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedNamespaces)
			}
		})
	}
}

func (s *NamespaceRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description       string
		NamespaceToUpdate namespace.Namespace
		ExpectedNamespace namespace.Namespace
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description: "should return error if namespace not found",
			NamespaceToUpdate: namespace.Namespace{
				ID:   uuid.New().String(),
				Name: "not-exist",
			},
			ErrString: "namespace doesn't exist",
		},
		{
			Description: "should return error if namespace id is empty",
			ErrString:   "namespace id is invalid",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.NamespaceToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ErrString == "" {
				s.Assert().NoError(err)
			}
			if !cmp.Equal(got, tc.ExpectedNamespace, cmpopts.IgnoreFields(namespace.Namespace{}, "ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedNamespace)
			}
		})
	}
}

func TestNamespaceRepository(t *testing.T) {
	suite.Run(t, new(NamespaceRepositoryTestSuite))
}
