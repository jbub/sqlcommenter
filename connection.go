package sqlcommenter

import (
	"context"
	"database/sql/driver"
)

var (
	_ driver.Pinger             = (*connection)(nil)
	_ driver.Execer             = (*connection)(nil) // nolint:staticcheck
	_ driver.ExecerContext      = (*connection)(nil)
	_ driver.Queryer            = (*connection)(nil) // nolint:staticcheck
	_ driver.QueryerContext     = (*connection)(nil)
	_ driver.Conn               = (*connection)(nil)
	_ driver.ConnPrepareContext = (*connection)(nil)
	_ driver.ConnBeginTx        = (*connection)(nil)
	_ driver.SessionResetter    = (*connection)(nil)
	_ driver.NamedValueChecker  = (*connection)(nil)
)

func newConn(conn driver.Conn, cmt *commenter) *connection {
	return &connection{
		Conn: conn,
		cmt:  cmt,
	}
}

type connection struct {
	driver.Conn
	cmt *commenter
}

func (c *connection) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	preparer, ok := c.Conn.(driver.ConnPrepareContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return preparer.PrepareContext(ctx, query)
}

func (c *connection) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	beginTx, ok := c.Conn.(driver.ConnBeginTx)
	if !ok {
		return nil, driver.ErrSkip
	}
	return beginTx.BeginTx(ctx, opts)
}

func (c *connection) Query(query string, args []driver.Value) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.Queryer) // nolint:staticcheck
	if !ok {
		return nil, driver.ErrSkip
	}
	return queryer.Query(c.withComment(context.Background(), query), args)
}

func (c *connection) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	queryer, ok := c.Conn.(driver.QueryerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return queryer.QueryContext(ctx, c.withComment(ctx, query), args)
}

func (c *connection) Exec(query string, args []driver.Value) (driver.Result, error) {
	execer, ok := c.Conn.(driver.Execer) // nolint:staticcheck
	if !ok {
		return nil, driver.ErrSkip
	}
	return execer.Exec(c.withComment(context.Background(), query), args)
}

func (c *connection) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	execer, ok := c.Conn.(driver.ExecerContext)
	if !ok {
		return nil, driver.ErrSkip
	}
	return execer.ExecContext(ctx, c.withComment(ctx, query), args)
}

func (c *connection) Ping(ctx context.Context) error {
	pinger, ok := c.Conn.(driver.Pinger)
	if !ok {
		return driver.ErrSkip
	}
	return pinger.Ping(ctx)
}

func (c *connection) CheckNamedValue(value *driver.NamedValue) error {
	checker, ok := c.Conn.(driver.NamedValueChecker)
	if !ok {
		return driver.ErrSkip
	}
	return checker.CheckNamedValue(value)
}

func (c *connection) ResetSession(ctx context.Context) error {
	resetter, ok := c.Conn.(driver.SessionResetter)
	if !ok {
		return driver.ErrSkip
	}
	return resetter.ResetSession(ctx)
}

func (c *connection) withComment(ctx context.Context, query string) string {
	return c.cmt.comment(ctx, query)
}
