package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/raystack/frontier/core/kyc"
	"github.com/raystack/frontier/pkg/db"
)

type OrgKycRepository struct {
	dbc *db.Client
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

	var kycModel KYC
	_, err := r.GetByOrgID(ctx, input.OrgID)
	if err == nil {
		//kyc for org exists, prepare UPDATE query
		query, params, err = dialect.Update(TABLE_ORGANIZATIONS_KYC).Set(goqu.Record{
			"status": input.Status,
			"link":   input.Link,
		}).Where(goqu.Ex{"org_id": input.OrgID}).
			Returning(&kycModel).ToSQL()
		if err != nil {
			return kyc.KYC{}, fmt.Errorf("%w: %w", queryErr, err)
		}
	} else if err.Error() == kyc.ErrNotExist.Error() {
		//kyc for org doesn't exist, so we should prepare INSERT query
		query, params, err = dialect.Insert(TABLE_ORGANIZATIONS_KYC).Rows(goqu.Record{
			"org_id": input.OrgID,
			"status": input.Status,
			"link":   input.Link,
		}).Returning(&kycModel).ToSQL()
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

	err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS_KYC, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&kycModel)
	})

	if err != nil {
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

	return kycModel.transformToKyc()
}
