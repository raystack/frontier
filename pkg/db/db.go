package db

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/cespare/xxhash"

	"github.com/raystack/frontier/internal/metrics"

	newrelic "github.com/newrelic/go-agent"

	"github.com/jmoiron/sqlx"
)

var (
	ErrLockBusy = errors.New("lock busy")
)

type Client struct {
	*sqlx.DB
	queryTimeOut time.Duration
}

func New(cfg Config) (*Client, error) {
	// TODO(kushsharma): add tracing support via otelsqlx
	d, err := sqlx.Open(cfg.Driver, cfg.URL)
	if err != nil {
		return nil, err
	}

	if err = d.Ping(); err != nil {
		return nil, err
	}

	d.SetMaxIdleConns(cfg.MaxIdleConns)
	d.SetMaxOpenConns(cfg.MaxOpenConns)
	d.SetConnMaxLifetime(cfg.ConnMaxLifeTime)

	return &Client{DB: d, queryTimeOut: cfg.MaxQueryTimeout}, err
}

func (c Client) WithTimeout(ctx context.Context, collection, operation string, op func(ctx context.Context) error) (err error) {
	nrCtx := newrelic.FromContext(ctx)
	if nrCtx != nil {
		nr := newrelic.DatastoreSegment{
			Product:    newrelic.DatastorePostgres,
			Collection: collection,
			Operation:  operation,
			StartTime:  nrCtx.StartSegmentNow(),
		}
		defer func() {
			_ = nr.End()
		}()
	}
	if metrics.DatabaseQueryLatency != nil {
		promCollect := metrics.DatabaseQueryLatency(collection, operation)
		defer promCollect()
	}

	ctxWithTimeout, cancel := context.WithTimeout(ctx, c.queryTimeOut)
	defer cancel()
	return op(ctxWithTimeout)
}

// WithTxn Handling transactions: https://stackoverflow.com/a/23502629/8244298
func (c Client) WithTxn(ctx context.Context, txnOptions sql.TxOptions, txFunc func(*sqlx.Tx) error) (err error) {
	txn, err := c.BeginTxx(ctx, &txnOptions)
	if err != nil {
		return err
	}

	defer func() {
		if p := recover(); p != nil {
			switch p := p.(type) {
			case error:
				err = p
			default:
				err = fmt.Errorf("%s", p)
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

type Lock struct {
	ID   int64
	conn *sqlx.Conn
}

// Unlock uses postgres advisory locks to release a lock on a given id
func (l Lock) Unlock(ctx context.Context) error {
	var errs []error
	_, err := l.conn.ExecContext(ctx, "SELECT pg_advisory_unlock($1)", l.ID)
	if err != nil {
		errs = append(errs, err)
	}

	err = l.conn.Close()
	if err != nil {
		errs = append(errs, err)
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// TryLock uses postgres advisory locks to acquire a lock on a given id
// if acquired, it returns the Lock object, else fail with ErrLockBusy
// In worst case if not unlocked, it will be released after the session ends
// which is configured via SetConnMaxLifetime
func (c Client) TryLock(ctx context.Context, id string) (*Lock, error) {
	newConn, err := c.Connx(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}

	hash := xxhash.Sum64String(id)
	intHash := int64(hash % uint64(math.MaxInt64)) // Reduce hash to fit within int64 range
	query := "SELECT pg_try_advisory_lock($1)"
	var acquired bool
	if err := c.GetContext(ctx, &acquired, query, intHash); err != nil {
		var errs []error
		errs = append(errs, err)
		if connErr := newConn.Close(); connErr != nil {
			errs = append(errs, fmt.Errorf("failed to close connection: %w", connErr))
		}
		return nil, errors.Join(errs...)
	}

	if !acquired {
		if connErr := newConn.Close(); connErr != nil {
			return nil, fmt.Errorf("failed to close connection: %w", connErr)
		}
		return nil, ErrLockBusy
	}

	lock := &Lock{
		ID:   intHash,
		conn: newConn,
	}
	return lock, nil
}
