package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/pkg/db"
)

type AuditRecordRepository struct {
	dbc *db.Client
}

func NewAuditRecordRepository(dbc *db.Client) *AuditRecordRepository {
	return &AuditRecordRepository{
		dbc: dbc,
	}
}

func (r AuditRecordRepository) Create(ctx context.Context, auditRecord auditrecord.AuditRecord) (auditrecord.AuditRecord, error) {
	dbRecord, err := transformFromDomain(auditRecord)
	if err != nil {
		return auditrecord.AuditRecord{}, err
	}

	var auditRecordModel AuditRecord
	query, params, err := dialect.Insert(TABLE_AUDITRECORDS).Rows(dbRecord).Returning(&AuditRecord{}).ToSQL()
	if err != nil {
		return auditrecord.AuditRecord{}, err
	}
	if err = r.dbc.WithTimeout(ctx, TABLE_AUDITRECORDS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&auditRecordModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return auditrecord.AuditRecord{}, auditrecord.ErrIdempotencyKeyConflict
		case errors.Is(err, ErrInvalidTextRepresentation):
			return auditrecord.AuditRecord{}, auditrecord.ErrInvalidUUID
		default:
			return auditrecord.AuditRecord{}, fmt.Errorf("%w: %w", dbErr, err)
		}
	}
	transformedAuditRecord, err := auditRecordModel.transformToDomain()
	if err != nil {
		return auditrecord.AuditRecord{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	return transformedAuditRecord, nil
}

func (r AuditRecordRepository) GetByID(ctx context.Context, id string) (auditrecord.AuditRecord, error) {
	return r.getByField(ctx, "id", id, "GetByID")
}

func (r AuditRecordRepository) GetByIdempotencyKey(ctx context.Context, key string) (auditrecord.AuditRecord, error) {
	return r.getByField(ctx, "idempotency_key", key, "GetByIdempotencyKey")
}

func (r AuditRecordRepository) getByField(ctx context.Context, field string, value interface{}, operation string) (auditrecord.AuditRecord, error) {
	if str, ok := value.(string); ok && str == "" {
		return auditrecord.AuditRecord{}, auditrecord.ErrNotFound
	}

	var auditRecordModel AuditRecord
	query, params, err := dialect.Select().From(TABLE_AUDITRECORDS).Where(goqu.Ex{field: value}).ToSQL()
	if err != nil {
		return auditrecord.AuditRecord{}, err
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_AUDITRECORDS, operation, func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&auditRecordModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return auditrecord.AuditRecord{}, auditrecord.ErrNotFound
		case errors.Is(err, ErrInvalidTextRepresentation):
			return auditrecord.AuditRecord{}, auditrecord.ErrInvalidUUID
		default:
			return auditrecord.AuditRecord{}, fmt.Errorf("%w: %w", dbErr, err)
		}
	}

	transformedAuditRecord, err := auditRecordModel.transformToDomain()
	if err != nil {
		return auditrecord.AuditRecord{}, fmt.Errorf("%w: %w", parseErr, err)
	}
	return transformedAuditRecord, nil
}
