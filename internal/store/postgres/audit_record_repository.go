package postgres

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
)

const (
	BATCH_SIZE         = 1000
	ContentTypeTextCSV = "text/csv"
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
		"id", "event", "actor_id", "actor_type", "actor_name", "resource_id", "resource_type", "resource_name",
		"target_id", "target_type", "target_name", "occurred_at", "org_id", "org_name", "request_id", "created_at", "idempotency_key",
	}
	auditRecordRQLSearchSupportedColumns = []string{
		"id", "event", "actor_id", "actor_type", "actor_name", "resource_id", "resource_type", "resource_name",
		"target_id", "target_type", "target_name", "org_id", "org_name", "request_id", "idempotency_key",
	}
	auditRecordRQLGroupSupportedColumns = []string{
		"event", "actor_type", "resource_type", "target_type", "org_id", "org_name",
	}
)

func buildOrgNameQuery(orgID interface{}) (string, []interface{}, error) {
	return dialect.Select("name").
		From(TABLE_ORGANIZATIONS).
		Where(goqu.Ex{"id": orgID, "deleted_at": nil}).
		ToSQL()
}

func (r AuditRecordRepository) Create(ctx context.Context, auditRecord auditrecord.AuditRecord) (auditrecord.AuditRecord, error) {
	// External RPC calls will have actor already enriched by service.
	// Internal service calls will have empty actor and need enrichment from context.
	if auditRecord.Actor.ID == "" {
		enrichActorFromContext(ctx, &auditRecord.Actor)
	}

	dbRecord, err := transformFromDomain(auditRecord)
	if err != nil {
		return auditrecord.AuditRecord{}, err
	}

	// Enrich organization name from DB
	if auditRecord.OrgID != "" {
		query, params, err := buildOrgNameQuery(auditRecord.OrgID)
		if err == nil {
			_ = r.dbc.QueryRowxContext(ctx, query, params...).Scan(&dbRecord.OrganizationName)
		}
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
	baseStmt, err := r.buildFilteredQuery(rqlQuery)
	if err != nil {
		fmt.Println("err", err)
		return auditrecord.AuditRecordsList{}, err
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

func (r AuditRecordRepository) Export(ctx context.Context, rqlQuery *rql.Query) (io.Reader, string, error) {
	baseStmt, err := r.buildFilteredQuery(rqlQuery)
	if err != nil {
		return nil, "", err
	}

	csvQuery := baseStmt.Select(
		goqu.L(`id`),
		goqu.L(`COALESCE(idempotency_key::text, '')`),
		goqu.L(`event`),
		goqu.L(`actor_id`),
		goqu.L(`actor_type`),
		goqu.L(`actor_name`),
		goqu.L(`COALESCE(actor_metadata::text, '{}')`),
		goqu.L(`resource_id`),
		goqu.L(`resource_type`),
		goqu.L(`resource_name`),
		goqu.L(`COALESCE(resource_metadata::text, '{}')`),
		goqu.L(`COALESCE(target_id, '')`),
		goqu.L(`COALESCE(target_type, '')`),
		goqu.L(`COALESCE(target_name, '')`),
		goqu.L(`COALESCE(target_metadata::text, '{}')`),
		goqu.L(`COALESCE(org_id::text, '')`),
		goqu.L(`COALESCE(org_name, '')`),
		goqu.L(`COALESCE(request_id, '')`),
		goqu.L(`to_char(occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS.MS TZ')`),
		goqu.L(`to_char(created_at AT TIME ZONE 'UTC', 'YYYY-MM-DD HH24:MI:SS.MS TZ')`),
		goqu.L(`COALESCE(metadata::text, '{}')`),
	)

	csvQuery, err = utils.AddRQLSortInQuery(csvQuery, rqlQuery)
	if err != nil {
		return nil, "", wrapValidationError(err)
	}

	reader, err := r.executeCursorQuery(ctx, csvQuery)
	if err != nil {
		_, err2 := r.handleListDatabaseError(err)
		return nil, "", err2
	}
	return reader, ContentTypeTextCSV, nil
}

func (r AuditRecordRepository) executeCursorQuery(ctx context.Context, selectQuery *goqu.SelectDataset) (io.Reader, error) {
	query, params, err := selectQuery.ToSQL()
	if err != nil {
		return nil, fmt.Errorf("failed to build SQL query: %w", err)
	}

	pr, pw := io.Pipe()

	go func() {
		defer pw.Close()

		err := r.dbc.WithTimeout(ctx, TABLE_AUDITRECORDS, "Export", func(ctx context.Context) error {
			// Start a read-only transaction for cursor
			tx, err := r.dbc.DB.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
			if err != nil {
				return fmt.Errorf("failed to begin transaction: %w", err)
			}
			defer tx.Rollback()

			// Generate unique cursor name
			cursorName, err := r.generateCursorName()
			if err != nil {
				return fmt.Errorf("failed to generate cursor name: %w", err)
			}

			// Declare cursor with the SELECT query
			declareSQL := fmt.Sprintf("DECLARE %s CURSOR FOR %s", cursorName, query)
			_, err = tx.ExecContext(ctx, declareSQL, params...)
			if err != nil {
				return fmt.Errorf("failed to declare cursor: %w", err)
			}

			// Stream data using cursor
			return r.streamCursorToCSV(ctx, tx, cursorName, pw)
		})

		if err != nil {
			pw.CloseWithError(fmt.Errorf("cursor export failed: %w", err))
		}
	}()

	return pr, nil
}

func (r AuditRecordRepository) buildFilteredQuery(rqlQuery *rql.Query) (*goqu.SelectDataset, error) {
	if rqlQuery == nil {
		rqlQuery = utils.NewRQLQuery("", utils.DefaultOffset, utils.DefaultLimit, []rql.Filter{}, []rql.Sort{}, []string{})
	}

	baseStmt := dialect.From(TABLE_AUDITRECORDS).Where(goqu.Ex{"deleted_at": nil})

	// Apply filters
	baseStmt, err := utils.AddRQLFiltersInQuery(baseStmt, rqlQuery, auditRecordRQLFilterSupportedColumns, auditrecord.AuditRecordRQLSchema{})
	if err != nil {
		return nil, wrapValidationError(err)
	}

	// Apply search
	baseStmt, err = utils.AddRQLSearchInQuery(baseStmt, rqlQuery, auditRecordRQLSearchSupportedColumns)
	if err != nil {
		return nil, wrapValidationError(err)
	}

	return baseStmt, nil
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

// generateCursorName creates a unique cursor name for the export operation
func (r AuditRecordRepository) generateCursorName() (string, error) {
	randomBytes := make([]byte, 8)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return "export_cursor_" + hex.EncodeToString(randomBytes), nil
}

// streamCursorToCSV fetches data from cursor in batches and streams as CSV
func (r AuditRecordRepository) streamCursorToCSV(ctx context.Context, tx *sql.Tx, cursorName string, pw *io.PipeWriter) error {
	csvBuffer := &bytes.Buffer{}
	csvWriter := csv.NewWriter(csvBuffer)

	// Write CSV headers first
	headers := []string{
		"Record ID", "Idempotency Key", "Event", "Actor ID", "Actor Type", "Actor Name", "Actor Metadata",
		"Resource ID", "Resource Type", "Resource Name", "Resource Metadata", "Target ID", "Target Type",
		"Target name", "Target Metadata", "Organization ID", "Organization Name", "Request ID", "Occurred At", "Created At", "Metadata",
	}
	if err := csvWriter.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}
	csvWriter.Flush()

	// Send headers as first chunk
	if _, err := pw.Write(csvBuffer.Bytes()); err != nil {
		return fmt.Errorf("failed to write headers: %w", err)
	}
	csvBuffer.Reset()
	csvWriter = csv.NewWriter(csvBuffer)

	rowCount := 0
	for {
		// Fetch batch from cursor
		fetchSQL := fmt.Sprintf("FETCH %d FROM %s", BATCH_SIZE, cursorName)
		rows, err := tx.QueryContext(ctx, fetchSQL)
		if err != nil {
			return fmt.Errorf("failed to fetch from cursor: %w", err)
		}

		batchRowCount := 0
		// Process batch rows
		for rows.Next() {
			// Scan all columns into string slice
			values := make([]string, len(headers))
			valuePtrs := make([]interface{}, len(headers))
			for i := range values {
				valuePtrs[i] = &values[i]
			}
			// scan values positionally
			if err := rows.Scan(valuePtrs...); err != nil {
				rows.Close()
				return fmt.Errorf("failed to scan row: %w", err)
			}

			// Write row to CSV
			if err := csvWriter.Write(values); err != nil {
				rows.Close()
				return fmt.Errorf("failed to write CSV row: %w", err)
			}

			batchRowCount++
			rowCount++

			// Send chunk every BATCH_SIZE rows
			if rowCount%BATCH_SIZE == 0 {
				csvWriter.Flush()
				if _, err := pw.Write(csvBuffer.Bytes()); err != nil {
					rows.Close()
					return fmt.Errorf("failed to write batch: %w", err)
				}
				csvBuffer.Reset()
			}
		}

		rows.Close()
		if err := rows.Err(); err != nil {
			return fmt.Errorf("cursor iteration error: %w", err)
		}

		// If no more rows, break
		if batchRowCount == 0 {
			break
		}
	}

	// Send final chunk if there's any remaining data
	csvWriter.Flush()
	if csvBuffer.Len() > 0 {
		finalBytes := csvBuffer.Bytes()
		if _, err := pw.Write(finalBytes); err != nil {
			return fmt.Errorf("failed to write final batch: %w", err)
		}
	}

	return nil
}
