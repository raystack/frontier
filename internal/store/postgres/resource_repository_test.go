package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/stretchr/testify/suite"
)

type ResourceRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	repository *postgres.ResourceRepository
	resources  []resource.Resource
	projects   []project.Project
	orgs       []organization.Organization
	namespaces []namespace.Namespace
	users      []user.User
}

func (s *ResourceRepositoryTestSuite) SetupSuite() {
	var err error

	s.client, err = newTestClient()
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

	s.projects, err = bootstrapProject(s.client, s.orgs)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *ResourceRepositoryTestSuite) SetupTest() {
	var err error
	s.resources, err = bootstrapResource(s.client, s.projects, s.namespaces, s.users)
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *ResourceRepositoryTestSuite) TearDownSuite() {
	if err := closeTestClient(s.client); err != nil {
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
			SelectedID:  s.resources[0].ID,
			ExpectedResource: resource.Resource{
				ID:            s.resources[0].ID,
				URN:           s.resources[0].URN,
				Name:          s.resources[0].Name,
				ProjectID:     s.resources[0].ProjectID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
		},
		{
			Description: "should return error if id is empty",
			ErrString:   resource.ErrInvalidID.Error(),
		},
		{
			Description: "should return error no exist if can't found resource",
			SelectedID:  utils.NewString(),
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
				ID:            s.resources[0].ID,
				URN:           s.resources[0].URN,
				Name:          s.resources[0].Name,
				ProjectID:     s.resources[0].ProjectID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
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
		ErrString        error
	}

	var testCases = []testCase{
		{
			Description: "should create a resource",
			ResourceToCreate: resource.Resource{
				URN:           "new-urn-4",
				Name:          "resource4",
				ProjectID:     s.resources[0].ProjectID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ExpectedResource: resource.Resource{
				URN:           "new-urn-4",
				Name:          "resource4",
				ProjectID:     s.resources[0].ProjectID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
		},
		{
			Description: "should return error if namespace id does not exist",
			ResourceToCreate: resource.Resource{
				URN:           "new-urn-notexist",
				Name:          "resource4",
				ProjectID:     s.resources[0].ProjectID,
				NamespaceID:   "some-ns",
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ErrString: resource.ErrInvalidDetail,
		},
		{
			Description: "should return error if project id does not exist",
			ResourceToCreate: resource.Resource{
				URN:           "new-urn-notexist",
				Name:          "resource4",
				ProjectID:     utils.NewString(),
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ErrString: resource.ErrInvalidDetail,
		},
		{
			Description: "should return error if project id is not uuid",
			ResourceToCreate: resource.Resource{
				URN:           "new-urn-notexist",
				Name:          "resource4",
				ProjectID:     "some-id",
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ErrString: resource.ErrInvalidUUID,
		},
		{
			Description: "should create an org resource that has no project",
			ResourceToCreate: resource.Resource{
				URN:           "new-urn-org-1",
				Name:          "store1",
				OrgID:         s.orgs[0].ID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ExpectedResource: resource.Resource{
				URN:           "new-urn-org-1",
				Name:          "store1",
				OrgID:         s.orgs[0].ID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
		},
		{
			Description: "should create a project resource that also has its org",
			ResourceToCreate: resource.Resource{
				URN:           "new-urn-5",
				Name:          "resource5",
				ProjectID:     s.projects[0].ID,
				OrgID:         s.projects[0].Organization.ID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ExpectedResource: resource.Resource{
				URN:           "new-urn-5",
				Name:          "resource5",
				ProjectID:     s.projects[0].ID,
				OrgID:         s.projects[0].Organization.ID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
		},
		{
			Description: "should return error if org id does not exist",
			ResourceToCreate: resource.Resource{
				URN:           "new-urn-notexist",
				Name:          "store1",
				OrgID:         utils.NewString(),
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ErrString: resource.ErrInvalidDetail,
		},
		{
			Description: "should return error if the resource has no project and no org",
			ResourceToCreate: resource.Resource{
				URN:           "new-urn-notexist",
				Name:          "resource4",
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ErrString: resource.ErrInvalidDetail,
		},
		{
			Description: "should return error if resource urn is empty",
			ErrString:   resource.ErrInvalidURN,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.ResourceToCreate)
			if tc.ErrString != nil {
				if !errors.Is(err, tc.ErrString) {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedResource, cmpopts.IgnoreFields(resource.Resource{},
				"ID",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedResource)
			}
		})
	}

	s.Run("should update the live resource that already has the urn", func() {
		existing := s.resources[0]
		got, err := s.repository.Create(s.ctx, resource.Resource{
			URN:           existing.URN,
			Name:          "renamed",
			ProjectID:     existing.ProjectID,
			NamespaceID:   existing.NamespaceID,
			PrincipalID:   existing.PrincipalID,
			PrincipalType: existing.PrincipalType,
		})
		if err != nil {
			s.T().Fatal(err)
		}
		if got.ID != existing.ID || got.Name != "renamed" {
			s.T().Fatalf("got %+v, expected row %s renamed in place", got, existing.ID)
		}
	})

	s.Run("should set the org on the live resource that already has the urn", func() {
		existing := s.resources[0]
		orgID := s.projects[0].Organization.ID
		got, err := s.repository.Create(s.ctx, resource.Resource{
			URN:           existing.URN,
			Name:          existing.Name,
			ProjectID:     existing.ProjectID,
			OrgID:         orgID,
			NamespaceID:   existing.NamespaceID,
			PrincipalID:   existing.PrincipalID,
			PrincipalType: existing.PrincipalType,
		})
		if err != nil {
			s.T().Fatal(err)
		}
		if got.ID != existing.ID || got.OrgID != orgID {
			s.T().Fatalf("got %+v, expected row %s with org %s", got, existing.ID, orgID)
		}
	})

	orgResource := func(urn, name string) resource.Resource {
		return resource.Resource{
			URN:           urn,
			Name:          name,
			OrgID:         s.orgs[0].ID,
			NamespaceID:   s.resources[0].NamespaceID,
			PrincipalID:   s.resources[0].PrincipalID,
			PrincipalType: s.resources[0].PrincipalType,
		}
	}
	projectResource := func(urn string, project project.Project) resource.Resource {
		return resource.Resource{
			URN:           urn,
			Name:          "item1",
			ProjectID:     project.ID,
			OrgID:         project.Organization.ID,
			NamespaceID:   s.resources[0].NamespaceID,
			PrincipalID:   s.resources[0].PrincipalID,
			PrincipalType: s.resources[0].PrincipalType,
		}
	}

	s.Run("should update the live org resource that already has the urn", func() {
		created, err := s.repository.Create(s.ctx, orgResource("org-resource-again", "store1"))
		s.Require().NoError(err)
		got, err := s.repository.Create(s.ctx, orgResource("org-resource-again", "renamed"))
		s.Require().NoError(err)
		s.Assert().Equal(created.ID, got.ID)
		s.Assert().Equal("renamed", got.Name)
	})

	s.Run("should return conflict when the urn belongs to a live resource of another project", func() {
		existing := s.resources[0]
		var other project.Project
		for _, p := range s.projects {
			if p.ID != existing.ProjectID {
				other = p
				break
			}
		}
		_, err := s.repository.Create(s.ctx, projectResource(existing.URN, other))
		s.Assert().ErrorIs(err, resource.ErrConflict)

		got, err := s.repository.GetByID(s.ctx, existing.ID)
		s.Require().NoError(err)
		s.Assert().Equal(existing.ProjectID, got.ProjectID)
	})

	s.Run("should return conflict when a project resource and an org resource share the urn", func() {
		orgOwned, err := s.repository.Create(s.ctx, orgResource("shared-by-org", "store1"))
		s.Require().NoError(err)
		_, err = s.repository.Create(s.ctx, projectResource("shared-by-org", s.projects[0]))
		s.Assert().ErrorIs(err, resource.ErrConflict)

		projectOwned, err := s.repository.Create(s.ctx, projectResource("shared-by-project", s.projects[0]))
		s.Require().NoError(err)
		_, err = s.repository.Create(s.ctx, orgResource("shared-by-project", "store1"))
		s.Assert().ErrorIs(err, resource.ErrConflict)

		got, err := s.repository.GetByID(s.ctx, orgOwned.ID)
		s.Require().NoError(err)
		s.Assert().Empty(got.ProjectID)
		got, err = s.repository.GetByID(s.ctx, projectOwned.ID)
		s.Require().NoError(err)
		s.Assert().Equal(s.projects[0].ID, got.ProjectID)
	})

	s.Run("should create a new row when the urn belongs to a soft-deleted resource", func() {
		deleted := s.resources[1]
		if _, err := s.client.ExecContext(s.ctx, "UPDATE resources SET deleted_at = now() WHERE id = $1", deleted.ID); err != nil {
			s.T().Fatal(err)
		}
		got, err := s.repository.Create(s.ctx, resource.Resource{
			URN:           deleted.URN,
			Name:          deleted.Name,
			ProjectID:     deleted.ProjectID,
			NamespaceID:   deleted.NamespaceID,
			PrincipalID:   deleted.PrincipalID,
			PrincipalType: deleted.PrincipalType,
		})
		if err != nil {
			s.T().Fatal(err)
		}
		if got.ID == deleted.ID {
			s.T().Fatalf("got the deleted row %s back, expected a new row", deleted.ID)
		}
		var rows int
		if err := s.client.QueryRowxContext(s.ctx, "SELECT count(*) FROM resources WHERE urn = $1", deleted.URN).Scan(&rows); err != nil {
			s.T().Fatal(err)
		}
		if rows != 2 {
			s.T().Fatalf("got %d rows with urn %s, expected the deleted row and the new one", rows, deleted.URN)
		}
	})

	s.Run("should return conflict when the id belongs to a soft-deleted resource", func() {
		deleted := s.resources[2]
		if _, err := s.client.ExecContext(s.ctx, "UPDATE resources SET deleted_at = now() WHERE id = $1", deleted.ID); err != nil {
			s.T().Fatal(err)
		}
		_, err := s.repository.Create(s.ctx, resource.Resource{
			ID:            deleted.ID,
			URN:           "another-urn",
			Name:          deleted.Name,
			ProjectID:     deleted.ProjectID,
			NamespaceID:   deleted.NamespaceID,
			PrincipalID:   deleted.PrincipalID,
			PrincipalType: deleted.PrincipalType,
		})
		if !errors.Is(err, resource.ErrConflict) {
			s.T().Fatalf("got error %v, expected %v", err, resource.ErrConflict)
		}
	})
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
				ProjectID: s.projects[1].ID,
			},
			ExpectedResources: []resource.Resource{
				{
					ID:            s.resources[1].ID,
					URN:           s.resources[1].URN,
					Name:          s.resources[1].Name,
					ProjectID:     s.resources[1].ProjectID,
					NamespaceID:   s.resources[1].NamespaceID,
					PrincipalID:   s.resources[1].PrincipalID,
					PrincipalType: s.resources[1].PrincipalType,
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
			ResourceID:  s.resources[0].ID,
			ResourceToUpdate: resource.Resource{
				ID:          s.resources[0].ID,
				Name:        "resource-1",
				ProjectID:   s.resources[0].ProjectID,
				NamespaceID: s.resources[0].NamespaceID,
			},
			ExpectedResource: resource.Resource{
				ID:            s.resources[0].ID,
				URN:           "resource-1-urn",
				Name:          "resource-1",
				ProjectID:     s.resources[0].ProjectID,
				NamespaceID:   s.resources[0].NamespaceID,
				PrincipalID:   s.resources[0].PrincipalID,
				PrincipalType: s.resources[0].PrincipalType,
			},
			ErrString: "",
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Update(s.ctx, tc.ResourceToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedResource, cmpopts.IgnoreFields(resource.Resource{},
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedResource)
			}
		})
	}

	s.Run("should move updated_at forward", func() {
		before, err := s.repository.GetByID(s.ctx, s.resources[0].ID)
		if err != nil {
			s.T().Fatal(err)
		}
		got, err := s.repository.Update(s.ctx, resource.Resource{
			ID:    before.ID,
			Title: before.Title,
		})
		if err != nil {
			s.T().Fatal(err)
		}
		if !got.UpdatedAt.After(before.UpdatedAt) {
			s.T().Fatalf("got updated_at %s, expected it to be after %s", got.UpdatedAt, before.UpdatedAt)
		}
	})
}

func (s *ResourceRepositoryTestSuite) TestDelete() {
	target := s.resources[0]

	err := s.repository.Delete(s.ctx, target.ID)
	s.Require().NoError(err)

	var deleted bool
	err = s.client.QueryRowxContext(s.ctx, "SELECT deleted_at IS NOT NULL FROM resources WHERE id = $1", target.ID).Scan(&deleted)
	s.Require().NoError(err)
	s.Assert().True(deleted)

	_, err = s.repository.GetByID(s.ctx, target.ID)
	s.Assert().ErrorIs(err, resource.ErrNotExist)

	err = s.repository.Delete(s.ctx, target.ID)
	s.Assert().ErrorIs(err, resource.ErrNotExist)

	err = s.repository.Delete(s.ctx, utils.NewString())
	s.Assert().ErrorIs(err, resource.ErrNotExist)
}

func TestResourceRepository(t *testing.T) {
	suite.Run(t, new(ResourceRepositoryTestSuite))
}

func (s *ResourceRepositoryTestSuite) TestSkipsSoftDeletedResources() {
	deleted := s.resources[0]
	before, err := s.repository.List(s.ctx, resource.Filter{})
	s.Require().NoError(err)

	_, err = s.client.ExecContext(s.ctx, "UPDATE resources SET deleted_at = now() WHERE id = $1", deleted.ID)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = s.repository.GetByID(s.ctx, deleted.ID)
	s.Assert().ErrorIs(err, resource.ErrNotExist)

	_, err = s.repository.GetByURN(s.ctx, deleted.URN)
	s.Assert().ErrorIs(err, resource.ErrNotExist)

	got, err := s.repository.List(s.ctx, resource.Filter{})
	s.Assert().NoError(err)
	s.Assert().Len(got, len(before)-1)
	for _, r := range got {
		s.Assert().NotEqual(deleted.ID, r.ID)
	}

	updated := deleted
	updated.Title = "changed"
	_, err = s.repository.Update(s.ctx, updated)
	s.Assert().ErrorIs(err, resource.ErrNotExist)

	// the id of a deleted row is still taken; a create that reuses it with a new URN must conflict
	_, err = s.repository.Create(s.ctx, resource.Resource{
		ID:            deleted.ID,
		URN:           "urn-reusing-a-deleted-id",
		Name:          "reused-id",
		ProjectID:     deleted.ProjectID,
		NamespaceID:   deleted.NamespaceID,
		PrincipalID:   deleted.PrincipalID,
		PrincipalType: deleted.PrincipalType,
	})
	s.Assert().ErrorIs(err, resource.ErrConflict)
}

func (s *ResourceRepositoryTestSuite) TestReadsResourceWithoutProject() {
	created, err := s.repository.Create(s.ctx, resource.Resource{
		URN:           "org-resource-urn",
		Name:          "store1",
		OrgID:         s.orgs[0].ID,
		NamespaceID:   s.resources[0].NamespaceID,
		PrincipalID:   s.resources[0].PrincipalID,
		PrincipalType: s.resources[0].PrincipalType,
	})
	s.Require().NoError(err)
	ignoreTimes := cmpopts.IgnoreFields(resource.Resource{}, "CreatedAt", "UpdatedAt")

	s.Run("by id", func() {
		got, err := s.repository.GetByID(s.ctx, created.ID)
		s.Require().NoError(err)
		s.Assert().True(cmp.Equal(got, created, ignoreTimes), cmp.Diff(got, created, ignoreTimes))
	})

	s.Run("by urn", func() {
		got, err := s.repository.GetByURN(s.ctx, created.URN)
		s.Require().NoError(err)
		s.Assert().True(cmp.Equal(got, created, ignoreTimes), cmp.Diff(got, created, ignoreTimes))
	})

	s.Run("in a list", func() {
		got, err := s.repository.List(s.ctx, resource.Filter{})
		s.Require().NoError(err)
		s.Require().Len(got, len(s.resources)+1)
		listed := got[len(got)-1]
		s.Assert().True(cmp.Equal(listed, created, ignoreTimes), cmp.Diff(listed, created, ignoreTimes))
	})
}

func (s *ResourceRepositoryTestSuite) TestListByOrg() {
	create := func(urn, orgID, projectID string) resource.Resource {
		created, err := s.repository.Create(s.ctx, resource.Resource{
			URN:           urn,
			Name:          urn,
			OrgID:         orgID,
			ProjectID:     projectID,
			NamespaceID:   s.resources[0].NamespaceID,
			PrincipalID:   s.resources[0].PrincipalID,
			PrincipalType: s.resources[0].PrincipalType,
		})
		s.Require().NoError(err)
		return created
	}
	live := create("org-live", s.orgs[0].ID, "")
	deleted := create("org-deleted", s.orgs[0].ID, "")
	s.Require().NoError(s.repository.Delete(s.ctx, deleted.ID))
	create("other-org", s.orgs[1].ID, "")
	create("in-a-project", s.orgs[0].ID, s.projects[0].ID)

	got, err := s.repository.List(s.ctx, resource.Filter{OrgID: s.orgs[0].ID})
	s.Require().NoError(err)
	s.Require().Len(got, 1)
	s.Assert().Equal(live.ID, got[0].ID)
}
