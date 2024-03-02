package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/pkg/db"
)

type AuditRepository struct {
	dbc *db.Client
}

func NewAuditRepository(dbc *db.Client) *AuditRepository {
	return &AuditRepository{
		dbc: dbc,
	}
}

func (a AuditRepository) Create(ctx context.Context, l *audit.Log) error {
	if strings.TrimSpace(l.Source) == "" || strings.TrimSpace(l.Action) == "" {
		return audit.ErrInvalidDetail
	}
	if l.ID == "" {
		l.ID = uuid.NewString()
	}

	marshaledActor, err := json.Marshal(l.Actor)
	if err != nil {
		return fmt.Errorf("%w: %s", parseErr, err)
	}
	marshaledTarget, err := json.Marshal(l.Target)
	if err != nil {
		return fmt.Errorf("%w: %s", parseErr, err)
	}
	marshaledMetadata, err := json.Marshal(l.Metadata)
	if err != nil {
		return fmt.Errorf("%w: %s", parseErr, err)
	}

	query, params, err := dialect.Insert(TABLE_AUDITLOGS).Rows(
		goqu.Record{
			"id":       l.ID,
			"org_id":   l.OrgID,
			"source":   l.Source,
			"action":   l.Action,
			"actor":    marshaledActor,
			"target":   marshaledTarget,
			"metadata": marshaledMetadata,
		}).Returning(&Audit{}).ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %s", queryErr, err)
	}

	var auditModel Audit
	if err = a.dbc.WithTimeout(ctx, TABLE_AUDITLOGS, "Create", func(ctx context.Context) error {
		return a.dbc.QueryRowxContext(ctx, query, params...).StructScan(&auditModel)
	}); err != nil {
		return fmt.Errorf("failed to insert audit in pg repo: %w", err)
	}
	return nil
}

func (a AuditRepository) List(ctx context.Context, flt audit.Filter) ([]audit.Log, error) {
	if flt.OrgID == "" {
		return nil, errors.New("org id is required")
	}

	sqlStatement := dialect.From(TABLE_AUDITLOGS)
	sqlStatement = sqlStatement.Where(goqu.Ex{"org_id": flt.OrgID})

	if flt.Source != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"source": flt.Source})
	}
	if flt.Action != "" {
		sqlStatement = sqlStatement.Where(goqu.Ex{"action": flt.Action})
	}
	if flt.StartTime.UnixNano() > 0 {
		sqlStatement = sqlStatement.Where(goqu.Ex{"created_at": goqu.Op{"gte": flt.StartTime}})
	}
	if flt.EndTime.UnixNano() > 0 {
		sqlStatement = sqlStatement.Where(goqu.Ex{"created_at": goqu.Op{"lte": flt.EndTime}})
	}

	query, params, err := sqlStatement.Order(goqu.C("created_at").Desc()).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %s", queryErr, err)
	}

	var fetched []Audit
	if err = a.dbc.WithTimeout(ctx, TABLE_AUDITLOGS, "List", func(ctx context.Context) error {
		return a.dbc.SelectContext(ctx, &fetched, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, nil
		default:
			return nil, fmt.Errorf("%w: %s", dbErr, err)
		}
	}

	transformedLogs := make([]audit.Log, 0, len(fetched))
	for _, v := range fetched {
		transformedGroup, err := v.transform()
		if err != nil {
			return nil, fmt.Errorf("%w: %s", parseErr, err)
		}
		transformedLogs = append(transformedLogs, transformedGroup)
	}

	return transformedLogs, nil
}

func (a AuditRepository) GetByID(ctx context.Context, id string) (audit.Log, error) {
	if strings.TrimSpace(id) == "" {
		return audit.Log{}, audit.ErrInvalidID
	}

	query, params, err := dialect.From(TABLE_AUDITLOGS).Where(
		goqu.Ex{
			"id": id,
		}).Where(notDisabledGroupExp).ToSQL()
	if err != nil {
		return audit.Log{}, fmt.Errorf("%w: %s", queryErr, err)
	}

	var logModel Audit
	if err = a.dbc.WithTimeout(ctx, TABLE_AUDITLOGS, "GetByID", func(ctx context.Context) error {
		return a.dbc.GetContext(ctx, &logModel, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return audit.Log{}, group.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return audit.Log{}, group.ErrInvalidUUID
		default:
			return audit.Log{}, err
		}
	}
	return logModel.transform()
}
