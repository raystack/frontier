package postgres

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/raystack/frontier/pkg/db"
)

// txnFailClient returns a client whose connections always fail, so any
// transaction begin returns an error.
func txnFailClient(t *testing.T) *db.Client {
	t.Helper()
	client := &db.Client{DB: sqlx.NewDb(sql.OpenDB(failConnector{}), "postgres")}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

type failConnector struct{}

func (failConnector) Connect(context.Context) (driver.Conn, error) {
	return nil, errors.New("connection failed")
}

func (failConnector) Driver() driver.Driver { return nil }
