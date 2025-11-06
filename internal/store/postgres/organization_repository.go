package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/raystack/frontier/pkg/auditrecord"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/db"
)

type OrganizationRepository struct {
	dbc *db.Client
}

func NewOrganizationRepository(dbc *db.Client) *OrganizationRepository {
	return &OrganizationRepository{
		dbc: dbc,
	}
}

var notDisabledOrgExp = goqu.Or(
	goqu.Ex{
		"state": nil,
	},
	goqu.Ex{
		"state": goqu.Op{"neq": organization.Disabled},
	},
)

func (r OrganizationRepository) GetByID(ctx context.Context, id string) (organization.Organization, error) {
	if strings.TrimSpace(id) == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"id": id,
	}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "GetByID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &orgModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.Organization{}, organization.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return organization.Organization{}, organization.ErrInvalidUUID
		default:
			return organization.Organization{}, err
		}
	}

	transformedOrg, err := orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return transformedOrg, nil
}

func (r OrganizationRepository) GetByIDs(ctx context.Context, ids []string) ([]organization.Organization, error) {
	if len(ids) == 0 {
		return nil, organization.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"id": goqu.Op{"in": ids},
	}).Where(notDisabledOrgExp).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgs []Organization
	// TODO(kushsharma): clean up this unnecessary newrelic blot over each query
	// abstract it over database using a facade
	if err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "GetByIDs", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &orgs, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, organization.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return nil, organization.ErrInvalidUUID
		default:
			return nil, err
		}
	}

	var transformedOrgs []organization.Organization
	for _, o := range orgs {
		to, err := o.transformToOrg()
		if err != nil {
			return nil, err
		}
		transformedOrgs = append(transformedOrgs, to)
	}
	return transformedOrgs, nil
}

func (r OrganizationRepository) GetByName(ctx context.Context, name string) (organization.Organization, error) {
	if strings.TrimSpace(name) == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_ORGANIZATIONS).Where(goqu.Ex{
		"name": name,
	}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "GetByName", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &orgModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.Organization{}, organization.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return organization.Organization{}, organization.ErrInvalidUUID
		default:
			return organization.Organization{}, err
		}
	}

	transformedOrg, err := orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return transformedOrg, nil
}

func (r OrganizationRepository) Create(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	if strings.TrimSpace(org.Name) == "" {
		return organization.Organization{}, organization.ErrInvalidDetail
	}
	org.Name = strings.ToLower(org.Name)

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	insertRow := goqu.Record{
		"name":     org.Name,
		"title":    org.Title,
		"avatar":   org.Avatar,
		"metadata": marshaledMetadata,
	}
	if org.State != "" {
		insertRow["state"] = org.State
	}
	query, params, err := dialect.Insert(TABLE_ORGANIZATIONS).Rows(insertRow).Returning(&Organization{}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgModel Organization

	// Use WithTxn for transaction with audit, wrapping WithTimeout for query timeout
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "Upsert", func(ctx context.Context) error {
			// Execute org insert
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&orgModel); err != nil {
				return err
			}

			// Build and insert audit record
			auditRecord := BuildAuditRecord(
				ctx,
				auditrecord.OrganizationCreateEvent,
				AuditResource{
					ID:       orgModel.ID,
					Type:     auditrecord.OrganizationType,
					Name:     nullStringToString(orgModel.Title),
					Metadata: org.Metadata,
				},
				nil,
				orgModel.ID,
				nil,
				orgModel.CreatedAt,
			)

			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return organization.Organization{}, organization.ErrConflict
		default:
			return organization.Organization{}, err
		}
	}

	transformedOrg, err := orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return transformedOrg, nil
}

