package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
)

func wrapValidationError(err error) error {
	return fmt.Errorf("%w: %s", auditrecord.ErrRepositoryBadInput, err.Error())
}

type AuditRecordRepository struct {
	dbc *db.Client
}

type AuditRecordGroupData struct {
	Name  string `db:"values"`
	Count int64  `db:"count"`
}

func (a AuditRecordGroupData) toGroupData() utils.GroupData {
	return utils.GroupData{
		Name:  a.Name,
		Count: int(a.Count),
	}
}

func NewAuditRecordRepository(dbc *db.Client) *AuditRecordRepository {
	return &AuditRecordRepository{
		dbc: dbc,
	}
}

var (
	auditRecordRQLFilterSupportedColumns = []string{
		"event", "actor_id", "actor_type", "actor_name", "resource_id", "resource_type", "resource_name",
		"target_id", "target_type", "target_name", "occurred_at", "org_id", "req_id", "created_at", "idempotency_key",
	}
	auditRecordRQLSearchSupportedColumns = []string{
		"id", "event", "actor_id", "actor_type", "actor_name", "resource_id", "resource_type", "resource_name",
		"target_id", "target_type", "target_name", "org_id", "req_id", "idempotency_key",
	}
	auditRecordRQLGroupSupportedColumns = []string{
		"event", "actor_type", "resource_type", "target_type", "org_id",
	}
)

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

func (r AuditRecordRepository) List(ctx context.Context, rqlQuery *rql.Query) (auditrecord.AuditRecordsList, error) {
	if rqlQuery == nil {
		rqlQuery = utils.NewRQLQuery("", utils.DefaultOffset, utils.DefaultLimit, []rql.Filter{}, []rql.Sort{}, []string{})
	}

	baseStmt := dialect.From(TABLE_AUDITRECORDS).Where(goqu.Ex{"deleted_at": nil})

	// apply filters
	baseStmt, err := utils.AddRQLFiltersInQuery(baseStmt, rqlQuery, auditRecordRQLFilterSupportedColumns, auditrecord.AuditRecordRQLSchema{})
	if err != nil {
		return auditrecord.AuditRecordsList{}, wrapValidationError(err)
	}

	// apply search
	baseStmt, err = utils.AddRQLSearchInQuery(baseStmt, rqlQuery, auditRecordRQLSearchSupportedColumns)
	if err != nil {
		return auditrecord.AuditRecordsList{}, wrapValidationError(err)
	}

	listStmt := baseStmt
	countStmt := baseStmt
	groupStmt := baseStmt

	// Get total row count
	countQuery, countParams, err := countStmt.Select(goqu.L("COUNT(*) as total")).ToSQL()
	if err != nil {
		return auditrecord.AuditRecordsList{}, err
	}
	var totalCount int64
	if err = r.dbc.WithTimeout(ctx, TABLE_AUDITRECORDS, "Count", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &totalCount, countQuery, countParams...)
	}); err != nil {
		return r.handleListDatabaseError(err)
	}

	// Get group by results
	var groupResults []AuditRecordGroupData

	if len(rqlQuery.GroupBy) > 0 {
		groupStmt, err = utils.AddGroupInQuery(groupStmt, rqlQuery, auditRecordRQLGroupSupportedColumns)
		if err != nil {
			return auditrecord.AuditRecordsList{}, wrapValidationError(err)
		}

		query, params, err := groupStmt.ToSQL()
		if err != nil {
			return auditrecord.AuditRecordsList{}, err
		}

		if err = r.dbc.WithTimeout(ctx, TABLE_AUDITRECORDS, "groupCount", func(ctx context.Context) error {
			return r.dbc.SelectContext(ctx, &groupResults, query, params...)
		}); err != nil {
			return r.handleListDatabaseError(err)
		}
	}

	// List audit records with pagination and sorting
	listStmt, err = utils.AddRQLSortInQuery(listStmt, rqlQuery)
	if err != nil {
		return auditrecord.AuditRecordsList{}, wrapValidationError(err)
	}
	listStmt, pagination := utils.AddRQLPaginationInQuery(listStmt, rqlQuery)

	query, params, err := listStmt.ToSQL()
	if err != nil {
		return auditrecord.AuditRecordsList{}, err
	}

	var auditRecordModels []AuditRecord
	if err = r.dbc.WithTimeout(ctx, TABLE_AUDITRECORDS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &auditRecordModels, query, params...)
	}); err != nil {
		return r.handleListDatabaseError(err)
	}

	page := utils.Page{
		Limit:      pagination.Limit,
		Offset:     pagination.Offset,
		TotalCount: totalCount,
	}

	var transformedAuditRecords []auditrecord.AuditRecord
	for _, ar := range auditRecordModels {
		transformedAuditRecord, err := ar.transformToDomain()
		if err != nil {
			return auditrecord.AuditRecordsList{}, err
		}
		transformedAuditRecords = append(transformedAuditRecords, transformedAuditRecord)
	}

	groupData := make([]utils.GroupData, len(groupResults))
	for i, result := range groupResults {
		groupData[i] = result.toGroupData()
	}

	return auditrecord.AuditRecordsList{
		AuditRecords: transformedAuditRecords,
		Group: &utils.Group{
			Name: strings.Join(rqlQuery.GroupBy, ","),
			Data: groupData,
		},
		Page: page,
	}, nil
}

func (r AuditRecordRepository) handleListDatabaseError(err error) (auditrecord.AuditRecordsList, error) {
	err = checkPostgresError(err)
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return auditrecord.AuditRecordsList{}, auditrecord.ErrNotFound
	case errors.Is(err, ErrInvalidTextRepresentation):
		return auditrecord.AuditRecordsList{}, auditrecord.ErrInvalidUUID
	default:
		return auditrecord.AuditRecordsList{}, err
	}
}
