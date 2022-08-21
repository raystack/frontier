package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/pkg/db"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
)

type UserRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.UserRepository
	users      []user.User
}

func (s *UserRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewUserRepository(s.client)
}

func (s *UserRepositoryTestSuite) SetupTest() {
	var err error
	s.users, err = bootstrapUser(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *UserRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_USERS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *UserRepositoryTestSuite) TestGetByID() {
	type testCase struct {
		Description  string
		SelectedID   string
		ExpectedUser user.User
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description: "should get a user",
			SelectedID:  s.users[0].ID,
			ExpectedUser: user.User{
				ID:       s.users[0].ID,
				Name:     s.users[0].Name,
				Email:    s.users[0].Email,
				Metadata: s.users[0].Metadata,
			},
		},
		{
			Description: "should return error if id is empty",
			SelectedID:  "",
			ErrString:   user.ErrInvalidID.Error(),
		},
		{
			Description: "should return error no exist if can't found user",
			SelectedID:  uuid.NewString(),
			ErrString:   user.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			SelectedID:  "not-uuid",
			ErrString:   user.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByID(s.ctx, tc.SelectedID)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedUser, cmpopts.IgnoreFields(user.User{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUser)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestGetByEmail() {
	type testCase struct {
		Description   string
		SelectedEmail string
		ExpectedUser  user.User
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description:   "should get a user",
			SelectedEmail: "jane.dee@odpf.io",
			ExpectedUser: user.User{
				Name:  "Jane Dee",
				Email: "jane.dee@odpf.io",
				Metadata: map[string]any{
					"key-string":  "value-string",
					"key-integer": 123,
					"key-json": map[string]any{
						"k1": "v1",
						"k2": "v2",
					},
				},
			},
		},
		{
			Description:   "should return error if email is empty",
			SelectedEmail: "",
			ErrString:     user.ErrInvalidEmail.Error(),
		},
		{
			Description:   "should return error no exist if can't found user",
			SelectedEmail: "random",
			ErrString:     user.ErrNotExist.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByEmail(s.ctx, tc.SelectedEmail)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedUser, cmpopts.IgnoreFields(user.User{}, "ID", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUser)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description   string
		UserToCreate  user.User
		ExpectedEmail string
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description: "should create a user",
			UserToCreate: user.User{
				Name:  "new user",
				Email: "new.user@odpf.io",
			},
			ExpectedEmail: "new.user@odpf.io",
		},
		{
			Description: "should return error if user already exist",
			UserToCreate: user.User{
				Name:  "new user",
				Email: "new.user@odpf.io",
			},
			ErrString: user.ErrConflict.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.UserToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ExpectedEmail != "" && (got.Email != tc.ExpectedEmail) {
				s.T().Fatalf("got result %+v, expected was %+v", got.ID, tc.ExpectedEmail)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestList() {
	type testCase struct {
		Description   string
		Filter        user.Filter
		ExpectedUsers []user.User
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description:   "should get all users",
			ExpectedUsers: s.users,
		},
		{
			Description: "should return empty users if keyword not match any",
			Filter: user.Filter{
				Keyword: "some-keyword",
			},
		},
		{
			Description: "should return 1 if filter with page",
			Filter: user.Filter{
				Limit: 1,
				Page:  1,
			},
			ExpectedUsers: []user.User{
				{
					Name:  "John Doe",
					Email: "john.doe@odpf.io",
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.List(s.ctx, tc.Filter)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedUsers, cmpopts.IgnoreFields(user.User{}, "ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUsers)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestGetByIDs() {
	type testCase struct {
		Description   string
		IDs           []string
		ExpectedUsers []user.User
		ErrString     string
	}

	var testCases = []testCase{
		{
			Description:   "should get all users with ids",
			IDs:           []string{s.users[0].ID, s.users[1].ID},
			ExpectedUsers: s.users,
		},
		{
			Description: "should return empty users if ids not exist",
			IDs:         []string{uuid.NewString(), uuid.NewString()},
		},
		{
			Description:   "should return error if ids not uuid",
			IDs:           []string{"a", "b"},
			ExpectedUsers: []user.User{},
			ErrString:     user.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByIDs(s.ctx, tc.IDs)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedUsers, cmpopts.IgnoreFields(user.User{}, "ID", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUsers)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestUpdateByEmail() {
	type testCase struct {
		Description  string
		UserToUpdate user.User
		ExpectedName string
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description: "should update a user",
			UserToUpdate: user.User{
				Name:  "Doe John",
				Email: "john.doe@odpf.io",
			},
			ExpectedName: "Doe John",
		},
		{
			Description: "should return error if user not found",
			UserToUpdate: user.User{
				Email: "random@email.com",
			},
			ErrString: user.ErrNotExist.Error(),
		},
		{
			Description: "should return error if user email is empty",
			ErrString:   user.ErrInvalidEmail.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateByEmail(s.ctx, tc.UserToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ExpectedName != "" && (got.Name != tc.ExpectedName) {
				s.T().Fatalf("got result %+v, expected was %+v", got.ID, tc.ExpectedName)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestUpdateByID() {
	type testCase struct {
		Description  string
		UserToUpdate user.User
		ExpectedName string
		ErrString    string
	}

	var testCases = []testCase{
		{
			Description: "should update a user",
			UserToUpdate: user.User{
				ID:    s.users[0].ID,
				Name:  "Doe John",
				Email: "john.doe@odpf.io",
			},
			ExpectedName: "Doe John",
		},
		{
			Description: "should return error if user not found",
			UserToUpdate: user.User{
				ID:    uuid.NewString(),
				Name:  "Doe John",
				Email: "john.doe@odpf.io",
			},
			ErrString: user.ErrNotExist.Error(),
		},
		{
			Description: "should return error if user already exist",
			UserToUpdate: user.User{
				ID:    s.users[1].ID,
				Name:  "Doe John",
				Email: "john.doe@odpf.io",
			},
			ErrString: user.ErrConflict.Error(),
		},
		{
			Description: "should return error if user id is empty",
			ErrString:   user.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateByID(s.ctx, tc.UserToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if tc.ExpectedName != "" && (got.Name != tc.ExpectedName) {
				s.T().Fatalf("got result %+v, expected was %+v", got.ID, tc.ExpectedName)
			}
		})
	}
}
func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
