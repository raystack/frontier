package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/log"
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
				Title:    s.users[0].Title,
				Metadata: s.users[0].Metadata,
				State:    user.Enabled,
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
			if !cmp.Equal(got, tc.ExpectedUser, cmpopts.IgnoreFields(user.User{}, "ID", "Metadata", "CreatedAt", "UpdatedAt")) {
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
			SelectedEmail: s.users[0].Email,
			ExpectedUser: user.User{
				ID:       s.users[0].ID,
				Name:     s.users[0].Name,
				Email:    s.users[0].Email,
				Title:    s.users[0].Title,
				Metadata: s.users[0].Metadata,
				State:    user.Enabled,
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
				Title: "new user",
				Email: "new.user@raystack.org",
				Name:  "test_user_slug",
				Metadata: metadata.Metadata{
					"key": "value",
				},
			},
			ExpectedEmail: "new.user@raystack.org",
		},
		{
			Description: "should return error if user already exist",
			UserToCreate: user.User{
				Title:    "new user",
				Email:    "new.user@raystack.org",
				Name:     "test_user_slug",
				Metadata: nil,
			},
			ErrString: user.ErrConflict.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.UserToCreate)
			if err != nil && tc.ErrString != "" {
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
				State: user.Enabled,
			},
			ExpectedUsers: []user.User{
				{
					Name:  s.users[0].Name,
					Email: s.users[0].Email,
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
			if !(len(got) == len(tc.ExpectedUsers)) {
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
			IDs:           []string{s.users[0].ID, s.users[0].ID},
			ExpectedUsers: []user.User{s.users[0]},
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
			if !cmp.Equal(got, tc.ExpectedUsers, cmpopts.IgnoreFields(user.User{}, "ID", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUsers)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestUpdateByEmail() {
	type testCase struct {
		Description  string
		UserToUpdate user.User
		ExpectedUser user.User
		Err          error
	}

	var testCases = []testCase{
		{
			Description: "should update a user",
			UserToUpdate: user.User{
				Title: "Doe John",
				Email: s.users[0].Email,
				Name:  s.users[0].Name,
				Metadata: metadata.Metadata{
					"label":       "Label",
					"description": "Description",
				},
				State: user.Enabled,
			},
			ExpectedUser: user.User{
				Title: "Doe John",
				Email: s.users[0].Email,
				Name:  s.users[0].Name,
				Metadata: metadata.Metadata{
					"label":       "Label",
					"description": "Description",
				},
				State: user.Enabled,
			},
		},
		{
			Description: "should return error if user not found",
			UserToUpdate: user.User{
				Email: "random@email.com",
			},
			Err: user.ErrNotExist,
		},
		{
			Description: "should return error if user email is empty",
			UserToUpdate: user.User{
				Email: "",
			},
			Err: user.ErrInvalidEmail,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateByEmail(s.ctx, tc.UserToUpdate)
			if tc.Err != nil && tc.Err.Error() != "" {
				if errors.Unwrap(err) == tc.Err {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err)
				}
			}

			// TODO(kushsharma): remove metadata field from ignore once metadata is refactored
			if !cmp.Equal(got, tc.ExpectedUser, cmpopts.IgnoreFields(user.User{},
				"ID", "CreatedAt", "UpdatedAt", "Metadata")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUser)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestUpdateByID() {
	type testCase struct {
		Description  string
		UserToUpdate user.User
		ExpectedUser user.User
		Err          error
	}

	var testCases = []testCase{
		{
			Description: "should update a user",
			UserToUpdate: user.User{
				ID:    s.users[0].ID,
				Title: "Doe John",
				Email: s.users[0].Email,
				Name:  s.users[0].Name,
				Metadata: metadata.Metadata{
					"label":       "Label",
					"description": "Description",
				},
			},
			ExpectedUser: user.User{
				ID:    s.users[0].ID,
				Title: "Doe John",
				Email: s.users[0].Email,
				Name:  s.users[0].Name,
				Metadata: metadata.Metadata{
					"label":       "Label",
					"description": "Description",
				},
				State: user.Enabled,
			},
		},
		{
			Description: "should return error if user not found",
			UserToUpdate: user.User{
				ID:    uuid.NewString(),
				Title: "Doe John",
				Email: s.users[0].Email,
				Name:  s.users[0].Name,
			},
			Err: user.ErrNotExist,
		},
		{
			Description: "should not update the user email",
			UserToUpdate: user.User{
				ID:       s.users[0].ID,
				Title:    "Doe John",
				Email:    s.users[1].Email,
				Name:     s.users[0].Name,
				Metadata: s.users[0].Metadata,
			},
			ExpectedUser: user.User{
				ID:       s.users[0].ID,
				Title:    "Doe John",
				Email:    s.users[0].Email,
				Name:     s.users[0].Name,
				Metadata: s.users[0].Metadata,
				State:    user.Enabled,
			},
		},
		{
			Description: "should return error if user id is empty",
			Err:         user.ErrInvalidID,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateByID(s.ctx, tc.UserToUpdate)
			if tc.Err != nil && tc.Err.Error() != "" {
				if errors.Unwrap(err) == tc.Err {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err)
				}
			}
			// TODO(kushsharma): remove metadata field from ignore once metadata is refactored
			if !cmp.Equal(got, tc.ExpectedUser, cmpopts.IgnoreFields(user.User{},
				"ID", "CreatedAt", "UpdatedAt", "Metadata")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUser)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestDelete() {
	type testCase struct {
		Description string
		User        string
		Err         error
	}

	var testCases = []testCase{
		{
			Description: "should delete a user",
			User:        s.users[0].ID,
		},
		{
			Description: "should return error if user not found",
			User:        uuid.NewString(),
			Err:         user.ErrNotExist,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			err := s.repository.Delete(s.ctx, tc.User)
			if tc.Err != nil && tc.Err.Error() != "" {
				if err != tc.Err {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err)
				}
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestSetState() {
	type testCase struct {
		Description string
		User        string
		State       user.State
		Err         error
	}

	var testCases = []testCase{
		{
			Description: "should set state to enabled",
			User:        s.users[0].ID,
			State:       user.Enabled,
		},
		{
			Description: "should error if user not found",
			User:        uuid.NewString(),
			State:       user.Enabled,
			Err:         user.ErrNotExist,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			err := s.repository.SetState(s.ctx, tc.User, tc.State)
			if tc.Err != nil && tc.Err.Error() != "" {
				if err != tc.Err {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err)
				}
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestGetByName() {
	type testCase struct {
		Description  string
		Name         string
		ExpectedUser user.User
		Err          error
	}

	var testCases = []testCase{
		{
			Description:  "should get a user by name",
			Name:         s.users[0].Name,
			ExpectedUser: s.users[0],
		},
		{
			Description: "should return error if user not found",
			Name:        "John Doe",
			Err:         user.ErrNotExist,
		},
		{
			Description: "should return error if name is empty",
			Name:        "",
			Err:         user.ErrMissingName,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByName(s.ctx, tc.Name)
			if tc.Err != nil && tc.Err.Error() != "" {
				if errors.Unwrap(err) == tc.Err {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err)
				}
			}
			if !cmp.Equal(got, tc.ExpectedUser) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUser)
			}
		})
	}
}

func (s *UserRepositoryTestSuite) TestUpdateByName() {
	type testCase struct {
		Description  string
		UserToUpdate user.User
		ExpectedUser user.User
		Err          error
	}

	var testCases = []testCase{
		{
			Description: "should update a user",
			UserToUpdate: user.User{
				Name: s.users[0].Name, Title: "Doe John", Email: s.users[0].Email,
			},
			ExpectedUser: user.User{
				ID:        s.users[0].ID,
				Title:     "Doe John",
				Email:     s.users[0].Email,
				Name:      s.users[0].Name,
				Metadata:  s.users[0].Metadata,
				CreatedAt: s.users[0].CreatedAt,
				State:     s.users[0].State,
			},
		},
		{
			Description:  "should return error if user not found",
			UserToUpdate: user.User{Name: "John Doe", Title: "Doe John", Email: "test@raystack.org"},
			Err:          user.ErrNotExist,
		},
		{
			Description:  "should return error if name is empty",
			UserToUpdate: user.User{Name: ""},
			Err:          user.ErrMissingName,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateByName(s.ctx, tc.UserToUpdate)
			if tc.Err != nil && tc.Err.Error() != "" {
				if errors.Unwrap(err) == tc.Err {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err)
				}
			}

			// ignore ID, UpdatedAt fields
			if diff := cmp.Diff(got, tc.ExpectedUser, cmpopts.IgnoreFields(user.User{},
				"ID", "UpdatedAt")); diff != "" {
				s.T().Errorf("mismatch (-got +want):\n%s", diff)
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedUser)
			}
		})
	}
}

func TestUserRepository(t *testing.T) {
	suite.Run(t, new(UserRepositoryTestSuite))
}
