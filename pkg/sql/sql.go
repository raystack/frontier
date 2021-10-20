package sql

import (
	"context"
	"time"

	"github.com/jmoiron/sqlx"
)

type Config struct {
	Driver          string
	URL             string
	MaxIdleConns    int
	MaxOpenConns    int
	ConnMaxLifeTime time.Duration
}

func New(config Config) (*sqlx.DB, error) {
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

	return d, err
}

func WithTimeout(ctx context.Context, timeout time.Duration, op func(ctx context.Context) error) (err error) {
	ctxWithTimeout, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return op(ctxWithTimeout)
}
