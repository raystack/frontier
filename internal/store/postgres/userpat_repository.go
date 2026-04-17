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
	"github.com/google/uuid"

	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/rql"
)

var (
	patRQLFilterSupportedColumns = []string{"id", "title", "expires_at", "created_at", "regenerated_at"}
	patRQLSearchSupportedColumns = []string{"id", "title"}
)

type UserPATRepository struct {
	dbc *db.Client
}

func NewUserPATRepository(dbc *db.Client) *UserPATRepository {
	return &UserPATRepository{
		dbc: dbc,
	}
}

func (r UserPATRepository) Create(ctx context.Context, pat models.PAT) (models.PAT, error) {
	if strings.TrimSpace(pat.ID) == "" {
		pat.ID = uuid.New().String()
	}

	marshaledMetadata, err := json.Marshal(pat.Metadata)
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	var model UserPAT
	query, params, err := dialect.Insert(TABLE_USER_PATS).Rows(
		goqu.Record{
			"id":          pat.ID,
			"user_id":     pat.UserID,
			"org_id":      pat.OrgID,
			"title":       pat.Title,
			"secret_hash": pat.SecretHash,
			"metadata":    marshaledMetadata,
			"expires_at":  pat.ExpiresAt,
		}).Returning(&UserPAT{}).ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "Create", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&model)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, ErrDuplicateKey) {
			return models.PAT{}, paterrors.ErrConflict
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) CountActive(ctx context.Context, userID, orgID string) (int64, error) {
	now := time.Now()
	query, params, err := dialect.Select(goqu.COUNT("*")).From(TABLE_USER_PATS).Where(
		goqu.Ex{"user_id": userID},
		goqu.Ex{"org_id": orgID},
		goqu.Ex{"deleted_at": nil},
		goqu.C("expires_at").Gt(now),
	).ToSQL()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", queryErr, err)
	}

	var count int64
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "CountActive", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &count, query, params...)
	}); err != nil {
		return 0, fmt.Errorf("%w: %w", dbErr, err)
	}

	return count, nil
}

func (r UserPATRepository) GetByID(ctx context.Context, id string) (models.PAT, error) {
	query, params, err := dialect.From(TABLE_USER_PATS).
		Select(&UserPAT{}).
		Where(
			goqu.Ex{"id": id},
			goqu.Ex{"deleted_at": nil},
		).Limit(1).ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var model UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "GetByID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &model, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.PAT{}, paterrors.ErrNotFound
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) List(ctx context.Context, userID, orgID string, rqlQuery *rql.Query) (models.PATList, error) {
	baseStmt, err := r.buildPATFilteredQuery(userID, orgID, rqlQuery)
	if err != nil {
		return models.PATList{}, err
	}

	totalCount, err := r.countPATs(ctx, baseStmt)
	if err != nil {
		return models.PATList{}, err
	}

	listStmt := r.applySort(baseStmt.Select(&UserPAT{}), rqlQuery)
	listStmt, pagination := utils.AddRQLPaginationInQuery(listStmt, rqlQuery)

	query, params, err := listStmt.ToSQL()
	if err != nil {
		return models.PATList{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var rows []UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "List", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &rows, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.PATList{}, nil
		}
		return models.PATList{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	pats := make([]models.PAT, 0, len(rows))
	for _, row := range rows {
		pat, err := row.transform()
		if err != nil {
			return models.PATList{}, err
		}
		pats = append(pats, pat)
	}

	return models.PATList{
		PATs: pats,
		Page: utils.Page{
			Limit:      pagination.Limit,
			Offset:     pagination.Offset,
			TotalCount: totalCount,
		},
	}, nil
}

func (r UserPATRepository) buildPATFilteredQuery(userID, orgID string, rqlQuery *rql.Query) (*goqu.SelectDataset, error) {
	if rqlQuery == nil {
		rqlQuery = utils.NewRQLQuery("", utils.DefaultOffset, utils.DefaultLimit, []rql.Filter{}, []rql.Sort{}, []string{})
	}

	baseStmt := dialect.From(TABLE_USER_PATS).Where(
		goqu.Ex{"user_id": userID},
		goqu.Ex{"org_id": orgID},
		goqu.Ex{"deleted_at": nil},
	)

	baseStmt, err := utils.AddRQLFiltersInQuery(baseStmt, rqlQuery, patRQLFilterSupportedColumns, models.PAT{})
	if err != nil {
		return nil, fmt.Errorf("invalid filter: %w", err)
	}

	baseStmt, err = utils.AddRQLSearchInQuery(baseStmt, rqlQuery, patRQLSearchSupportedColumns)
	if err != nil {
		return nil, fmt.Errorf("invalid search: %w", err)
	}

	return baseStmt, nil
}

func (r UserPATRepository) countPATs(ctx context.Context, baseStmt *goqu.SelectDataset) (int64, error) {
	countQuery, countParams, err := baseStmt.Select(goqu.L("COUNT(*) as total")).ToSQL()
	if err != nil {
		return 0, fmt.Errorf("%w: %w", queryErr, err)
	}
	var totalCount int64
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "ListCount", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &totalCount, countQuery, countParams...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil
		}
		return 0, fmt.Errorf("%w: %w", dbErr, err)
	}
	return totalCount, nil
}

