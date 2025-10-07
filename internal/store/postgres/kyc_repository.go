package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/doug-martin/goqu/v9"
	"github.com/jmoiron/sqlx"
	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/pkg/db"
)

const (
	COLUMN_STATUS = "status"
	COLUMN_LINK   = "link"

	// Table column references
	ORG_ID         = TABLE_ORGANIZATIONS + "." + COLUMN_ID
	ORG_KYC_ID     = TABLE_ORGANIZATIONS_KYC + "." + COLUMN_ORG_ID
	KYC_STATUS     = TABLE_ORGANIZATIONS_KYC + "." + COLUMN_STATUS
	KYC_LINK       = TABLE_ORGANIZATIONS_KYC + "." + COLUMN_LINK
	KYC_CREATED_AT = TABLE_ORGANIZATIONS_KYC + "." + COLUMN_CREATED_AT
	KYC_UPDATED_AT = TABLE_ORGANIZATIONS_KYC + "." + COLUMN_UPDATED_AT
)

type OrgKycRepository struct {
	dbc *db.Client
}

// Define a struct to hold the joined result
type joinResult struct {
	OrgID     string    `db:"id"`
	Status    bool      `db:"status"`
	Link      string    `db:"link"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

func NewOrgKycRepository(dbc *db.Client) *OrgKycRepository {
	return &OrgKycRepository{
		dbc: dbc,
	}
}

func (r OrgKycRepository) GetByOrgID(ctx context.Context, orgID string) (kyc.KYC, error) {
	query, params, err := dialect.From(TABLE_ORGANIZATIONS_KYC).Where(goqu.Ex{
		"org_id": orgID,
	}).ToSQL()

	if err != nil {
		return kyc.KYC{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var kycModel KYC
	err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS_KYC, "GetByOrgID", func(ctx context.Context) error {
		return r.dbc.GetContext(ctx, &kycModel, query, params...)
	})
	err = checkPostgresError(err)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return kyc.KYC{}, kyc.ErrNotExist
		case errors.Is(err, ErrInvalidTextRepresentation):
			return kyc.KYC{}, kyc.ErrInvalidUUID
		default:
			return kyc.KYC{}, err
		}
	}

	return kycModel.transformToKyc()
}

func (r OrgKycRepository) Upsert(ctx context.Context, input kyc.KYC) (kyc.KYC, error) {
	var query string
	var params []interface{}

	// Struct to hold KYC data + org name
	type kycWithOrgName struct {
		KYC
		OrgName string `db:"org_name"`
	}
	var result kycWithOrgName

	_, err := r.GetByOrgID(ctx, input.OrgID)
	if err == nil {
		// kyc for org exists, prepare UPDATE query with JOIN
		query, params, err = dialect.Update(TABLE_ORGANIZATIONS_KYC).
			From(TABLE_ORGANIZATIONS).
			Set(goqu.Record{
				"status": input.Status,
				"link":   input.Link,
			}).
			Where(
				goqu.Ex{
					TABLE_ORGANIZATIONS_KYC + ".org_id": input.OrgID,
				},
				goqu.Ex{
					TABLE_ORGANIZATIONS + ".id": input.OrgID,
				},
			).
			Returning(
				goqu.I(TABLE_ORGANIZATIONS_KYC+".*"),
				goqu.I(TABLE_ORGANIZATIONS+".name").As("org_name"),
			).ToSQL()
		if err != nil {
			return kyc.KYC{}, fmt.Errorf("%w: %w", queryErr, err)
		}
	} else if err.Error() == kyc.ErrNotExist.Error() {
		// kyc for org doesn't exist, prepare INSERT query with subquery for org name
		orgNameSubquery := dialect.From(TABLE_ORGANIZATIONS).
			Select("name").
			Where(goqu.Ex{"id": input.OrgID})

		query, params, err = dialect.Insert(TABLE_ORGANIZATIONS_KYC).
			Rows(goqu.Record{
				"org_id": input.OrgID,
				"status": input.Status,
				"link":   input.Link,
			}).
			Returning(
				goqu.I(TABLE_ORGANIZATIONS_KYC+".*"),
				orgNameSubquery.As("org_name"),
			).ToSQL()
		if err != nil {
			return kyc.KYC{}, fmt.Errorf("%w: %w", queryErr, err)
		}
	} else if errors.Is(err, kyc.ErrInvalidUUID) {
		// invalid UUID provided
		return kyc.KYC{}, kyc.ErrInvalidUUID
	} else {
		// unexpected error happened in getting org kyc
		return kyc.KYC{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	if err = r.dbc.WithTxn(ctx, sql.TxOptions{}, func(tx *sqlx.Tx) error {
		return r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS_KYC, "Upsert", func(ctx context.Context) error {
			if err := tx.QueryRowxContext(ctx, query, params...).StructScan(&result); err != nil {
				return err
			}

			// Determine event based on status
			event := "kyc.unverified"
			if result.Status {
				event = "kyc.verified"
			}

			auditRecord := BuildAuditRecord(
				ctx,
				event,
				AuditResource{
					ID:   result.OrgID,
					Type: "organization",
					Name: result.OrgName,
				},
				&AuditTarget{
					ID:   result.OrgID,
					Type: "kyc",
					Metadata: map[string]interface{}{
						"status": result.Status,
						"link":   result.Link,
					},
				},
				result.OrgID,
				nil,
				result.UpdatedAt,
			)

			return InsertAuditRecordInTx(ctx, tx, auditRecord)
		})
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrCheckViolation):
			return kyc.KYC{}, kyc.ErrKycLinkNotSet
		case errors.Is(err, ErrForeignKeyViolation):
			return kyc.KYC{}, kyc.ErrOrgDoesntExist
		default:
			return kyc.KYC{}, err
		}
	}

	return result.KYC.transformToKyc()
}

func (r OrgKycRepository) List(ctx context.Context) ([]kyc.KYC, error) {
	// Define table references
	orgs := goqu.T(TABLE_ORGANIZATIONS)
	orgKycs := goqu.T(TABLE_ORGANIZATIONS_KYC)

	// Build query with join condition and COALESCE expressions
	query, params, err := dialect.From(orgs).
		LeftJoin(orgKycs, goqu.On(orgs.Col(COLUMN_ID).Eq(orgKycs.Col(COLUMN_ORG_ID)))).
		Select(
			orgs.Col(COLUMN_ID),
			goqu.COALESCE(orgKycs.Col(COLUMN_STATUS), false).As(COLUMN_STATUS),
			goqu.COALESCE(orgKycs.Col(COLUMN_LINK), "").As(COLUMN_LINK),
			goqu.COALESCE(orgKycs.Col(COLUMN_CREATED_AT), time.Time{}).As(COLUMN_CREATED_AT),
			goqu.COALESCE(orgKycs.Col(COLUMN_UPDATED_AT), time.Time{}).As(COLUMN_UPDATED_AT),
		).Prepared(true).ToSQL()

	if err != nil {
		return nil, fmt.Errorf("%w: %w", queryErr, err)
	}

	var results []joinResult
	err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS_KYC, "OrgKYCs", func(ctx context.Context) error {
		return r.dbc.SelectContext(ctx, &results, query, params...)
	})

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return []kyc.KYC{}, nil
		default:
			return nil, err
		}
	}

	return transformResults(results), nil
}

func transformResults(results []joinResult) []kyc.KYC {
	kycList := make([]kyc.KYC, len(results))
	for i, result := range results {
		kycList[i] = kyc.KYC{OrgID: result.OrgID, Status: result.Status, Link: result.Link, CreatedAt: result.CreatedAt, UpdatedAt: result.UpdatedAt}
	}
	return kycList
}
