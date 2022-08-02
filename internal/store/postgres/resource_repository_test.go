package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/group"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/pkg/db"
	"github.com/odpf/shield/pkg/uuid"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
)

type ResourceRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.ResourceRepository
	resources  []resource.Resource
	groups     []group.Group
	projects   []project.Project
	orgs       []organization.Organization
	namespaces []namespace.Namespace
	users      []user.User
}

func (s *ResourceRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewResourceRepository(s.client)

	s.namespaces, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	s.orgs, err = bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	s.users, err = bootstrapUser(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	s.groups, err = bootstrapGroup(s.client, s.orgs)
	if err != nil {
		s.T().Fatal(err)
	}

	s.projects, err = bootstrapProject(s.client, s.orgs)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *ResourceRepositoryTestSuite) SetupTest() {
	var err error
	s.resources, err = bootstrapResource(s.client, s.groups, s.projects, s.orgs, s.namespaces, s.users)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *ResourceRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ResourceRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ResourceRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_RESOURCES),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *ResourceRepositoryTestSuite) TestGetByID() {
	type testCase struct {
		Description      string
		SelectedID       string
		ExpectedResource resource.Resource
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description: "should get a resource",
			SelectedID:  s.resources[0].Idxa,
			ExpectedResource: resource.Resource{
				Idxa:           s.resources[0].Idxa,
				URN:            s.resources[0].URN,
				Name:           s.resources[0].Name,
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
		},
		{
			Description: "should return error if id is empty",
			ErrString:   resource.ErrInvalidID.Error(),
		},
		{
			Description: "should return error no exist if can't found resource",
			SelectedID:  uuid.NewString(),
			ErrString:   resource.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			SelectedID:  "10000",
			ErrString:   resource.ErrInvalidUUID.Error(),
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
			if !cmp.Equal(got, tc.ExpectedResource, cmpopts.IgnoreFields(resource.Resource{},
				"Project",
				"Group",
				"Organization",
				"Namespace",
				"User",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedResource)
			}
		})
	}
}

func (s *ResourceRepositoryTestSuite) TestGetByURN() {
	type testCase struct {
		Description      string
		SelectedURN      string
		ExpectedResource resource.Resource
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description: "should get a resource",
			SelectedURN: s.resources[0].URN,
			ExpectedResource: resource.Resource{
				Idxa:           s.resources[0].Idxa,
				URN:            s.resources[0].URN,
				Name:           s.resources[0].Name,
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
		},
		{
			Description: "should return error if urn is empty",
			ErrString:   resource.ErrInvalidURN.Error(),
		},
		{
			Description: "should return error no exist if can't found resource",
			SelectedURN: "some-urn",
			ErrString:   resource.ErrNotExist.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetByURN(s.ctx, tc.SelectedURN)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedResource, cmpopts.IgnoreFields(resource.Resource{},
				"Project",
				"Group",
				"Organization",
				"Namespace",
				"User",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedResource)
			}
		})
	}
}

func (s *ResourceRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description      string
		ResourceToCreate resource.Resource
		ExpectedResource resource.Resource
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description: "should create a resource",
			ResourceToCreate: resource.Resource{
				URN:            "new-urn-4",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ExpectedResource: resource.Resource{
				URN:            "new-urn-4",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
		},
		{
			Description: "should return error if namespace id does not exist",
			ResourceToCreate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    "some-ns",
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if org id does not exist",
			ResourceToCreate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: uuid.NewString(),
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if org id is not uuid",
			ResourceToCreate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: "some-str",
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidUUID.Error(),
		},

		{
			Description: "should return error if group id does not exist",
			ResourceToCreate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        uuid.NewString(),
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if group id is not uuid",
			ResourceToCreate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        "some-group",
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if project id does not exist",
			ResourceToCreate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      uuid.NewString(),
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidDetail.Error(),
		},
		{
			Description: "should return error if project id is not uuid",
			ResourceToCreate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      "some-id",
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if resource urn is empty",
			ErrString:   resource.ErrInvalidURN.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.ResourceToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedResource, cmpopts.IgnoreFields(resource.Resource{},
				"Idxa",
				"Project",
				"Group",
				"Organization",
				"Namespace",
				"User",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedResource)
			}
		})
	}
}

