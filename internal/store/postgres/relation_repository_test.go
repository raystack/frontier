package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/goto/salt/log"
	"github.com/goto/shield/core/relation"
	"github.com/goto/shield/internal/store/postgres"
	"github.com/goto/shield/pkg/db"
	"github.com/goto/shield/pkg/errors"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
)

type RelationRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.RelationRepository
	relations  []relation.RelationV2
}

func (s *RelationRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewRelationRepository(s.client)

	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapRole(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *RelationRepositoryTestSuite) SetupTest() {
	var err error
	s.relations, err = bootstrapRelation(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *RelationRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *RelationRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *RelationRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_RELATIONS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *RelationRepositoryTestSuite) TestGet() {
	type testCase struct {
		Description      string
		SelectedID       string
		ExpectedRelation relation.RelationV2
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description:      "should get a relation",
			SelectedID:       s.relations[0].ID,
			ExpectedRelation: s.relations[0],
		},
		{
			Description: "should return error if id is empty",
			ErrString:   relation.ErrInvalidID.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			SelectedID:  "10000",
			ErrString:   relation.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error no exist if can't found relation",
			SelectedID:  uuid.NewString(),
			ErrString:   relation.ErrNotExist.Error(),
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
			if !cmp.Equal(got, tc.ExpectedRelation, cmpopts.IgnoreFields(relation.Relation{}, "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRelation)
			}
		})
	}
}

func (s *RelationRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description      string
		RelationToCreate relation.RelationV2
		ExpectedRelation relation.RelationV2
		Err              error
	}

	var testCases = []testCase{
		{
			Description: "should create a relation with type role",
			RelationToCreate: relation.RelationV2{
				Subject: relation.Subject{
					ID:        "uuid1",
					Namespace: "ns1",
					RoleID:    "role1",
				},
				Object: relation.Object{
					ID:          "uuid2",
					NamespaceID: "ns1",
				},
			},
			ExpectedRelation: relation.RelationV2{
				Subject: relation.Subject{
					ID:        "uuid1",
					Namespace: "ns1",
					RoleID:    "ns1:role1",
				},
				Object: relation.Object{
					ID:          "uuid2",
					NamespaceID: "ns1",
				},
			},
		},
		{
			Description: "should return error if subject namespace id does not exist",
			RelationToCreate: relation.RelationV2{
				Subject: relation.Subject{
					ID:        "uuid1",
					Namespace: "ns1-random",
					RoleID:    "role1",
				},
				Object: relation.Object{
					ID:          "uuid2",
					NamespaceID: "ns1",
				},
			},
			Err: relation.ErrInvalidDetail,
		},
		{
			Description: "should return error if role id does not exist",
			RelationToCreate: relation.RelationV2{
				Subject: relation.Subject{
					ID:        "uuid1",
					Namespace: "ns1",
					RoleID:    "role1-random",
				},
				Object: relation.Object{
					ID:          "uuid2",
					NamespaceID: "ns1",
				},
			},
			Err: relation.ErrInvalidDetail,
		},
		{
			Description: "should return error if object namespace id does not exist",
			RelationToCreate: relation.RelationV2{
				Subject: relation.Subject{
					ID:        "uuid1",
					Namespace: "ns1",
					RoleID:    "role1",
				},
				Object: relation.Object{
					ID:          "uuid2",
					NamespaceID: "ns10",
				},
			},
			Err: relation.ErrInvalidDetail,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.RelationToCreate)
			if tc.Err != nil {
				if errors.Is(tc.Err, err) {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.Err.Error())
				}
			}
			if !cmp.Equal(got, tc.ExpectedRelation, cmpopts.IgnoreFields(relation.RelationV2{},
				"ID",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf(cmp.Diff(got, tc.ExpectedRelation))
			}
		})
	}
}

func (s *RelationRepositoryTestSuite) TestList() {
	type testCase struct {
		Description       string
		ExpectedRelations []relation.RelationV2
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description: "should get all relations",
			ExpectedRelations: []relation.RelationV2{
				{
					Subject: relation.Subject{
						ID:        "uuid1",
						Namespace: "ns1",
						RoleID:    "ns1:role1",
					},
					Object: relation.Object{
						ID:          "uuid2",
						NamespaceID: "ns1",
					},
				},
				{
					Subject: relation.Subject{
						ID:        "uuid3",
						Namespace: "ns2",
						RoleID:    "ns2:role2",
					},
					Object: relation.Object{
						ID:          "uuid4",
						NamespaceID: "ns2",
					},
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
			if !cmp.Equal(got, tc.ExpectedRelations, cmpopts.IgnoreFields(relation.RelationV2{},
				"ID",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf(cmp.Diff(got, tc.ExpectedRelations))
			}
		})
	}
}

func (s *RelationRepositoryTestSuite) TestDeleteByID() {
	type testCase struct {
		Description string
		DeletedID   string
		ErrString   string
	}

	var testCases = []testCase{
		{
			Description: "should delete a relation",
			DeletedID:   s.relations[0].ID,
		},
		{
			Description: "should return error if id is empty",
			ErrString:   relation.ErrInvalidID.Error(),
		},
		{
			Description: "should return error if id not exist",
			DeletedID:   uuid.NewString(),
			ErrString:   relation.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			DeletedID:   "random",
			ErrString:   relation.ErrInvalidUUID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			err := s.repository.DeleteByID(s.ctx, tc.DeletedID)
			if err != nil && tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
		})
	}
}

func TestRelationRepository(t *testing.T) {
	suite.Run(t, new(RelationRepositoryTestSuite))
}
