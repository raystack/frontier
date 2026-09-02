package db

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"io"
	"sync"
	"testing"
	"time"

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
		rollbackErr := errors.New("connection lost")
		conn := &fakeConn{rollbackErr: rollbackErr}
		client := newFakeClient(t, conn)

		err := client.WithTxn(context.Background(), sql.TxOptions{}, func(*sqlx.Tx) error {
			return callbackErr
		})

		assert.EqualError(t, err, "rollback error: connection lost while executing: insert failed")
		assert.ErrorIs(t, err, callbackErr)
		assert.ErrorIs(t, err, rollbackErr)
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

// lockConn is a fake connection for the advisory lock tests. It answers the
// try-lock query with a fixed result and records every query together with
// the id of the connection that ran it, so tests can check that lock and
// unlock happen on the same session. It refuses to run queries on a canceled
// context, like a real driver.
type lockConn struct {
	id       int
	rec      *queryRecorder
	acquired bool
}

func (c *lockConn) Prepare(string) (driver.Stmt, error) { return nil, errors.New("not implemented") }
func (c *lockConn) Close() error                        { c.rec.closed(c.id); return nil }
func (c *lockConn) Begin() (driver.Tx, error)           { return nil, errors.New("not implemented") }

func (c *lockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := c.rec.takeFailure(); err != nil {
		return nil, err
	}
	_, hasDeadline := ctx.Deadline()
	c.rec.add(c.id, query, args, hasDeadline)
	return &boolRows{value: c.acquired}, nil
}

// boolRows is a result set with a single row holding a single boolean column.
type boolRows struct {
	value bool
	done  bool
}

func (r *boolRows) Columns() []string { return []string{"acquired"} }
func (r *boolRows) Close() error      { return nil }
func (r *boolRows) Next(dest []driver.Value) error {
	if r.done {
		return io.EOF
	}
	r.done = true
	dest[0] = r.value
	return nil
}

type recordedQuery struct {
	connID      int
	query       string
	arg         driver.Value
	hasDeadline bool
}

type queryRecorder struct {
	mu       sync.Mutex
	queries  []recordedQuery
	closes   []int
	failNext error
}

func (r *queryRecorder) add(connID int, query string, args []driver.NamedValue, hasDeadline bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	q := recordedQuery{connID: connID, query: query, hasDeadline: hasDeadline}
	if len(args) > 0 {
		q.arg = args[0].Value
	}
	r.queries = append(r.queries, q)
}

func (r *queryRecorder) all() []recordedQuery {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]recordedQuery(nil), r.queries...)
}

func (r *queryRecorder) closed(connID int) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.closes = append(r.closes, connID)
}

func (r *queryRecorder) closedConns() []int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]int(nil), r.closes...)
}

func (r *queryRecorder) failNextQuery(err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.failNext = err
}

func (r *queryRecorder) takeFailure() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	err := r.failNext
	r.failNext = nil
	return err
}

// lockConnector hands out a fresh numbered connection on every Connect call,
// the way a real pool dials new sessions.
type lockConnector struct {
	rec      *queryRecorder
	acquired bool
	mu       sync.Mutex
	next     int
}

func (f *lockConnector) Connect(context.Context) (driver.Conn, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.next++
	return &lockConn{id: f.next, rec: f.rec, acquired: f.acquired}, nil
}
func (f *lockConnector) Driver() driver.Driver { return nil }

func newLockClient(t *testing.T, acquired bool) (Client, *queryRecorder) {
	t.Helper()
	rec := &queryRecorder{}
	client := Client{DB: sqlx.NewDb(sql.OpenDB(&lockConnector{rec: rec, acquired: acquired}), "postgres")}
	t.Cleanup(func() { _ = client.Close() })
	return client, rec
}

func TestTryLock(t *testing.T) {
	t.Run("acquires and releases on the same connection", func(t *testing.T) {
		client, rec := newLockClient(t, true)

		lock, err := client.TryLock(context.Background(), "some-job")
		require.NoError(t, err)
		require.NotNil(t, lock)
		require.NoError(t, lock.Unlock(context.Background()))

		queries := rec.all()
		require.Len(t, queries, 2)
		assert.Contains(t, queries[0].query, "pg_try_advisory_lock")
		assert.Contains(t, queries[1].query, "pg_advisory_unlock")
		assert.Equal(t, queries[0].connID, queries[1].connID)
		assert.Equal(t, queries[0].arg, queries[1].arg)
	})

	t.Run("returns ErrLockBusy when the lock is already held", func(t *testing.T) {
		client, _ := newLockClient(t, false)

		lock, err := client.TryLock(context.Background(), "some-job")
		assert.ErrorIs(t, err, ErrLockBusy)
		assert.Nil(t, lock)
	})
}

func TestUnlock(t *testing.T) {
	t.Run("releases even when the caller's context is canceled", func(t *testing.T) {
		client, rec := newLockClient(t, true)

		lock, err := client.TryLock(context.Background(), "some-job")
		require.NoError(t, err)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		require.NoError(t, lock.Unlock(ctx))

		queries := rec.all()
		require.Len(t, queries, 2)
		assert.Contains(t, queries[1].query, "pg_advisory_unlock")
	})

	t.Run("reports when the session does not hold the lock", func(t *testing.T) {
		client, _ := newLockClient(t, false)

		conn, err := client.Connx(context.Background())
		require.NoError(t, err)
		lock := &Lock{ID: 42, conn: conn}

		assert.ErrorIs(t, lock.Unlock(context.Background()), ErrLockNotHeld)
	})

	t.Run("discards the connection when the release fails", func(t *testing.T) {
		client, rec := newLockClient(t, true)

		lock, err := client.TryLock(context.Background(), "some-job")
		require.NoError(t, err)

		queryErr := errors.New("network down")
		rec.failNextQuery(queryErr)
		assert.ErrorIs(t, lock.Unlock(context.Background()), queryErr)

		queries := rec.all()
		require.NotEmpty(t, queries)
		assert.Contains(t, rec.closedConns(), queries[0].connID)
	})

	t.Run("applies the client's query timeout to the release", func(t *testing.T) {
		rec := &queryRecorder{}
		client := Client{
			DB:           sqlx.NewDb(sql.OpenDB(&lockConnector{rec: rec, acquired: true}), "postgres"),
			queryTimeOut: time.Second,
		}
		t.Cleanup(func() { _ = client.Close() })

		lock, err := client.TryLock(context.Background(), "some-job")
		require.NoError(t, err)
		require.NoError(t, lock.Unlock(context.Background()))

		queries := rec.all()
		require.Len(t, queries, 2)
		assert.Contains(t, queries[1].query, "pg_advisory_unlock")
		assert.True(t, queries[1].hasDeadline)
	})
}