func (s *ResourceRepositoryTestSuite) TestList() {
	type testCase struct {
		Description       string
		Filter            resource.Filter
		ExpectedResources []resource.Resource
		ErrString         string
	}

	var testCases = []testCase{
		{
			Description:       "should get all resources",
			ExpectedResources: s.resources,
		},
		{
			Description: "should get filtered resources",
			Filter: resource.Filter{
				ProjectID:      s.projects[1].ID,
				GroupID:        s.groups[1].ID,
				OrganizationID: s.orgs[1].ID,
				NamespaceID:    s.namespaces[1].ID,
			},
			ExpectedResources: []resource.Resource{
				{
					Idxa:           s.resources[1].Idxa,
					URN:            s.resources[1].URN,
					Name:           s.resources[1].Name,
					ProjectID:      s.resources[1].ProjectID,
					GroupID:        s.resources[1].GroupID,
					OrganizationID: s.resources[1].OrganizationID,
					NamespaceID:    s.resources[1].NamespaceID,
					UserID:         s.resources[1].UserID,
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
			if !cmp.Equal(got, tc.ExpectedResources, cmpopts.IgnoreFields(resource.Resource{},
				"Project",
				"Group",
				"Organization",
				"Namespace",
				"User",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedResources)
			}
		})
	}
}

func (s *ResourceRepositoryTestSuite) TestUpdate() {
	type testCase struct {
		Description      string
		ResourceID       string
		ResourceToUpdate resource.Resource
		ExpectedResource resource.Resource
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description: "should update a resource",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            "new-urn-4",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ExpectedResource: resource.Resource{
				Idxa:           s.resources[0].Idxa,
				URN:            "new-urn-4",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
		},
		{
			Description: "should return error if namespace id does not exist",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    "some-ns",
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrNotExist.Error(),
		},
		{
			Description: "should return error if org id does not exist",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: uuid.NewString(),
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrNotExist.Error(),
		},
		{
			Description: "should return error if org id is not uuid",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: "some-str",
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if group id does not exist",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        uuid.NewString(),
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrNotExist.Error(),
		},
		{
			Description: "should return error if group id is not uuid",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        "some-group",
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if project id does not exist",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      uuid.NewString(),
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrNotExist.Error(),
		},
		{
			Description: "should return error if project id is not uuid",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            "new-urn-notexist",
				Name:           "resource4",
				ProjectID:      "some-id",
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if urn already exist",
			ResourceID:  s.resources[0].Idxa,
			ResourceToUpdate: resource.Resource{
				URN:            s.resources[2].URN,
				Name:           "resource4",
				ProjectID:      s.resources[0].ProjectID,
				GroupID:        s.resources[0].GroupID,
				OrganizationID: s.resources[0].OrganizationID,
				NamespaceID:    s.resources[0].NamespaceID,
				UserID:         s.resources[0].UserID,
			},
			ErrString: resource.ErrConflict.Error(),
		},
		{
			Description: "should return error if resource id is empty",
			ErrString:   resource.ErrInvalidID.Error(),
		},
		{
			Description: "should return error if resource urn is empty",
			ResourceID:  uuid.NewString(),
			ErrString:   resource.ErrInvalidURN.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.ResourceID, tc.ResourceToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedResource, cmpopts.IgnoreFields(resource.Resource{},
				"Project",
				"Group",
				"Organization",
				"Namespace",
				"User",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedResource)
			}
		})
	}
}

func TestResourceRepository(t *testing.T) {
	suite.Run(t, new(ResourceRepositoryTestSuite))
}
