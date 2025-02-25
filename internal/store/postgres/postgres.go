package postgres

import (
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/postgres"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgerrcode"
	_ "github.com/jackc/pgx/v4/stdlib"
)

var (
	parseErr = errors.New("parsing error")
	queryErr = errors.New("error while creating the query")
	dbErr    = errors.New("error while running query")
	txnErr   = errors.New("error while running transaction")
	dialect  = goqu.Dialect("postgres")
)

const (
	TABLE_PERMISSIONS            = "permissions"
	TABLE_GROUPS                 = "groups"
	TABLE_NAMESPACES             = "namespaces"
	TABLE_ORGANIZATIONS          = "organizations"
	TABLE_ORGANIZATIONS_KYC      = "organizations_kyc"
	TABLE_POLICIES               = "policies"
	TABLE_PROJECTS               = "projects"
	TABLE_RELATIONS              = "relations"
	TABLE_RESOURCES              = "resources"
	TABLE_ROLES                  = "roles"
	TABLE_USERS                  = "users"
	TABLE_METASCHEMA             = "metaschema"
	TABLE_FLOWS                  = "flows"
	TABLE_SESSIONS               = "sessions"
	TABLE_INVITATIONS            = "invitations"
	TABLE_SERVICEUSER            = "serviceusers"
	TABLE_SERVICEUSERCREDENTIALS = "serviceuser_credentials"
	TABLE_AUDITLOGS              = "auditlogs"
	TABLE_DOMAINS                = "domains"
	TABLE_PREFERENCES            = "preferences"
	TABLE_BILLING_CUSTOMERS      = "billing_customers"
	TABLE_BILLING_PLANS          = "billing_plans"
	TABLE_BILLING_PRODUCTS       = "billing_products"
	TABLE_BILLING_PRICES         = "billing_prices"
	TABLE_BILLING_FEATURES       = "billing_features"
	TABLE_BILLING_SUBSCRIPTIONS  = "billing_subscriptions"
	TABLE_BILLING_CHECKOUTS      = "billing_checkouts"
	TABLE_BILLING_TRANSACTIONS   = "billing_transactions"
	TABLE_BILLING_INVOICES       = "billing_invoices"
	TABLE_WEBHOOK_ENDPOINTS      = "webhook_endpoints"
)

func checkPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%w [%s]", ErrDuplicateKey, pgErr.Detail)
		case pgerrcode.CheckViolation:
			return fmt.Errorf("%w [%s]", ErrCheckViolation, pgErr.Detail)
		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%w [%s]", ErrForeignKeyViolation, pgErr.Detail)
		case pgerrcode.InvalidTextRepresentation:
			return fmt.Errorf("%w: [%s %s]", ErrInvalidTextRepresentation, pgErr.Detail, pgErr.Message)
		}
	}
	return err
}
