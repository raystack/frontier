package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeConn is a minimal database/sql/driver connection, in the same style as
// failConnector in internal/store/postgres/fakes_test.go. It hands out fake
// transactions, records commits and rollbacks, and can be told to fail any of
// them, so WithTxn can be tested without a real database.
type fakeConn struct {
	beginErr    error
	commitErr   error
	rollbackErr error
	commits     int
	rollbacks   int
}

func (c *fakeConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not implemented") }
func (c *fakeConn) Close() error                        { return nil }
func (c *fakeConn) Begin() (driver.Tx, error) {
	if c.beginErr != nil {
		return nil, c.beginErr
	}
	return &fakeTx{conn: c}, nil
}

type fakeTx struct{ conn *fakeConn }

func (t *fakeTx) Commit() error   { t.conn.commits++; return t.conn.commitErr }
func (t *fakeTx) Rollback() error { t.conn.rollbacks++; return t.conn.rollbackErr }

type fakeConnector struct{ conn *fakeConn }

func (f fakeConnector) Connect(context.Context) (driver.Conn, error) { return f.conn, nil }
func (f fakeConnector) Driver() driver.Driver                        { return nil }

func newFakeClient(t *testing.T, conn *fakeConn) Client {
	t.Helper()
	client := Client{DB: sqlx.NewDb(sql.OpenDB(fakeConnector{conn: conn}), "postgres")}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestWithTxn(t *testing.T) {
	t.Run("commits when the callback succeeds", func(t *testing.T) {
		conn := &fakeConn{}
		client := newFakeClient(t, conn)

		err := client.WithTxn(context.Background(), sql.TxOptions{}, func(*sqlx.Tx) error {
			return nil
		})

		require.NoError(t, err)
		assert.Equal(t, 1, conn.commits)
		assert.Equal(t, 0, conn.rollbacks)
	})

	t.Run("returns the begin error without running the callback", func(t *testing.T) {
		beginErr := errors.New("begin failed")
		client := newFakeClient(t, &fakeConn{beginErr: beginErr})

		called := false
		err := client.WithTxn(context.Background(), sql.TxOptions{}, func(*sqlx.Tx) error {
			called = true
			return nil
		})

		assert.ErrorIs(t, err, beginErr)
		assert.False(t, called)
	})

	t.Run("wraps the callback error once when rollback succeeds", func(t *testing.T) {
		callbackErr := errors.New("insert failed")
		conn := &fakeConn{}
		client := newFakeClient(t, conn)

		err := client.WithTxn(context.Background(), sql.TxOptions{}, func(*sqlx.Tx) error {
			return callbackErr
		})

		assert.EqualError(t, err, "rollback: insert failed")
		assert.ErrorIs(t, err, callbackErr)
		assert.Equal(t, 1, conn.rollbacks)
		assert.Equal(t, 0, conn.commits)
	})

	t.Run("reports both errors when rollback fails", func(t *testing.T) {
		callbackErr := errors.New("insert failed")
		conn := &fakeConn{rollbackErr: errors.New("connection lost")}
		client := newFakeClient(t, conn)

		err := client.WithTxn(context.Background(), sql.TxOptions{}, func(*sqlx.Tx) error {
			return callbackErr
		})

		assert.EqualError(t, err, "rollback error: connection lost while executing: insert failed")
		assert.ErrorIs(t, err, callbackErr)
		assert.Equal(t, 1, conn.rollbacks)
	})

	t.Run("returns the commit error", func(t *testing.T) {
		commitErr := errors.New("commit failed")
		conn := &fakeConn{commitErr: commitErr}
		client := newFakeClient(t, conn)

		err := client.WithTxn(context.Background(), sql.TxOptions{}, func(*sqlx.Tx) error {
			return nil
		})

		assert.ErrorIs(t, err, commitErr)
		assert.Equal(t, 1, conn.commits)
	})

	t.Run("rolls back and repanics when the callback panics", func(t *testing.T) {
		conn := &fakeConn{}
		client := newFakeClient(t, conn)

		defer func() {
			p := recover()
			require.NotNil(t, p)
			assert.Equal(t, "boom", p)
			assert.Equal(t, 1, conn.rollbacks)
			assert.Equal(t, 0, conn.commits)
		}()

		_ = client.WithTxn(context.Background(), sql.TxOptions{}, func(*sqlx.Tx) error {
			panic("boom")
		})
	})
}
