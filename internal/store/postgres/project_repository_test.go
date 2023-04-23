package postgres_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/google/uuid"
	"github.com/odpf/salt/log"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/internal/schema"
	"github.com/odpf/shield/internal/store/postgres"
	"github.com/odpf/shield/pkg/db"
	"github.com/ory/dockertest"
	"github.com/stretchr/testify/suite"
)

type ProjectRepositoryTestSuite struct {
	suite.Suite
	ctx                 context.Context
	client              *db.Client
	pool                *dockertest.Pool
	resource            *dockertest.Resource
	repository          *postgres.ProjectRepository
	relationRepository  *postgres.RelationRepository
	namespaceRepository *postgres.NamespaceRepository
	roleRepository      *postgres.RoleRepository
	projects            []project.Project
	orgs                []organization.Organization
	users               []user.User
}

func (s *ProjectRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewProjectRepository(s.client)

	_, err = bootstrapMetadataKeys(s.client)
	if err != nil {
		s.T().Fatal(err)
	}
	s.users, err = bootstrapUser(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	s.orgs, err = bootstrapOrganization(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	s.relationRepository = postgres.NewRelationRepository(s.client)
	s.namespaceRepository = postgres.NewNamespaceRepository(s.client)
	s.roleRepository = postgres.NewRoleRepository(s.client)
}

func (s *ProjectRepositoryTestSuite) SetupTest() {
	var err error
	s.projects, err = bootstrapProject(s.client, s.orgs)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapNamespace(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapAction(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = bootstrapRole(s.client)
	if err != nil {
		s.T().Fatal(err)
	}

	_, err = s.relationRepository.Create(context.Background(), relation.RelationV2{
		Subject: relation.Subject{
			ID:        s.users[0].ID,
			Namespace: schema.UserPrincipal,
			RoleID:    schema.OwnerRole,
		},
		Object: relation.Object{
			ID:        s.projects[0].ID,
			Namespace: schema.ProjectNamespace,
		},
	})
	if err != nil {
		s.T().Fatal(err)
	}
}

func (s *ProjectRepositoryTestSuite) TearDownSuite() {
	// Clean tests
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ProjectRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *ProjectRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_PROJECTS),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_RELATIONS),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_ROLES),
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_NAMESPACES),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *ProjectRepositoryTestSuite) TestGetByID() {
	type testCase struct {
		Description     string
		SelectedID      string
		ExpectedProject project.Project
		ErrString       string
	}

	var testCases = []testCase{
		{
			Description: "should get a project",
			SelectedID:  s.projects[0].ID,
			ExpectedProject: project.Project{
				Name: "project1",
				Slug: "project-1",
				Organization: organization.Organization{
					ID: s.projects[0].ID,
				},
			},
		},
		{
			Description: "should return error no exist if can't found project",
			SelectedID:  uuid.NewString(),
			ErrString:   project.ErrNotExist.Error(),
		},
		{
			Description: "should return error if id empty",
			ErrString:   project.ErrInvalidID.Error(),
		},
		{
			Description: "should return error if id is not uuid",
			SelectedID:  "10000",
			ErrString:   project.ErrInvalidUUID.Error(),
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
			if !cmp.Equal(got, tc.ExpectedProject, cmpopts.IgnoreFields(project.Project{},
				"ID",
				"Organization", // TODO need to do deeper comparison
				"Metadata",
				"CreatedAt",
				"UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedProject)
			}
		})
	}
}

func (s *ProjectRepositoryTestSuite) TestGetBySlug() {
	type testCase struct {
		Description     string
		SelectedSlug    string
		ExpectedProject project.Project
		ErrString       string
	}

	var testCases = []testCase{
		{
			Description:  "should get a project",
			SelectedSlug: "project-1",
			ExpectedProject: project.Project{
				Name: "project1",
				Slug: "project-1",
			},
		},
		{
			Description:  "should return error no exist if can't found project",
			SelectedSlug: "randomslug",
			ErrString:    project.ErrNotExist.Error(),
		},
		{
			Description: "should return error if slug empty",
			ErrString:   project.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.GetBySlug(s.ctx, tc.SelectedSlug)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedProject, cmpopts.IgnoreFields(project.Project{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedProject)
			}
		})
	}
}

func (s *ProjectRepositoryTestSuite) TestCreate() {
	type testCase struct {
		Description     string
		ProjectToCreate project.Project
		ExpectedProject project.Project
		ErrString       string
	}

	var testCases = []testCase{
		{
			Description: "should create a project",
			ProjectToCreate: project.Project{
				Name: "new-project",
				Slug: "new-project-slug",
				Organization: organization.Organization{
					ID: s.orgs[0].ID,
				},
			},
			ExpectedProject: project.Project{
				Name: "new-project",
				Slug: "new-project-slug",
			},
		},
		{
			Description: "should return error if project slug already exist",
			ProjectToCreate: project.Project{
				Name: "newslug",
				Slug: "project-2",
				Organization: organization.Organization{
					ID: s.orgs[0].ID,
				},
			},
			ErrString: project.ErrConflict.Error(),
		},
		{
			Description: "should return error if org id not an uuid",
			ProjectToCreate: project.Project{
				Name: "newslug",
				Slug: "projectnewslug",
				Organization: organization.Organization{
					ID: "someid",
				},
			},
			ErrString: organization.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id does not exist",
			ProjectToCreate: project.Project{
				Name: "newslug",
				Slug: "projectnewslug",
				Organization: organization.Organization{
					ID: uuid.NewString(),
				},
			},
			ErrString: project.ErrNotExist.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.Create(s.ctx, tc.ProjectToCreate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedProject, cmpopts.IgnoreFields(project.Project{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedProject)
			}
		})
	}
}

