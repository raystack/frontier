package postgres

import (
	"errors"

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
)

func NewStore(db *sql.SQL) Store {
	return Store{
		DB: db,
	}
}
