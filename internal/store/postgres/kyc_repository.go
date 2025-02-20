package postgres

import (
	"context"
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
	if err != nil {
		return kyc.KYC{}, err
	}
	return kycModel.transformToKyc()
}

func (r OrgKycRepository) Upsert(ctx context.Context, input kyc.KYC) (kyc.KYC, error) {
	insertRow := goqu.Record{
		"org_id": input.OrgID,
		"status": input.Status,
		"link":   input.Link,
	}
	query, params, err := dialect.Insert(TABLE_ORGANIZATIONS_KYC).Rows(insertRow).Returning(&KYC{}).ToSQL()
	if err != nil {
		return kyc.KYC{}, fmt.Errorf("%w: %w", queryErr, err)
	}

	var kycModel KYC

	err = r.dbc.WithTimeout(ctx, TABLE_ORGANIZATIONS_KYC, "Upsert", func(ctx context.Context) error {
		return r.dbc.QueryRowxContext(ctx, query, params...).StructScan(&kycModel)
	})

	if err != nil {
		return kyc.KYC{}, err
	}
	return kycModel.transformToKyc()
}
