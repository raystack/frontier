package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"database/sql"

	"github.com/doug-martin/goqu/v9"
	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/organization"
	"github.com/odpf/shield/core/project"
	"github.com/odpf/shield/core/role"
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

func (r ProjectRepository) GetByID(ctx context.Context, id string) (project.Project, error) {
	if strings.TrimSpace(id) == "" {
		return project.Project{}, project.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_PROJECTS).Where(goqu.ExOr{
		"id": id,
	}).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &projectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return project.Project{}, project.ErrInvalidUUID
		default:
			return project.Project{}, err
		}
	}

	transformedProject, err := projectModel.transformToProject()
	if err != nil {
		return project.Project{}, err
	}

	return transformedProject, nil
}

func (r ProjectRepository) GetBySlug(ctx context.Context, slug string) (project.Project, error) {
	if strings.TrimSpace(slug) == "" {
		return project.Project{}, project.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_PROJECTS).Where(goqu.Ex{
		"slug": slug,
	}).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &projectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return project.Project{}, project.ErrInvalidUUID
		default:
			return project.Project{}, err
		}
	}

	transformedProject, err := projectModel.transformToProject()
	if err != nil {
		return project.Project{}, err
	}

	return transformedProject, nil
}

func (r ProjectRepository) Create(ctx context.Context, prj project.Project) (project.Project, error) {
	if strings.TrimSpace(prj.Name) == "" || strings.TrimSpace(prj.Slug) == "" {
		return project.Project{}, project.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(prj.Metadata)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_PROJECTS).Rows(
		goqu.Record{
			"name":     prj.Name,
			"slug":     prj.Slug,
			"org_id":   prj.Organization.ID,
			"metadata": marshaledMetadata,
		}).Returning(&Project{}).ToSQL()

	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&projectModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, errForeignKeyViolation):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return project.Project{}, organization.ErrInvalidUUID
		case errors.Is(err, errDuplicateKey):
			return project.Project{}, project.ErrConflict
		default:
			return project.Project{}, err
		}
	}

	transformedProj, err := projectModel.transformToProject()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	return transformedProj, nil
}

func (r ProjectRepository) List(ctx context.Context) ([]project.Project, error) {
	query, params, err := dialect.From(TABLE_PROJECTS).ToSQL()
	if err != nil {
		return []project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var projectModels []Project
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &projectModels, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []project.Project{}, project.ErrNotExist
		}
		return []project.Project{}, fmt.Errorf("%w: %s", dbErr, err)
	}

	var transformedProjects []project.Project
	for _, p := range projectModels {
		transformedProj, err := p.transformToProject()
		if err != nil {
			return []project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
		}

		transformedProjects = append(transformedProjects, transformedProj)
	}

	return transformedProjects, nil
}

func (r ProjectRepository) UpdateByID(ctx context.Context, prj project.Project) (project.Project, error) {
	if strings.TrimSpace(prj.ID) == "" {
		return project.Project{}, project.ErrInvalidID
	}

	if strings.TrimSpace(prj.Name) == "" || strings.TrimSpace(prj.Slug) == "" {
		return project.Project{}, project.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(prj.Metadata)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"name":       prj.Name,
			"slug":       prj.Slug,
			"org_id":     prj.Organization.ID,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": prj.ID,
	}).Returning(&Project{}).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &projectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return project.Project{}, project.ErrInvalidUUID
		case errors.Is(err, errDuplicateKey):
			return project.Project{}, project.ErrConflict
		case errors.Is(err, errForeignKeyViolation):
			return project.Project{}, organization.ErrNotExist
		default:
			return project.Project{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	prj, err = projectModel.transformToProject()
	if err != nil {
		return project.Project{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return prj, nil
}

func (r ProjectRepository) UpdateBySlug(ctx context.Context, prj project.Project) (project.Project, error) {
	if strings.TrimSpace(prj.Slug) == "" {
		return project.Project{}, project.ErrInvalidID
	}

	if strings.TrimSpace(prj.Name) == "" {
		return project.Project{}, project.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(prj.Metadata)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"name":       prj.Name,
			"slug":       prj.Slug,
			"org_id":     prj.Organization.ID,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"slug": prj.Slug,
	}).Returning(&Project{}).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &projectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, errInvalidTexRepresentation):
			return project.Project{}, project.ErrInvalidUUID
		case errors.Is(err, errDuplicateKey):
			return project.Project{}, project.ErrConflict
		case errors.Is(err, errForeignKeyViolation):
			return project.Project{}, organization.ErrNotExist
		default:
			return project.Project{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	prj, err = projectModel.transformToProject()
	if err != nil {
		return project.Project{}, fmt.Errorf("%s: %w", parseErr, err)
	}

	return prj, nil
}

func (r ProjectRepository) ListAdmins(ctx context.Context, projectID string) ([]user.User, error) {
	var fetchedUsers []User

	query, params, err := dialect.Select(
		goqu.I("u.id").As("id"),
		goqu.I("u.name").As("name"),
		goqu.I("u.email").As("email"),
		goqu.I("u.metadata").As("metadata"),
		goqu.I("u.created_at").As("created_at"),
		goqu.I("u.updated_at").As("updated_at"),
	).
		From(goqu.T(TABLE_RELATIONS).As("r")).Join(
		goqu.T(TABLE_USERS).As("u"), goqu.On(
			goqu.I("u.id").Cast("VARCHAR").Eq(goqu.I("r.subject_id")),
		)).Where(goqu.Ex{
		"r.object_id":            projectID,
		"r.role_id":              role.DefinitionProjectAdmin.ID,
		"r.subject_namespace_id": namespace.DefinitionUser.ID,
		"r.object_namespace_id":  namespace.DefinitionProject.ID,
	}).ToSQL()
	if err != nil {
		return []user.User{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &fetchedUsers, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []user.User{}, nil
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