func (r UserPATRepository) applySort(query *goqu.SelectDataset, rqlQuery *rql.Query) *goqu.SelectDataset {
	if len(rqlQuery.Sort) > 0 {
		for _, sortItem := range rqlQuery.Sort {
			switch sortItem.Order {
			case "desc":
				query = query.OrderAppend(goqu.C(sortItem.Name).Desc())
			default:
				query = query.OrderAppend(goqu.C(sortItem.Name).Asc())
			}
		}
	} else {
		query = query.Order(goqu.C("created_at").Desc())
	}
	return query
}

func (r UserPATRepository) IsTitleAvailable(ctx context.Context, userID, orgID, title string) (bool, error) {
	query, params, err := dialect.Select(goqu.L("NOT EXISTS(?)",
		dialect.From(TABLE_USER_PATS).Select(goqu.L("1")).Where(
			goqu.Ex{"user_id": userID},
			goqu.Ex{"org_id": orgID},
			goqu.Func("LOWER", goqu.C("title")).Eq(goqu.Func("LOWER", title)),
			goqu.Ex{"deleted_at": nil},
		).Limit(1),
	)).ToSQL()
	if err != nil {
		return false, fmt.Errorf("%w: %w", queryErr, err)
	}

	var available bool
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "IsTitleAvailable", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &available, query, params...)
	}); err != nil {
		return false, fmt.Errorf("%w: %w", dbErr, err)
	}

	return available, nil
}

func (r UserPATRepository) GetBySecretHash(ctx context.Context, secretHash string) (models.PAT, error) {
	query, params, err := dialect.From(TABLE_USER_PATS).
		Select(&UserPAT{}).
		Where(
			goqu.Ex{"secret_hash": secretHash},
			goqu.Ex{"deleted_at": nil},
		).Limit(1).ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var model UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "GetBySecretHash", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &model, query, params...)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.PAT{}, paterrors.ErrNotFound
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) UpdateUsedAt(ctx context.Context, id string, at time.Time) error {
	query, params, err := dialect.Update(TABLE_USER_PATS).
		Set(goqu.Record{"used_at": at}).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "UpdateUsedAt", func(ctx context.Context) error {
		_, err := r.dbc.ExecContext(ctx, query, params...)
		return err
	}); err != nil {
		return fmt.Errorf("%w: %w", dbErr, err)
	}

	return nil
}

