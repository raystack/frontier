package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/raystack/frontier/pkg/db"
)

// txnFailClient returns a client whose connections always fail, so any
// transaction begin returns an error.
func txnFailClient() *db.Client {
	return &db.Client{DB: sqlx.NewDb(sql.OpenDB(failConnector{}), "postgres")}
}

type failConnector struct{}

func (failConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, errors.New("connection failed")
}

func (failConnector) Driver() driver.Driver { return nil }
