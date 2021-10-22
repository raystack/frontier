package sql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type Config struct {
	Driver          string
	URL             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifeTime time.Duration
}

type SQL struct {
	*sqlx.DB
	queryTimeOut time.Duration
}

func New(config Config) (*SQL, error) {
	d, err := sqlx.Open(config.Driver, config.URL)
	if err != nil {
		return nil, err
	}

	if err = d.Ping(); err != nil {
		return nil, err
	}

	d.SetMaxIdleConns(config.MaxIdleConns)
	d.SetMaxOpenConns(config.MaxOpenConns)
	d.SetConnMaxLifetime(config.ConnMaxLifeTime)

	return &SQL{DB: d, queryTimeOut: 10 * time.Millisecond}, err // TODO: queryTimeOut configurable
}

func (s SQL) WithTimeout(ctx context.Context, op func(ctx context.Context) error) (err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.queryTimeOut)
	defer cancel()

	return op(ctxWithTimeout)
}

// Handling transactions: https://stackoverflow.com/a/23502629/8244298
func (s SQL) WithTxn(ctx context.Context, txnOptions sql.TxOptions, txFunc func(*sqlx.Tx) error) (err error) {
	txn, err := s.BeginTxx(ctx, &txnOptions)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = errors.Errorf("%s", p)
			}
			err = txn.Rollback()
			panic(p)
		} else if err != nil {
			if rlbErr := txn.Rollback(); err != nil {
				err = fmt.Errorf("rollback error: %s while executing: %w", rlbErr, err)
			} else {
				err = fmt.Errorf("rollback: %w", err)
			}
			err = fmt.Errorf("rollback: %w", err)
		} else {
			err = txn.Commit()
		}
	}()

	err = txFunc(txn)
	return err
}
