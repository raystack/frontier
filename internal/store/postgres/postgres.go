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
	TABLE_ACTIONS       = "actions"
	TABLE_GROUPS        = "groups"
	TABLE_NAMESPACES    = "namespaces"
	TABLE_ORGANIZATIONS = "organizations"
	TABLE_POLICIES      = "policies"
	TABLE_PROJECTS      = "projects"
	TABLE_RELATIONS     = "relations"
	TABLE_RESOURCES     = "resources"
	TABLE_ROLES         = "roles"
	TABLE_USERS         = "users"
	TABLE_METADATA      = "metadata"
	TABLE_METADATA_KEYS = "metadata_keys"
	TABLE_FLOWS         = "flows"
	TABLE_SESSIONS      = "sessions"
)

func checkPostgresError(err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			return fmt.Errorf("%w [%s]", errDuplicateKey, pgErr.Detail)
		case pgerrcode.CheckViolation:
			return fmt.Errorf("%w [%s]", errCheckViolation, pgErr.Detail)
		case pgerrcode.ForeignKeyViolation:
			return fmt.Errorf("%w [%s]", errForeignKeyViolation, pgErr.Detail)
		case pgerrcode.InvalidTextRepresentation:
			return fmt.Errorf("%w [%s]", errInvalidTexRepresentation, pgErr.Detail)
		}
	}
	return err
}
