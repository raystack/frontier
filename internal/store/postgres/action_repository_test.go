package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/ory/dockertest"
	"github.com/raystack/salt/log"
	"github.com/raystack/shield/core/action"
	"github.com/raystack/shield/core/namespace"
	"github.com/raystack/shield/internal/store/postgres"
	"github.com/raystack/shield/pkg/db"
	"github.com/stretchr/testify/suite"
)

type ActionRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.ActionRepository
}

func (s *ActionRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewActionRepository(s.client)

	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *ActionRepositoryTestSuite) SetupTest() {
	var err error
	_, err = bootstrapAction(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *ActionRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ActionRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ActionRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_ACTIONS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *ActionRepositoryTestSuite) TestGet() {
	type testCase struct {
		Description    string
		SelectedID     string
		ExpectedAction action.Action
		ErrString      string
	}

	var testCases = []testCase{
		{
			Description: "should get an action",
			SelectedID:  "action2",
			ExpectedAction: action.Action{
				ID:          "action2",
				Name:        "action-get",
				NamespaceID: "ns1",
			},
		},
		{
			Description: "should return error no exist if can't found action",
			SelectedID:  "10000",
			ErrString:   action.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id is empty",
			ErrString:   action.ErrInvalidID.Error(),
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
			if !cmp.Equal(got, tc.ExpectedAction, cmpopts.IgnoreFields(action.Action{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedAction)
			}
		})
	}
}

func (s *ActionRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description    string
		ActionToCreate action.Action
		ExpectedAction action.Action
		ErrString      string
	}

	var testCases = []testCase{
		{
			Description: "should create an action",
			ActionToCreate: action.Action{
				ID:          "123",
				Name:        "action-123",
				NamespaceID: "ns2",
			},
			ExpectedAction: action.Action{
				ID:          "123",
				Name:        "action-123",
				NamespaceID: "ns2",
			},
		},
		{
			Description: "should return error if namespace id not exist",
			ActionToCreate: action.Action{
				ID:          "12345",
				Name:        "action-123",
				NamespaceID: "random-ns",
			},
			ErrString: namespace.ErrNotExist.Error(),
		},
		{
			Description: "should return error if action id is empty",
			ErrString:   "action id is invalid",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.ActionToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedAction, cmpopts.IgnoreFields(action.Action{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedAction)
			}
		})
	}
}

func (s *ActionRepositoryTestSuite) TestList() {
	type testCase struct {
		Description     string
		ExpectedActions []action.Action
		ErrString       string
	}

	var testCases = []testCase{
		{
			Description: "should get all actions",
			ExpectedActions: []action.Action{
				{
					ID:          "action1",
					Name:        "action-post",
					NamespaceID: "ns1",
				},
				{
					ID:          "action2",
					Name:        "action-get",
					NamespaceID: "ns1",
				},
				{
					ID:          "action3",
					Name:        "action-put",
					NamespaceID: "ns2",
				},
				{
					ID:          "action4",
					Name:        "action-delete",
					NamespaceID: "ns2",
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
			if !cmp.Equal(got, tc.ExpectedActions, cmpopts.IgnoreFields(action.Action{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedActions)
			}
		})
	}
}

func (s *ActionRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description    string
		ActionToUpdate action.Action
		ExpectedAction action.Action
		ErrString      string
	}

	var testCases = []testCase{
		{
			Description: "should update an action",
			ActionToUpdate: action.Action{
				ID:          "action2",
				Name:        "action-get-updated",
				NamespaceID: "ns2",
			},
			ExpectedAction: action.Action{
				ID:          "action2",
				Name:        "action-get-updated",
				NamespaceID: "ns2",
			},
		},
		{
			Description: "should return error if namespace id does not exist",
			ActionToUpdate: action.Action{
				ID:          "action2",
				Name:        "action-get-updated",
				NamespaceID: "random-ns2",
			},
			ErrString: namespace.ErrNotExist.Error(),
		},
		{
			Description: "should return error if action not found",
			ActionToUpdate: action.Action{
				ID:   "123131",
				Name: "not-exist",
			},
			ErrString: "action doesn't exist",
		},
		{
			Description: "should return error if action id is empty",
			ErrString:   "action id is invalid",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.ActionToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedAction, cmpopts.IgnoreFields(action.Action{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedAction)
			}
		})
	}
}

func TestActionRepository(t *testing.T) {
	suite.Run(t, new(ActionRepositoryTestSuite))
}
