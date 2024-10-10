package sqlcommenter

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"testing"
)

func TestWrapDriver(t *testing.T) {
	cases := []struct {
		name    string
		makeCtx func() context.Context
		options []Option
		perform func(*testing.T, context.Context, *sql.DB)
		assert  func(*testing.T, *mockConn)
	}{
		{
			name: "QueryContext no attrs",
			perform: func(t *testing.T, ctx context.Context, db *sql.DB) {
				_, _ = db.QueryContext(ctx, "SELECT 1")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertQueryContext(t, "SELECT 1", 0)
			},
		},
		{
			name:    "QueryContext with attrs",
			options: []Option{WithAttrPairs("key", "value"), WithAttrPairs("key2", "value 2")},
			perform: func(t *testing.T, ctx context.Context, db *sql.DB) {
				_, _ = db.QueryContext(ctx, "SELECT 1")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertQueryContext(t, "SELECT 1 /*key='value',key2='value%202'*/", 0)
			},
		},
		{
			name: "QueryContext attrs from context",
			options: []Option{WithAttrFunc(func(ctx context.Context) Attrs {
				return AttrPairs("user-key", userKeyFromContext(ctx))
			})},
			makeCtx: func() context.Context {
				return withUserKey(context.Background(), "my-key")
			},
			perform: func(t *testing.T, ctx context.Context, db *sql.DB) {
				_, _ = db.QueryContext(ctx, "SELECT 1")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertQueryContext(t, "SELECT 1 /*user-key='my-key'*/", 0)
			},
		},
		{
			name: "ExecContext no attrs",
			perform: func(t *testing.T, ctx context.Context, db *sql.DB) {
				_, _ = db.ExecContext(ctx, "UPDATE users SET name = 'joe'")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertExecContext(t, "UPDATE users SET name = 'joe'", 0)
			},
		},
		{
			name:    "ExecContext with attrs",
			options: []Option{WithAttrPairs("key", "value")},
			perform: func(t *testing.T, ctx context.Context, db *sql.DB) {
				_, _ = db.ExecContext(ctx, "UPDATE users SET name = 'joe'")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertExecContext(t, "UPDATE users SET name = 'joe' /*key='value'*/", 0)
			},
		},
		{
			name: "ExecContext attrs from context",
			options: []Option{WithAttrFunc(func(ctx context.Context) Attrs {
				return AttrPairs("user-key", userKeyFromContext(ctx))
			})},
			makeCtx: func() context.Context {
				return withUserKey(context.Background(), "my-key")
			},
			perform: func(t *testing.T, ctx context.Context, db *sql.DB) {
				_, _ = db.ExecContext(ctx, "UPDATE users SET name = 'joe'")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertExecContext(t, "UPDATE users SET name = 'joe' /*user-key='my-key'*/", 0)
			},
		},
		{
			name: "QueryContext attrs from context in transaction",
			options: []Option{WithAttrFunc(func(ctx context.Context) Attrs {
				return AttrPairs("user-key", userKeyFromContext(ctx))
			})},
			makeCtx: func() context.Context {
				return withUserKey(context.Background(), "my-key")
			},
			perform: func(t *testing.T, ctx context.Context, db *sql.DB) {
				tx, err := db.Begin()
				assertNoError(t, err)
				defer func() { _ = tx.Commit() }()

				_, _ = tx.QueryContext(ctx, "SELECT 1")
				_, _ = tx.QueryContext(ctx, "SELECT 2")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertQueryContext(t, "SELECT 1 /*user-key='my-key'*/", 0)
				conn.assertQueryContext(t, "SELECT 2 /*user-key='my-key'*/", 1)
			},
		},
		{
			name: "ExecContext attrs from context in transaction",
			options: []Option{WithAttrFunc(func(ctx context.Context) Attrs {
				return AttrPairs("user-key", userKeyFromContext(ctx))
			})},
			makeCtx: func() context.Context {
				return withUserKey(context.Background(), "my-key")
			},
			perform: func(t *testing.T, ctx context.Context, db *sql.DB) {
				tx, err := db.Begin()
				assertNoError(t, err)
				defer func() { _ = tx.Commit() }()

				_, _ = tx.ExecContext(ctx, "UPDATE users SET name = 'joe'")
				_, _ = tx.ExecContext(ctx, "UPDATE users SET name = 'doe'")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertExecContext(t, "UPDATE users SET name = 'joe' /*user-key='my-key'*/", 0)
				conn.assertExecContext(t, "UPDATE users SET name = 'doe' /*user-key='my-key'*/", 1)
			},
		},
	}

	drivers := []struct {
		name      string
		newDriver func(conn *mockConn) driver.Driver
	}{
		{
			name: "driver",
			newDriver: func(conn *mockConn) driver.Driver {
				return &mockDriver{conn: conn}
			},
		},
		{
			name: "driverctx",
			newDriver: func(conn *mockConn) driver.Driver {
				return &mockDriverContext{conn: conn}
			},
		},
	}

	for i, cs := range cases {
		for j, drv := range drivers {
			t.Run(cs.name+" "+drv.name, func(t *testing.T) {
				var ctx context.Context
				if cs.makeCtx != nil {
					ctx = cs.makeCtx()
				} else {
					ctx = context.Background()
				}

				conn := &mockConn{}
				orig := drv.newDriver(conn)
				drv := WrapDriver(orig, cs.options...)

				driverName := fmt.Sprintf("driver-%v-%v", i, j)
				sql.Register(driverName, drv)

				db, err := sql.Open(driverName, "")
				assertNoError(t, err)
				defer db.Close()

				cs.perform(t, ctx, db)
				cs.assert(t, conn)
			})
		}
	}
}

