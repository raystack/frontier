package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/db"
)

type ProjectRepository struct {
	dbc *db.Client
}

func NewProjectRepository(dbc *db.Client) *ProjectRepository {
	return &ProjectRepository{
		dbc: dbc,
	}
}

var notDisabledProjectExp = goqu.Or(
	goqu.Ex{
		"state": nil,
	},
	goqu.Ex{
		"state": goqu.Op{"neq": project.Disabled},
	},
)

func (r ProjectRepository) GetByID(ctx context.Context, id string) (project.Project, error) {
	if strings.TrimSpace(id) == "" {
		return project.Project{}, project.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_PROJECTS).Where(goqu.ExOr{
		"id": id,
	}).Where(notDisabledProjectExp).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "GetByID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &projectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
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

func (r ProjectRepository) GetByName(ctx context.Context, name string) (project.Project, error) {
	if strings.TrimSpace(name) == "" {
		return project.Project{}, project.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_PROJECTS).Where(goqu.Ex{
		"name": name,
	}).Where(notDisabledProjectExp).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "GetByName", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &projectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
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
	if strings.TrimSpace(prj.Name) == "" {
		return project.Project{}, project.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(prj.Metadata)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	insertRow := goqu.Record{
		"name":     prj.Name,
		"title":    prj.Title,
		"org_id":   prj.Organization.ID,
		"metadata": marshaledMetadata,
	}
	if prj.State != "" {
		insertRow["state"] = prj.State
	}
	query, params, err := dialect.Insert(TABLE_PROJECTS).Rows(insertRow).Returning(&Project{}).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&projectModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrForeignKeyViolation):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return project.Project{}, organization.ErrInvalidUUID
		case errors.Is(err, ErrDuplicateKey):
			return project.Project{}, project.ErrConflict
		default:
			return project.Project{}, err
		}
	}

	transformedProj, err := projectModel.transformToProject()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return transformedProj, nil
}

func (r ProjectRepository) List(ctx context.Context, flt project.Filter) ([]project.Project, error) {
	stmt := dialect.From(TABLE_PROJECTS)
	if flt.OrgID != "" {
		stmt = stmt.Where(goqu.Ex{
			"org_id": flt.OrgID,
		})
	}
	if len(flt.ProjectIDs) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"id": goqu.Op{"in": flt.ProjectIDs},
		})
	}
	if flt.State == "" {
		stmt = stmt.Where(notDisabledProjectExp)
	} else {
		stmt = stmt.Where(goqu.Ex{
			"state": flt.State.String(),
		})
	}
	query, params, err := stmt.ToSQL()
	if err != nil {
		return []project.Project{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var projectModels []Project
	if err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &projectModels, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []project.Project{}, project.ErrNotExist
		}
		return []project.Project{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	var transformedProjects []project.Project
	for _, p := range projectModels {
		transformedProj, err := p.transformToProject()
		if err != nil {
			return []project.Project{}, fmt.Errorf("%w: %w", parseErr, err)
		}

		transformedProjects = append(transformedProjects, transformedProj)
	}

	return transformedProjects, nil
}

func (r ProjectRepository) UpdateByID(ctx context.Context, prj project.Project) (project.Project, error) {
	if strings.TrimSpace(prj.ID) == "" {
		return project.Project{}, project.ErrInvalidID
	}

	marshaledMetadata, err := json.Marshal(prj.Metadata)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"title":      prj.Title,
			"org_id":     prj.Organization.ID,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{"id": prj.ID}).Returning(&Project{}).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "Update", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &projectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return project.Project{}, project.ErrInvalidUUID
		case errors.Is(err, ErrDuplicateKey):
			return project.Project{}, project.ErrConflict
		case errors.Is(err, ErrForeignKeyViolation):
			return project.Project{}, organization.ErrNotExist
		default:
			return project.Project{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	prj, err = projectModel.transformToProject()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return prj, nil
}

func (r ProjectRepository) UpdateByName(ctx context.Context, prj project.Project) (project.Project, error) {
	if strings.TrimSpace(prj.Name) == "" {
		return project.Project{}, project.ErrInvalidID
	}

	marshaledMetadata, err := json.Marshal(prj.Metadata)
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"title":      prj.Title,
			"org_id":     prj.Organization.ID,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{"name": prj.Name}).Returning(&Project{}).ToSQL()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var projectModel Project
	if err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "UpdateByName", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &projectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.Project{}, project.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return project.Project{}, project.ErrInvalidUUID
		case errors.Is(err, ErrDuplicateKey):
			return project.Project{}, project.ErrConflict
		case errors.Is(err, ErrForeignKeyViolation):
			return project.Project{}, organization.ErrNotExist
		default:
			return project.Project{}, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	prj, err = projectModel.transformToProject()
	if err != nil {
		return project.Project{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return prj, nil
}

func (r ProjectRepository) SetState(ctx context.Context, id string, state project.State) error {
	query, params, err := dialect.Update(TABLE_PROJECTS).Set(
		goqu.Record{
			"state": state.String(),
		}).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "SetState", func(ctx context.Context) error {
		if _, err = r.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return user.ErrNotExist
		default:
			return err
		}
	}
	return nil
}

func (r ProjectRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_PROJECTS).Where(
		goqu.Ex{
			"id": id,
		},
	).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_PROJECTS, "Delete", func(ctx context.Context) error {
		if _, err = r.dbc.DB.ExecContext(ctx, query, params...); err != nil {
			return err
		}
		return nil
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return project.ErrNotExist
		default:
			return err
		}
	}
	return nil
}
