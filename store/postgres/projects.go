package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/doug-martin/goqu/v9"

	"github.com/odpf/shield/internal/bootstrap/definition"
	"github.com/odpf/shield/internal/project"
	"github.com/odpf/shield/model"
)

type Project struct {
	Id        string       `db:"id"`
	Name      string       `db:"name"`
	Slug      string       `db:"slug"`
	OrgId     string       `db:"org_id"`
	Metadata  []byte       `db:"metadata"`
	CreatedAt time.Time    `db:"created_at"`
	UpdatedAt time.Time    `db:"updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at"`
}

// *Get Projects Query
func buildGetProjectsBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	getProjectsBySlugQuery, _, err := dialect.From(TABLE_PROJECTS).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).ToSQL()

	return getProjectsBySlugQuery, err
}

func buildGetProjectsByIdQuery(dialect goqu.DialectWrapper) (string, error) {
	getProjectsByIdQuery, _, err := dialect.From(TABLE_PROJECTS).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).ToSQL()

	return getProjectsByIdQuery, err
}

// *Create Project Query
func buildCreateProjectQuery(dialect goqu.DialectWrapper) (string, error) {
	createProjectQuery, _, err := dialect.Insert(TABLE_PROJECTS).Rows(
		goqu.Record{
			"name":     goqu.L("$1"),
			"slug":     goqu.L("$2"),
			"org_id":   goqu.L("$3"),
			"metadata": goqu.L("$4"),
		}).Returning(&Project{}).ToSQL()

	return createProjectQuery, err
}

// *List Project Query
func buildListProjectAdminsQuery(dialect goqu.DialectWrapper) (string, error) {
	listProjectAdminsQuery, _, err := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).From(goqu.T(TABLE_RELATION).As("r")).Join(
		goqu.T(TABLE_USER).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            goqu.L("$1"),
		"r.role_id":              definition.ProjectAdminRole.Id,
		"r.subject_namespace_id": definition.UserNamespace.Id,
		"r.object_namespace_id":  definition.ProjectNamespace.Id,
	}).ToSQL()

	return listProjectAdminsQuery, err
}

func buildListProjectQuery(dialect goqu.DialectWrapper) (string, error) {
	listProjectQuery, _, err := dialect.From(TABLE_PROJECTS).ToSQL()

	return listProjectQuery, err
}

// *Update Project Query
func buildUpdateProjectBySlugQuery(dialect goqu.DialectWrapper) (string, error) {
	updateProjectQuery, _, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"name":       goqu.L("$2"),
			"slug":       goqu.L("$3"),
			"org_id":     goqu.L("$4"),
			"metadata":   goqu.L("$5"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"slug": goqu.L("$1"),
	}).Returning(&Project{}).ToSQL()

	return updateProjectQuery, err
}

func buildUpdateProjectByIdQuery(dialect goqu.DialectWrapper) (string, error) {
	updateProjectQuery, _, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"name":       goqu.L("$3"),
			"slug":       goqu.L("$4"),
			"org_id":     goqu.L("$5"),
			"metadata":   goqu.L("$6"),
			"updated_at": goqu.L("now()"),
		}).Where(goqu.ExOr{
		"id":   goqu.L("$1"),
		"slug": goqu.L("$2"),
	}).Returning(&Project{}).ToSQL()

	return updateProjectQuery, err
}

// GetProject Supports Slug
func (s Store) GetProject(ctx context.Context, id string) (model.Project, error) {
	var fetchedProject Project
	var getProjectsQuery string
	var err error
	id = strings.TrimSpace(id)
	isUuid := isUUID(id)

	if isUuid {
		getProjectsQuery, err = buildGetProjectsByIdQuery(dialect)
	} else {
		getProjectsQuery, err = buildGetProjectsBySlugQuery(dialect)
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &fetchedProject, getProjectsQuery, id, id)
		})
	} else {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &fetchedProject, getProjectsQuery, id)
		})
	}

	if errors.Is(err, sql.ErrNoRows) {
		return model.Project{}, project.ProjectDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Project{}, project.InvalidUUID
	} else if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	if err != nil {
		return model.Project{}, err
	}

	transformedProject, err := transformToProject(fetchedProject)
	if err != nil {
		return model.Project{}, err
	}

	return transformedProject, nil
}