func (r OrganizationRepository) List(ctx context.Context, flt organization.Filter) ([]organization.Organization, error) {
	stmt := dialect.From(TABLE_ORGANIZATIONS)
	if flt.State == "" {
		stmt = stmt.Where(notDisabledOrgExp)
	} else {
		stmt = stmt.Where(goqu.Ex{
			"state": flt.State.String(),
		})
	}
	if len(flt.IDs) > 0 {
		stmt = stmt.Where(goqu.Ex{
			"id": goqu.Op{"in": flt.IDs},
		})
	}

	if flt.Pagination != nil {
		offset := flt.Pagination.Offset()
		limit := flt.Pagination.PageSize

		totalCountStmt := stmt.Select(goqu.COUNT("*"))
		totalCountQuery, _, err := totalCountStmt.ToSQL()
		if err != nil {
			return []organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
		}

		var totalCount int32
		if err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "Count", func(ctx context.Context) error {
			return r.dbc.GetContext(ctx, &totalCount, totalCountQuery)
		}); err != nil {
			return nil, fmt.Errorf("%w: %w", dbErr, err)
		}

		flt.Pagination.SetCount(totalCount)
		stmt = stmt.Limit(uint(limit)).Offset(uint(offset)).Order(goqu.C("created_at").Desc())
	}

	query, params, err := stmt.ToSQL()
	if err != nil {
		return []organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgModels []Organization
	if err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &orgModels, query, params...)
	}); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []organization.Organization{}, nil
		}
		return []organization.Organization{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	var transformedOrgs []organization.Organization
	for _, o := range orgModels {
		transformedOrg, err := o.transformToOrg()
		if err != nil {
			return []organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
		}
		transformedOrgs = append(transformedOrgs, transformedOrg)
	}

	return transformedOrgs, nil
}

// buildOrgUpdateAuditRecord creates an audit record for organization updates
func buildOrgUpdateAuditRecord(ctx context.Context, orgBeforeUpdate, orgAfterUpdate Organization, metadata map[string]interface{}) AuditRecord {
	title := nullStringToString(orgBeforeUpdate.Title)
	updatedTitle := nullStringToString(orgAfterUpdate.Title)

	auditMetadata := metadata
	if title != updatedTitle {
		// Create a new one to avoid mutating the original
		auditMetadata = make(map[string]interface{})
		if metadata != nil {
			for k, v := range metadata {
				auditMetadata[k] = v
			}
		}
		auditMetadata["title"] = title
		auditMetadata["updated_title"] = updatedTitle
	}

	auditRecord := BuildAuditRecord(
		ctx,
		auditrecord.OrganizationUpdateEvent,
		AuditResource{
			ID:       orgAfterUpdate.ID,
			Type:     auditrecord.OrganizationType,
			Name:     title,
			Metadata: auditMetadata,
		},
		nil,
		orgAfterUpdate.ID,
		nil,
		orgAfterUpdate.UpdatedAt,
	)
	// Pre-populate OrganizationName with title to avoid fetching the updated title from DB
	auditRecord.OrganizationName = title
	return auditRecord
}

