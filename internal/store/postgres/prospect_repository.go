package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/prospect"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
)

var (
	rqlFilerSupportedColumns  = []string{"id", "name", "email", "phone", "activity", "status", "changed_at", "source", "verified", "created_at", "updated_at"}
	rqlSearchSupportedColumns = []string{"id", "name", "email", "phone", "activity", "status", "source", "verified"}
	rqlGroupSupportedColumns  = []string{"activity", "status", "source", "verified"}
)

type ProspectRepository struct {
	dbc *db.Client
}

func NewProspectRepository(dbc *db.Client) *ProspectRepository {
	return &ProspectRepository{
		dbc: dbc,
	}
}

func (r ProspectRepository) Create(ctx context.Context, prspct prospect.Prospect) (prospect.Prospect, error) {
	marshaledMetadata, err := json.Marshal(prspct.Metadata)
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	insertRow := goqu.Record{
		"name":     prspct.Name,
		"email":    prspct.Email,
		"phone":    prspct.Phone,
		"activity": prspct.Activity,
		"status":   string(prspct.Status),
		"source":   prspct.Source,
		"verified": prspct.Verified,
		"metadata": marshaledMetadata,
	}

	createQuery, params, err := dialect.Insert(TABLE_PROSPECTS).Rows(insertRow).Returning(&Prospect{}).ToSQL()
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	tx, err := r.dbc.BeginTxx(ctx, nil)
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", beginTnxErr, err)
	}

	var prospectModel Prospect
	if err = r.dbc.WithTimeout(ctx, TABLE_PROSPECTS, "Create", func(ctx context.Context) error {
		return tx.QueryRowxContext(ctx, createQuery, params...).StructScan(&prospectModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return prospect.Prospect{}, prospect.ErrEmailActivityAlreadyExists
		default:
			if rbErr := tx.Rollback(); rbErr != nil {
				return prospect.Prospect{}, rbErr
			}
			return prospect.Prospect{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return prospect.Prospect{}, err
	}
	transformedProspect, err := prospectModel.transformToProspect()
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	return transformedProspect, nil
}

func (r ProspectRepository) Get(ctx context.Context, id string) (prospect.Prospect, error) {
	stmt := dialect.From(TABLE_PROSPECTS).Where(goqu.Ex{"id": id})
	query, params, err := stmt.ToSQL()
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var prospectModel Prospect
	if err = r.dbc.WithTimeout(ctx, TABLE_PROSPECTS, "Get", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &prospectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return prospect.Prospect{}, prospect.ErrNotExist
		}
		return prospect.Prospect{}, err
	}

	transformedProspect, err := prospectModel.transformToProspect()
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	return transformedProspect, nil
}

func (r ProspectRepository) List(ctx context.Context, rqlQuery *rql.Query) ([]prospect.Prospect, error) {
	stmt := dialect.From(TABLE_PROSPECTS)
	stmt, err := utils.AddRQLFiltersInQuery(stmt, rqlQuery, rqlFilerSupportedColumns, prospect.Prospect{})
	if err != nil {
		return []prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}
	stmt, err = utils.AddRQLSearchInQuery(stmt, rqlQuery, rqlSearchSupportedColumns)
	if err != nil {
		return []prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}
	stmt, err = utils.AddRQLSortInQuery(stmt, rqlQuery)
	if err != nil {
		return []prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}
	stmt = utils.AddRQLPaginationInQuery(stmt, rqlQuery)

	query, params, err := stmt.ToSQL()
	if err != nil {
		return []prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}
	fmt.Println(query)
	var prospectModel []Prospect
	if err = r.dbc.WithTimeout(ctx, TABLE_PROSPECTS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &prospectModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return []prospect.Prospect{}, prospect.ErrNotExist
		}
		return []prospect.Prospect{}, err
	}

	groupStmt, err := utils.AddGroupInQuery(dialect.From(TABLE_PROSPECTS), rqlQuery, rqlGroupSupportedColumns)
	if err != nil {
		fmt.Println("err groupStmt", err)
		return []prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}
	groupCountQuery, params, err := groupStmt.ToSQL()
	if err != nil {
		fmt.Println("err groupCountQuery", err)
		return []prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}
	fmt.Println("groupCountQuery", groupCountQuery)
	type GroupCount struct {
		Column1 string `db:"activity"`
		Column2 string `db:"status,omitempty"`
		Count   int    `db:"count"`
	}
	var groupResults []GroupCount
	// var groupResults []map[string]interface{}
	if err = r.dbc.WithTimeout(ctx, TABLE_PROSPECTS, "groupCount", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &groupResults, groupCountQuery, params...)
	}); err != nil {
		fmt.Println("err groupResults::", err)
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return []prospect.Prospect{}, prospect.ErrNotExist
		}
		return []prospect.Prospect{}, err
	}
	fmt.Printf("Group count results: %#v\n", groupResults)

	/*
			Group count results:
			[{Column1:copilot Column2:subscribed Count:1}
		{Column1:newsletter Column2:unsubscribed Count:8}
		{Column1:newsletter Column2:subscribed Count:6}
		{Column1:waitlist Column2:unsubscribed Count:3}
		{Column1:waitlist Column2:subscribed Count:1}]
	*/

	var transformedProspects []prospect.Prospect
	for _, a := range prospectModel {
		transformedProspect, err := a.transformToProspect()
		if err != nil {
			return []prospect.Prospect{}, fmt.Errorf("%w: %w", parseErr, err)
		}
		transformedProspects = append(transformedProspects, transformedProspect)
	}
	return transformedProspects, nil
}

