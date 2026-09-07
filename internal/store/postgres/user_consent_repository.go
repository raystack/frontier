package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/raystack/frontier/core/consent"
	"github.com/raystack/frontier/pkg/db"
)

// UserConsentRepository only writes: the table has BEFORE UPDATE and BEFORE
// DELETE triggers, so there is no other operation to offer.
type UserConsentRepository struct {
	dbc *db.Client
}

func NewUserConsentRepository(dbc *db.Client) *UserConsentRepository {
	return &UserConsentRepository{
		dbc: dbc,
	}
}

// Create writes one record inside the transaction it is given: it has to land
// with the user row or not at all, so rolling back is the caller's job.
func (r UserConsentRepository) Create(ctx context.Context, tx *sqlx.Tx, cnst consent.Consent) (consent.Consent, error) {
	if tx == nil {
		return consent.Consent{}, fmt.Errorf("%w: no transaction", consent.ErrInvalidGrant)
	}

	documents, err := marshalConsentDocuments(cnst.Documents)
	if err != nil {
		return consent.Consent{}, fmt.Errorf("%w: %w", errParse, err)
	}

	createQuery, params, err := dialect.Insert(TABLE_USER_CONSENTS).Rows(UserConsent{
		UserID:       cnst.UserID,
		UserEmail:    cnst.UserEmail,
		Documents:    documents,
		Source:       cnst.Source,
		AuthStrategy: toNullString(cnst.AuthStrategy),
		IPAddress:    toNullString(cnst.IPAddress),
		ConsentedAt:  cnst.ConsentedAt,
	}).Returning(&UserConsent{}).ToSQL()
	if err != nil {
		return consent.Consent{}, fmt.Errorf("%w: %w", errQuery, err)
	}

	var consentModel UserConsent
	if err = r.dbc.WithTimeout(ctx, TABLE_USER_CONSENTS, "Create", func(ctx context.Context) error {
		return tx.QueryRowxContext(ctx, createQuery, params...).StructScan(&consentModel)
	}); err != nil {
		err = checkPostgresError(err)
		switch {
		case errors.Is(err, ErrDuplicateKey):
			return consent.Consent{}, consent.ErrConsentExists
		default:
			return consent.Consent{}, fmt.Errorf("%w: %w", errDB, err)
		}
	}

	transformedConsent, err := consentModel.transformToConsent()
	if err != nil {
		return consent.Consent{}, fmt.Errorf("%w: %w", errParse, err)
	}
	return transformedConsent, nil
}