func (r OrganizationRepository) UpdateByID(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	if strings.TrimSpace(org.ID) == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	if strings.TrimSpace(org.Name) == "" {
		return organization.Organization{}, organization.ErrInvalidDetail
	}

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	// Query to fetch org title before update
	getQuery, getParams, err := dialect.From(TABLE_ORGANIZATIONS).
		Select("title").
		Where(goqu.Ex{"id": org.ID}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	query, params, err := dialect.Update(TABLE_ORGANIZATIONS).Set(
		goqu.Record{
			"title":      org.Title,
			"avatar":     org.Avatar,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(goqu.Ex{
		"id": org.ID,
	}).Returning(&Organization{}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "Update", func(ctx context.Context) error {
			// Fetch title before update
			var orgBeforeUpdate Organization
			if err := tx.QueryRowxContext(ctx, getQuery, getParams...).StructScan(&orgBeforeUpdate); err != nil {
				return err
			}

			// Execute update
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&orgModel); err != nil {
				return err
			}

			// Create audit record
			auditRecord := buildOrgUpdateAuditRecord(ctx, orgBeforeUpdate, orgModel, org.Metadata)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.Organization{}, organization.ErrNotExist
		case errors.Is(err, ErrDuplicateKey):
			return organization.Organization{}, organization.ErrConflict
		case errors.Is(err, ErrInvalidTextRepresentation):
			return organization.Organization{}, organization.ErrInvalidUUID
		default:
			return organization.Organization{}, fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	org, err = orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return org, nil
}

func (r OrganizationRepository) UpdateByName(ctx context.Context, org organization.Organization) (organization.Organization, error) {
	if strings.TrimSpace(org.Name) == "" {
		return organization.Organization{}, organization.ErrInvalidID
	}

	marshaledMetadata, err := json.Marshal(org.Metadata)
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	// Query to fetch org data before update
	getQuery, getParams, err := dialect.From(TABLE_ORGANIZATIONS).
		Select("title").
		Where(goqu.Ex{"name": org.Name}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	query, params, err := dialect.Update(TABLE_ORGANIZATIONS).Set(
		goqu.Record{
			"title":      org.Title,
			"avatar":     org.Avatar,
			"metadata":   marshaledMetadata,
			"updated_at": goqu.L("now()"),
		}).Where(
		goqu.Ex{
			"name": org.Name,
		}).Returning(&Organization{}).ToSQL()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "UpdateByName", func(ctx context.Context) error {
			// Fetch title before update
			var orgBeforeUpdate Organization
			if err := tx.QueryRowxContext(ctx, getQuery, getParams...).StructScan(&orgBeforeUpdate); err != nil {
				return err
			}

			// Execute update
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&orgModel); err != nil {
				return err
			}

			auditRecord := buildOrgUpdateAuditRecord(ctx, orgBeforeUpdate, orgModel, org.Metadata)
			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.Organization{}, organization.ErrNotExist
		case errors.Is(err, ErrDuplicateKey):
			return organization.Organization{}, organization.ErrConflict
		default:
			return organization.Organization{}, fmt.Errorf("%w: %w", txnErr, err)
		}
	}

	org, err = orgModel.transformToOrg()
	if err != nil {
		return organization.Organization{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	return org, nil
}

func (r OrganizationRepository) SetState(ctx context.Context, id string, state organization.State) error {
	query, params, err := dialect.Update(TABLE_ORGANIZATIONS).Set(
		goqu.Record{
			"state": state.String(),
		}).Where(
		goqu.Ex{
			"id": id,
		},
	).Returning(&Organization{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "SetState", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&orgModel); err != nil {
				return err
			}

			auditRecord := BuildAuditRecord(
				ctx,
				auditrecord.OrganizationStateChangeEvent,
				AuditResource{
					ID:   orgModel.ID,
					Type: auditrecord.OrganizationType,
					Name: nullStringToString(orgModel.Title),
					Metadata: map[string]interface{}{
						"state": state.String(),
					},
				},
				nil,
				orgModel.ID,
				nil,
				orgModel.UpdatedAt,
			)

			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.ErrNotExist
		default:
			return err
		}
	}
	return nil
}

func (r OrganizationRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Delete(TABLE_ORGANIZATIONS).Where(
		goqu.Ex{
			"id": id,
		},
	).Returning(&Organization{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", queryErr, err)
	}

	var orgModel Organization
	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS, "Delete", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&orgModel); err != nil {
				return err
			}

			auditRecord := BuildAuditRecord(
				ctx,
				auditrecord.OrganizationDeleteEvent,
				AuditResource{
					ID:   orgModel.ID,
					Type: auditrecord.OrganizationType,
					Name: nullStringToString(orgModel.Title),
				},
				nil,
				orgModel.ID,
				nil,
				orgModel.DeletedAt.Time,
			)

			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return organization.ErrNotExist
		default:
			return err
		}
	}
	return nil
}