func (r ProspectRepository) Update(ctx context.Context, prspct prospect.Prospect) (prospect.Prospect, error) {
	marshaledMetadata, err := json.Marshal(prspct.Metadata)
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	updateRow := goqu.Record{
		"name":     prspct.Name,
		"email":    prspct.Email, // todo check if we can update email (maybe not after validation?)
		"phone":    prspct.Phone,
		"activity": prspct.Activity,
		"status":   string(prspct.Status),
		"source":   prspct.Source,
		"verified": prspct.Verified,
		"metadata": marshaledMetadata,
	}
	updateQuery, params, err := dialect.Update(TABLE_PROSPECTS).
		Set(updateRow).
		Where(goqu.Ex{"id": prspct.ID}).
		Returning(&Prospect{}).
		ToSQL()
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	tx, err := r.dbc.BeginTxx(ctx, nil)
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", beginTnxErr, err)
	}

	var prospectModel Prospect
	if err = r.dbc.WithTimeout(ctx, TABLE_PROSPECTS, "Update",
		func(ctx context.Context) error {
			return tx.QueryRowxContext(ctx, updateQuery, params...).StructScan(&prospectModel)
		}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return prospect.Prospect{}, prospect.ErrNotExist
		case errors.Is(err, ErrDuplicateKey):
			return prospect.Prospect{}, prospect.ErrEmailActivityAlreadyExists
		default:
			if rbErr := tx.Rollback(); rbErr != nil {
				return prospect.Prospect{}, rbErr
			}
			return prospect.Prospect{}, err
		}
	}
	if err = tx.Commit(); err != nil {
		return prospect.Prospect{}, err
	}
	transformedProspect, err := prospectModel.transformToProspect()
	if err != nil {
		return prospect.Prospect{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	return transformedProspect, nil
}

func (r ProspectRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_PROSPECTS).Where(goqu.Ex{"id": id}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", queryErr, err)
	}
	return r.dbc.WithTimeout(ctx, TABLE_PROSPECTS, "Delete", func(ctx context.Context) error {
		_, err := r.dbc.ExecContext(ctx, query, params...)
		if err != nil {
			err = checkPostgresError(err)
			switch {
			case errors.Is(err, sql.ErrNoRows):
				return prospect.ErrNotExist
			default:
				return err
			}
		}
		return nil
	})
}
