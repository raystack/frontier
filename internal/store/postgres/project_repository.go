package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/user"
	"github.com/odpf/shield/pkg/db"
)

type ProjectRepository struct {
	dbc *db.Client
}

func NewProjectRepository(dbc *db.Client) *ProjectRepository {
	return &ProjectRepository{
		dbc: dbc,
	}
}

func (r ProjectRepository) Get(ctx context.Context, id string) (project.Project, error) {
	var fetchedProject Project
	var getProjectsQuery string
	var err error
	id = strings.TrimSpace(id)
	isUuid := isUUID(id)

	//TODO needs to decouple this to make this cleaner
	if isUuid {
		getProjectsQuery, err = buildGetProjectsByIDQuery(dialect)
	} else {
		getProjectsQuery, err = buildGetProjectsBySlugQuery(dialect)
	}
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &fetchedProject, getProjectsQuery, id, id)
		})
	} else {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &fetchedProject, getProjectsQuery, id)
		})
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return project.Project{}, project.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return project.Project{}, project.ErrInvalidUUID
		}
		return project.Project{}, err
	}

	transformedProject, err := transformToProject(fetchedProject)
	if err != nil {
		return project.Project{}, err
	}

	return transformedProject, nil
}

func (r ProjectRepository) Create(ctx context.Context, projectToCreate project.Project) (project.Project, error) {
	marshaledMetadata, err := json.Marshal(projectToCreate.Metadata)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newProject Project
	createProjectQuery, err := buildCreateProjectQuery(dialect)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &newProject, createProjectQuery, projectToCreate.Name, projectToCreate.Slug, projectToCreate.Organization.ID, marshaledMetadata)
	}); err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedProj, err := transformToProject(newProject)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedProj, nil
}

func (r ProjectRepository) List(ctx context.Context) ([]project.Project, error) {
	var fetchedProjects []Project
	listProjectQuery, err := buildListProjectQuery(dialect)
	if err != nil {
		return []project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedProjects, listProjectQuery)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []project.Project{}, project.ErrNotExist
		}
		return []project.Project{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedProjects []project.Project
	for _, p := range fetchedProjects {
		transformedProj, err := transformToProject(p)
		if err != nil {
			return []project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedProjects = append(transformedProjects, transformedProj)
	}

	return transformedProjects, nil
}

func (r ProjectRepository) Update(ctx context.Context, toUpdate project.Project) (project.Project, error) {
	var updatedProject Project

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updateProjectQuery string
	toUpdate.ID = strings.TrimSpace(toUpdate.ID)
	isUuid := isUUID(toUpdate.ID)

	if isUuid {
		updateProjectQuery, err = buildUpdateProjectByIDQuery(dialect)
	} else {
		updateProjectQuery, err = buildUpdateProjectBySlugQuery(dialect)
	}
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &updatedProject, updateProjectQuery, toUpdate.ID, toUpdate.ID, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.ID, marshaledMetadata)
		})
	} else {
		err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &updatedProject, updateProjectQuery, toUpdate.ID, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.ID, marshaledMetadata)
		})
	}
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return project.Project{}, project.ErrNotExist
		}
		if errors.Is(err, errInvalidTexRepresentation) {
			return project.Project{}, fmt.Errorf("%w: %s", project.ErrInvalidUUID, err)
		}
		return project.Project{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToProject(updatedProject)
	if err != nil {
		return project.Project{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

func (r ProjectRepository) ListAdmins(ctx context.Context, id string) ([]user.User, error) {
	var fetchedUsers []User

	listProjectAdminsQuery, err := buildListProjectAdminsQuery(dialect)
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	id = strings.TrimSpace(id)
	fetchedProject, err := r.Get(ctx, id)
	if err != nil {
		return []user.User{}, err
	}
	id = fetchedProject.ID

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, listProjectAdminsQuery, id)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, project.ErrNoAdminsExist
		}
		return []user.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []user.User
	for _, u := range fetchedUsers {
		transformedUser, err := u.transformToUser()
		if err != nil {
			return []user.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}
