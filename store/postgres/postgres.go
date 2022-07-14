package postgres

import (
	"errors"
	"fmt"

	"github.com/doug-martin/goqu/v9"
	"github.com/google/uuid"

	"github.com/odpf/shield/pkg/sql"
)

type Store struct {
	DB *sql.SQL
}

var (
	parseErr = errors.New("parsing error")
	queryErr = errors.New("error while creating the query")
	dbErr    = errors.New("error while running query")
	txnErr   = errors.New("error while running transaction")
	dialect  = goqu.Dialect("postgres")
)

const (
	TABLE_ACTION    = "actions"
	TABLE_GROUPS    = "groups"
	TABLE_NAMESPACE = "namespaces"
	TABLE_ORG       = "organizations"
	TABLE_POLICY    = "policies"
	TABLE_PROJECTS  = "projects"
	TABLE_RELATION  = "relations"
	TABLE_RESOURCE  = "resources"
	TABLE_ROLES     = "roles"
	TABLE_USER      = "users"
)

func NewStore(db *sql.SQL) Store {
	return Store{
		DB: db,
	}
}

func isUUID(key string) bool {
	_, err := uuid.Parse(key)
	fmt.Println(err)
	return err == nil
}