func (r UserPATRepository) Update(ctx context.Context, pat models.PAT) (models.PAT, error) {
	marshaledMetadata, err := json.Marshal(pat.Metadata)
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", parseErr, err)
	}

	query, params, err := dialect.Update(TABLE_USER_PATS).
		Set(goqu.Record{
			"title":    pat.Title,
			"metadata": marshaledMetadata,
		}).
		Where(
			goqu.Ex{"id": pat.ID},
			goqu.Ex{"deleted_at": nil},
		).
		Returning(&UserPAT{}).
		ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var model UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "Update", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&model)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.PAT{}, paterrors.ErrNotFound
		}
		if errors.Is(err, ErrDuplicateKey) {
			return models.PAT{}, paterrors.ErrConflict
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) Regenerate(ctx context.Context, id, secretHash string, expiresAt time.Time) (models.PAT, error) {
	query, params, err := dialect.Update(TABLE_USER_PATS).
		Set(goqu.Record{
			"secret_hash":    secretHash,
			"expires_at":     expiresAt,
			"regenerated_at": time.Now().UTC(),
		}).
		Where(
			goqu.Ex{"id": id},
			goqu.Ex{"deleted_at": nil},
		).
		Returning(&UserPAT{}).
		ToSQL()
	if err != nil {
		return models.PAT{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var model UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "Regenerate", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&model)
	}); err != nil {
		err = checkPostgresError(err)
		if errors.Is(err, sql.ErrNoRows) {
			return models.PAT{}, paterrors.ErrNotFound
		}
		if errors.Is(err, ErrDuplicateKey) {
			return models.PAT{}, paterrors.ErrConflict
		}
		return models.PAT{}, fmt.Errorf("%w: %w", dbErr, err)
	}

	return model.transform()
}

func (r UserPATRepository) ListExpiryReminderPending(ctx context.Context, days int) ([]models.PAT, error) {
	query, params, err := dialect.From(TABLE_USER_PATS).Where(
		goqu.Ex{"deleted_at": nil},
		goqu.L("expires_at > NOW()"),
		goqu.L("expires_at <= NOW() + make_interval(days => ?)", days),
		goqu.L("(metadata->>'expiry_reminder_sent_at') IS NULL"),
	).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", queryErr, err)
	}
	var rows []UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "ListExpiryReminderPending", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &rows, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", dbErr, err)
	}
	var pats []models.PAT
	for _, m := range rows {
		pat, err := m.transform()
		if err != nil {
			return nil, err
		}
		pats = append(pats, pat)
	}
	return pats, nil
}

func (r UserPATRepository) ListExpiredNoticePending(ctx context.Context) ([]models.PAT, error) {
	query, params, err := dialect.From(TABLE_USER_PATS).Where(
		goqu.Ex{"deleted_at": nil},
		goqu.L("expires_at < NOW()"),
		goqu.L("(metadata->>'expired_notice_sent_at') IS NULL"),
	).ToSQL()
	if err != nil {
		return nil, fmt.Errorf("%w: %w", queryErr, err)
	}
	var rows []UserPAT
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "ListExpiredNoticePending", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &rows, query, params...)
	}); err != nil {
		return nil, fmt.Errorf("%w: %w", dbErr, err)
	}
	var pats []models.PAT
	for _, m := range rows {
		pat, err := m.transform()
		if err != nil {
			return nil, err
		}
		pats = append(pats, pat)
	}
	return pats, nil
}

func (r UserPATRepository) SetAlertSentMetadata(ctx context.Context, id string, key string) error {
	now := time.Now().UTC().Format(time.RFC3339)
	query, params, err := dialect.Update(TABLE_USER_PATS).
		Set(goqu.Record{
			"metadata": goqu.L("jsonb_set(COALESCE(metadata, '{}'), ?::text[], to_jsonb(?::text))",
				fmt.Sprintf("{%s}", key), now),
		}).
		Where(goqu.Ex{"id": id}).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", queryErr, err)
	}
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "SetAlertSentMetadata", func(ctx context.Context) error {
		_, execErr := r.dbc.ExecContext(ctx, query, params...)
		return execErr
	}); err != nil {
		return fmt.Errorf("%w: %w", dbErr, err)
	}
	return nil
}

func (r UserPATRepository) Delete(ctx context.Context, id string) error {
	query, params, err := dialect.Update(TABLE_USER_PATS).
		Set(goqu.Record{"deleted_at": time.Now().UTC()}).
		Where(
			goqu.Ex{"id": id},
			goqu.Ex{"deleted_at": nil},
		).
		ToSQL()
	if err != nil {
		return fmt.Errorf("%w: %w", queryErr, err)
	}

	var result sql.Result
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_PATS, "Delete", func(ctx context.Context) error {
		var execErr error
		result, execErr = r.dbc.ExecContext(ctx, query, params...)
		return execErr
	}); err != nil {
		return fmt.Errorf("%w: %w", dbErr, err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("%w: %w", dbErr, err)
	}
	if rowsAffected == 0 {
		return paterrors.ErrNotFound
	}

	return nil
}
