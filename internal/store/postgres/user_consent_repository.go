package postgres

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"

	"github.com/raystack/frontier/core/consent"
	"github.com/raystack/frontier/pkg/db"
)

// UserConsentRepository writes consent records and nothing else: the table has
// BEFORE UPDATE and BEFORE DELETE triggers, so there is no other operation to
// offer. Reads are left to reporting tools, which query the table directly.
type UserConsentRepository struct {
	dbc *db.Client
}

func NewUserConsentRepository(dbc *db.Client) *UserConsentRepository {
	return &UserConsentRepository{
		dbc: dbc,
	}
}

// Create writes one record inside the transaction it is given rather than
// opening its own, because it has to land with the user row or not at all.
// Rolling back is the caller's job: the transaction is wider than this insert.
func (r UserConsentRepository) Create(ctx context.Context, tx *sqlx.Tx, cnst consent.Consent) (consent.Consent, error) {
	// a nil transaction is a wiring mistake, and a panic is a poor way to report it
	if tx == nil {
		return consent.Consent{}, fmt.Errorf("%w: no transaction", consent.ErrInvalidGrant)
	}

	documents, err := marshalConsentDocuments(cnst.Documents)
	if err != nil {
		return consent.Consent{}, fmt.Errorf("%w: %w", errParse, err)
	}

	createQuery, params, err := dialect.Insert(TABLE_USER_CONSENTS).Rows(UserConsent{
		UserID:    cnst.UserID,
		UserEmail: cnst.UserEmail,
		Documents: documents,
		Source:    cnst.Source,
		// both nullable: a deployment that sets no client IP header stores no IP
		// rather than failing the signup
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
			// the partial unique index: a second signup write is a bug
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
