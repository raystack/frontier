package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/pkg/db"
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
	relations  []relation.Relation
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
		ExpectedRelation relation.Relation
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

func (s *RelationRepositoryTestSuite) TestGetByFields() {
	type testCase struct {
		Description      string
		SelectedRelation relation.Relation
		ExpectedRelation relation.Relation
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description:      "should get a relation",
			SelectedRelation: s.relations[0],
			ExpectedRelation: s.relations[0],
		},
		{
			Description: "should return error no exist if can't found relation",
			SelectedRelation: relation.Relation{
				SubjectID:          uuid.NewString(),
				SubjectNamespaceID: uuid.NewString(),
				ObjectID:           uuid.NewString(),
				ObjectNamespaceID:  uuid.NewString(),
			},
			ErrString: relation.ErrNotExist.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByFields(s.ctx, tc.SelectedRelation)
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
		ErrString        string
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
					RoleID:    "role1",
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
			ErrString: relation.ErrInvalidDetail.Error(),
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
			ErrString: relation.ErrInvalidDetail.Error(),
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
			ErrString: relation.ErrInvalidDetail.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.RelationToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
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
		ExpectedRelations []relation.Relation
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description: "should get all relations",
			ExpectedRelations: []relation.Relation{
				{
					SubjectNamespaceID: "ns1",
					SubjectID:          "uuid1",
					ObjectNamespaceID:  "ns1",
					ObjectID:           "uuid2",
					RoleID:             "role1",
					RelationType:       relation.RelationTypes.Role,
				},
				{
					SubjectNamespaceID: "ns2",
					SubjectID:          "uuid3",
					ObjectNamespaceID:  "ns2",
					ObjectID:           "uuid4",
					RoleID:             "role2",
					RelationType:       relation.RelationTypes.Role,
				},
				{
					SubjectNamespaceID: "ns1",
					SubjectID:          "uuid1",
					ObjectNamespaceID:  "ns2",
					ObjectID:           "uuid4",
					RoleID:             "ns2",
					RelationType:       relation.RelationTypes.Namespace,
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
			if !cmp.Equal(got, tc.ExpectedRelations, cmpopts.IgnoreFields(relation.Relation{},
				"ID",
				"SubjectNamespace",
				"ObjectNamespace",
				"Role",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRelations)
			}
		})
	}
}

func (s *RelationRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description      string
		RelationToUpdate relation.Relation
		ExpectedRelation relation.Relation
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description: "should update a relation with type role",
			RelationToUpdate: relation.Relation{
				ID:                 s.relations[0].ID,
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid1",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid2",
				RoleID:             "role1",
				RelationType:       relation.RelationTypes.Role,
			},
			ExpectedRelation: relation.Relation{
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid1",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid2",
				RoleID:             "role1",
				RelationType:       relation.RelationTypes.Role,
			},
		},
		{
			Description: "should update a relation with type namespace",
			RelationToUpdate: relation.Relation{
				ID:                 s.relations[0].ID,
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid5",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid6",
				RoleID:             "ns2",
				RelationType:       relation.RelationTypes.Namespace,
			},
			ExpectedRelation: relation.Relation{
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid5",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid6",
				RoleID:             "ns2",
				RelationType:       relation.RelationTypes.Namespace,
			},
		},
		{
			Description: "should return error if subject namespace id does not exist",
			RelationToUpdate: relation.Relation{
				ID:                 s.relations[0].ID,
				SubjectNamespaceID: "ns1-random",
				SubjectID:          "uuid1",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid2",
				RoleID:             "role1",
				RelationType:       relation.RelationTypes.Role,
			},
			ErrString: relation.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if object namespace id does not exist",
			RelationToUpdate: relation.Relation{
				ID:                 s.relations[0].ID,
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid1",
				ObjectNamespaceID:  "ns1-random",
				ObjectID:           "uuid2",
				RoleID:             "role1",
				RelationType:       relation.RelationTypes.Role,
			},
			ErrString: relation.ErrInvalidDetail.Error(),
		},

		{
			Description: "should return error if role id does not exist",
			RelationToUpdate: relation.Relation{
				ID:                 s.relations[0].ID,
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid1",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid2",
				RoleID:             "role1-random",
				RelationType:       relation.RelationTypes.Role,
			},
			ErrString: relation.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if namespace id does not exist",
			RelationToUpdate: relation.Relation{
				ID:                 s.relations[0].ID,
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid1",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid2",
				RoleID:             "role1",
				RelationType:       relation.RelationTypes.Namespace,
			},
			ErrString: relation.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			RelationToUpdate: relation.Relation{
				ID:                 "some-id",
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid1",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid2",
				RoleID:             "role1",
				RelationType:       relation.RelationTypes.Namespace,
			},
			ErrString: relation.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if id not found",
			RelationToUpdate: relation.Relation{
				ID:                 uuid.NewString(),
				SubjectNamespaceID: "ns1",
				SubjectID:          "uuid1",
				ObjectNamespaceID:  "ns1",
				ObjectID:           "uuid2",
				RoleID:             "role1",
				RelationType:       relation.RelationTypes.Namespace,
			},
			ErrString: relation.ErrNotExist.Error(),
		},
		{
			Description: "should return error if subject_namespace_id, subject_id, object_namespace_id, object_id already exist",
			RelationToUpdate: relation.Relation{
				ID:                 s.relations[0].ID,
				SubjectNamespaceID: "ns2",
				SubjectID:          "uuid3",
				ObjectNamespaceID:  "ns2",
				ObjectID:           "uuid4",
				RoleID:             "role2",
				RelationType:       relation.RelationTypes.Role,
			},
			ErrString: relation.ErrConflict.Error(),
		},
		{
			Description: "should return error if id is empty",
			ErrString:   relation.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.RelationToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedRelation, cmpopts.IgnoreFields(relation.Relation{},
				"ID",
				"SubjectNamespace",
				"ObjectNamespace",
				"Role",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedRelation)
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
