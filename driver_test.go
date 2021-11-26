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
		perform func(context.Context, *sql.DB)
		assert  func(*testing.T, *mockConn)
	}{
		{
			name: "QueryContext no attrs",
			perform: func(ctx context.Context, db *sql.DB) {
				db.QueryContext(ctx, "SELECT 1")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertQueryContext(t, "SELECT 1")
			},
		},
		{
			name:    "QueryContext with attrs",
			options: []Option{WithAttrPairs("key", "value"), WithAttrPairs("key2", "value 2")},
			perform: func(ctx context.Context, db *sql.DB) {
				db.QueryContext(ctx, "SELECT 1")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertQueryContext(t, "SELECT 1 /* key='value',key2='value%202' */")
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
			perform: func(ctx context.Context, db *sql.DB) {
				db.QueryContext(ctx, "SELECT 1")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertQueryContext(t, "SELECT 1 /* user-key='my-key' */")
			},
		},
		{
			name: "ExecContext no attrs",
			perform: func(ctx context.Context, db *sql.DB) {
				db.ExecContext(ctx, "UPDATE users SET name = 'joe'")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertExecContext(t, "UPDATE users SET name = 'joe'")
			},
		},
		{
			name:    "ExecContext with attrs",
			options: []Option{WithAttrPairs("key", "value")},
			perform: func(ctx context.Context, db *sql.DB) {
				db.ExecContext(ctx, "UPDATE users SET name = 'joe'")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertExecContext(t, "UPDATE users SET name = 'joe' /* key='value' */")
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
			perform: func(ctx context.Context, db *sql.DB) {
				db.ExecContext(ctx, "UPDATE users SET name = 'joe'")
			},
			assert: func(t *testing.T, conn *mockConn) {
				conn.assertExecContext(t, "UPDATE users SET name = 'joe' /* user-key='my-key' */")
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
				if err != nil {
					t.Fatal(err)
				}
				defer db.Close()

				cs.perform(ctx, db)
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
	execContext  string
	queryContext string
}

func (m *mockConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	m.queryContext = query
	return nil, nil
}

func (m *mockConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	m.execContext = query
	return nil, nil
}

func (m *mockConn) Close() error {
	return nil
}

func (m *mockConn) assertQueryContext(t *testing.T, query string) {
	t.Helper()

	if m.queryContext != query {
		t.Errorf("got '%v', want '%v'", m.queryContext, query)
	}
}

func (m *mockConn) assertExecContext(t *testing.T, query string) {
	t.Helper()

	if m.execContext != query {
		t.Errorf("got '%v', want '%v'", m.execContext, query)
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
