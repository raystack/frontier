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
	dbErr    = errors.New("error while running query")
	txnErr   = errors.New("error while running transaction")
)

func NewStore(db *sql.SQL) Store {
	return Store{
		DB: db,
	}
}

//func (s Store) WithTx(ctx context.Context, txFunc func(context.Context) error) (err error) {
//	txOptions := &sql.TxOptions{
//		Isolation: sql.LevelReadCommitted,
//	}
//	return db.Transact(ctx, s.DB, txOptions, txFunc)
//}