type contextKey int

const contextUserKey contextKey = 0

func withUserKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, contextUserKey, key)
}

func userKeyFromContext(ctx context.Context) string {
	return ctx.Value(contextUserKey).(string)
}

type mockDriverContext struct {
	conn *mockConn
}

func (m *mockDriverContext) Open(name string) (driver.Conn, error) {
	return m.conn, nil
}

func (m *mockDriverContext) OpenConnector(name string) (driver.Connector, error) {
	return &mockConnector{
		drv:  m,
		conn: m.conn,
	}, nil
}

type mockConn struct {
	driver.Conn
	execContext  []string
	queryContext []string
}

func (m *mockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	m.queryContext = append(m.queryContext, query)
	return &mockRows{}, nil
}

func (m *mockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	m.execContext = append(m.execContext, query)
	return nil, nil
}

func (m *mockConn) Begin() (driver.Tx, error) {
	return &mockTx{}, nil
}

func (m *mockConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	return &mockTx{}, nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) assertQueryContext(t *testing.T, query string, idx int) {
	t.Helper()

	if idx+1 > len(m.queryContext) {
		t.Errorf("invalid idx '%v' from '%v'", idx, len(m.execContext))
	}

	for i, q := range m.queryContext {
		if i == idx {
			if q != query {
				t.Errorf("got '%v', want '%v'", m.queryContext, query)
			}
		}
	}
}

func (m *mockConn) assertExecContext(t *testing.T, query string, idx int) {
	t.Helper()

	if idx+1 > len(m.execContext) {
		t.Errorf("invalid idx '%v' from '%v'", idx, len(m.execContext))
	}

	for i, q := range m.execContext {
		if i == idx {
			if q != query {
				t.Errorf("got '%v', want '%v'", m.queryContext, query)
			}
		}
	}
}

func assertNoError(t *testing.T, err error) {
	t.Helper()

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

type mockConnector struct {
	drv  *mockDriverContext
	conn *mockConn
}

func (m *mockConnector) Connect(ctx context.Context) (driver.Conn, error) {
	return m.conn, nil
}

func (m *mockConnector) Driver() driver.Driver {
	return m.drv
}

type mockDriver struct {
	conn *mockConn
}

func (m *mockDriver) Open(name string) (driver.Conn, error) {
	return m.conn, nil
}

type mockTx struct{}

func (m *mockTx) Commit() error {
	return nil
}

func (m *mockTx) Rollback() error {
	return nil
}

type mockRows struct {
}

func (m *mockRows) Columns() []string {
	return nil
}

func (m *mockRows) Close() error {
	return nil
}

func (m *mockRows) Next(dest []driver.Value) error {
	return nil
}