func (s *ProjectRepositoryTestSuite) TestList() {
	type testCase struct {
		Description      string
		ExpectedProjects []project.Project
		ErrString        string
	}

	var testCases = []testCase{
		{
			Description: "should get all projects",
			ExpectedProjects: []project.Project{
				{
					Name: "project1",
					Slug: "project-1",
				},
				{
					Name: "project2",
					Slug: "project-2",
				},
				{
					Name: "project3",
					Slug: "project-3",
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
			if !cmp.Equal(got, tc.ExpectedProjects, cmpopts.IgnoreFields(project.Project{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedProjects)
			}
		})
	}
}

func (s *ProjectRepositoryTestSuite) TestUpdateByID() {
	type testCase struct {
		Description     string
		ProjectToUpdate project.Project
		ExpectedProject project.Project
		ErrString       string
	}

	var testCases = []testCase{
		{
			Description: "should update a project",
			ProjectToUpdate: project.Project{
				ID:   s.projects[0].ID,
				Name: "new project update",
				Slug: "new-project-update",
				Organization: organization.Organization{
					ID: s.orgs[0].ID,
				},
			},
			ExpectedProject: project.Project{
				Name: "new project update",
				Slug: "new-project-update",
			},
		},
		{
			Description: "should return error if project slug already exist",
			ProjectToUpdate: project.Project{
				ID:   s.projects[0].ID,
				Name: "new-project-2",
				Slug: "project-2",
				Organization: organization.Organization{
					ID: s.orgs[0].ID,
				},
			},
			ErrString: project.ErrConflict.Error(),
		},
		{
			Description: "should return error if project not found",
			ProjectToUpdate: project.Project{
				ID:   uuid.NewString(),
				Name: "not-exist",
				Slug: "some-slug",
				Organization: organization.Organization{
					ID: s.orgs[0].ID,
				},
			},
			ErrString: project.ErrNotExist.Error(),
		},
		{
			Description: "should return error if project id is not uuid",
			ProjectToUpdate: project.Project{
				ID:   "12345",
				Name: "not-exist",
				Slug: "some-slug",
				Organization: organization.Organization{
					ID: s.orgs[0].ID,
				},
			},
			ErrString: project.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id is not uuid",
			ProjectToUpdate: project.Project{
				ID:   s.projects[0].ID,
				Slug: "new-prj",
				Name: "not-exist",
				Organization: organization.Organization{
					ID: "not-uuid",
				},
			},
			ErrString: project.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id not exist",
			ProjectToUpdate: project.Project{
				ID:   s.projects[0].ID,
				Slug: "new-prj",
				Name: "not-exist",
				Organization: organization.Organization{
					ID: uuid.NewString(),
				},
			},
			ErrString: organization.ErrNotExist.Error(),
		},
		{
			Description: "should return error if project id is empty",
			ErrString:   project.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateByID(s.ctx, tc.ProjectToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedProject, cmpopts.IgnoreFields(project.Project{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedProject)
			}
		})
	}
}

func (s *ProjectRepositoryTestSuite) TestUpdateBySlug() {
	type testCase struct {
		Description     string
		ProjectToUpdate project.Project
		ExpectedProject project.Project
		ErrString       string
	}

	var testCases = []testCase{
		{
			Description: "should update a project",
			ProjectToUpdate: project.Project{
				Name: "new project update",
				Slug: "project-1",
				Organization: organization.Organization{
					ID: s.orgs[0].ID,
				},
			},
			ExpectedProject: project.Project{
				Name: "new project update",
				Slug: "project-1",
			},
		},
		{
			Description: "should return error if project not found",
			ProjectToUpdate: project.Project{
				Slug: "slug",
				Name: "not-exist",
				Organization: organization.Organization{
					ID: s.orgs[0].ID,
				},
			},
			ErrString: project.ErrNotExist.Error(),
		},
		{
			Description: "should return error if org id is not uuid",
			ProjectToUpdate: project.Project{
				Slug: "project-1",
				Name: "not-exist",
				Organization: organization.Organization{
					ID: "not-uuid",
				},
			},
			ErrString: organization.ErrInvalidUUID.Error(),
		},
		{
			Description: "should return error if org id not exist",
			ProjectToUpdate: project.Project{
				Slug: "project-1",
				Name: "not-exist",
				Organization: organization.Organization{
					ID: uuid.NewString(),
				},
			},
			ErrString: organization.ErrNotExist.Error(),
		},
		{
			Description: "should return error if project slug is empty",
			ErrString:   project.ErrInvalidID.Error(),
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			got, err := s.repository.UpdateBySlug(s.ctx, tc.ProjectToUpdate)
			if tc.ErrString != "" {
				if err.Error() != tc.ErrString {
					s.T().Fatalf("got error %s, expected was %s", err.Error(), tc.ErrString)
				}
			}
			if !cmp.Equal(got, tc.ExpectedProject, cmpopts.IgnoreFields(project.Project{}, "ID", "Organization", "Metadata", "CreatedAt", "UpdatedAt")) {
				s.T().Fatalf("got result %+v, expected was %+v", got, tc.ExpectedProject)
			}
		})
	}
}

func TestProjectRepository(t *testing.T) {
	suite.Run(t, new(ProjectRepositoryTestSuite))
}