func (s Store) CreateProject(ctx context.Context, projectToCreate model.Project) (model.Project, error) {
	marshaledMetadata, err := json.Marshal(projectToCreate.Metadata)
	if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var newProject Project
	createProjectQuery, err := buildCreateProjectQuery(dialect)
	if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.GetContext(ctx, &newProject, createProjectQuery, projectToCreate.Name, projectToCreate.Slug, projectToCreate.Organization.Id, marshaledMetadata)
	})

	if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	transformedProj, err := transformToProject(newProject)
	if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedProj, nil
}

func (s Store) ListProject(ctx context.Context) ([]model.Project, error) {
	var fetchedProjects []Project
	listProjectQuery, err := buildListProjectQuery(dialect)
	if err != nil {
		return []model.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedProjects, listProjectQuery)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.Project{}, project.ProjectDoesntExist
	}

	if err != nil {
		return []model.Project{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedProjects []model.Project

	for _, p := range fetchedProjects {
		transformedProj, err := transformToProject(p)
		if err != nil {
			return []model.Project{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedProjects = append(transformedProjects, transformedProj)
	}

	return transformedProjects, nil
}

// UpdateProject Supports Slug
func (s Store) UpdateProject(ctx context.Context, toUpdate model.Project) (model.Project, error) {
	var updatedProject Project

	marshaledMetadata, err := json.Marshal(toUpdate.Metadata)
	if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	var updateProjectQuery string
	toUpdate.Id = strings.TrimSpace(toUpdate.Id)
	isUuid := isUUID(toUpdate.Id)

	if isUuid {
		updateProjectQuery, err = buildUpdateProjectByIdQuery(dialect)
	} else {
		updateProjectQuery, err = buildUpdateProjectBySlugQuery(dialect)
	}
	if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if isUuid {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &updatedProject, updateProjectQuery, toUpdate.Id, toUpdate.Id, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.Id, marshaledMetadata)
		})
	} else {
		err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
			return s.DB.GetContext(ctx, &updatedProject, updateProjectQuery, toUpdate.Id, toUpdate.Name, toUpdate.Slug, toUpdate.Organization.Id, marshaledMetadata)
		})
	}
	if errors.Is(err, sql.ErrNoRows) {
		return model.Project{}, project.ProjectDoesntExist
	} else if err != nil && fmt.Sprintf("%s", err.Error()[0:38]) == "pq: invalid input syntax for type uuid" {
		// TODO: this uuid syntax is a error defined in db, not in library
		// need to look into better ways to implement this
		return model.Project{}, fmt.Errorf("%w: %s", project.InvalidUUID, err)
	} else if err != nil {
		return model.Project{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	toUpdate, err = transformToProject(updatedProject)
	if err != nil {
		return model.Project{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return toUpdate, nil
}

// ListProjectAdmins Supports Slug
func (s Store) ListProjectAdmins(ctx context.Context, id string) ([]model.User, error) {
	var fetchedUsers []User

	listProjectAdminsQuery, err := buildListProjectAdminsQuery(dialect)
	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	id = strings.TrimSpace(id)
	isUuid := isUUID(id)
	if !isUuid {
		fetchedProject, err := s.GetProject(ctx, id)
		if err != nil {
			return []model.User{}, err
		}
		id = fetchedProject.Id
	}

	err = s.DB.WithTimeout(ctx, func(ctx context.Context) error {
		return s.DB.SelectContext(ctx, &fetchedUsers, listProjectAdminsQuery, id)
	})

	if errors.Is(err, sql.ErrNoRows) {
		return []model.User{}, project.NoAdminsExist
	}

	if err != nil {
		return []model.User{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedUsers []model.User
	for _, u := range fetchedUsers {
		transformedUser, err := transformToUser(u)
		if err != nil {
			return []model.User{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedUsers = append(transformedUsers, transformedUser)
	}

	return transformedUsers, nil
}

func transformToProject(from Project) (model.Project, error) {
	var unmarshalledMetadata map[string]any
	if err := json.Unmarshal(from.Metadata, &unmarshalledMetadata); err != nil {
		return model.Project{}, err
	}

	return model.Project{
		Id:           from.Id,
		Name:         from.Name,
		Slug:         from.Slug,
		Organization: model.Organization{Id: from.OrgId},
		Metadata:     unmarshalledMetadata,
		CreatedAt:    from.CreatedAt,
		UpdatedAt:    from.UpdatedAt,
	}, nil
}
