package sqlcommenter

import (
	"context"
	"database/sql/driver"
)

// WrapDriver wraps sql driver with sqlcommenter support.
func WrapDriver(drv driver.Driver, opts ...Option) driver.Driver {
	return &commentDriver{
		drv: drv,
		cmt: newCommenter(opts...),
	}
}

type commentDriver struct {
	drv driver.Driver
	cmt *commenter
}

func (d *commentDriver) Open(name string) (driver.Conn, error) {
	conn, err := d.drv.Open(name)
	if err != nil {
		return nil, err
	}
	return newConn(conn, d.cmt), nil
}

func (d *commentDriver) OpenConnector(name string) (driver.Connector, error) {
	drvCtx, ok := d.drv.(driver.DriverContext)
	if !ok {
		return newDSNConnector(name, d), nil
	}

	ctr, err := drvCtx.OpenConnector(name)
	if err != nil {
		return nil, err
	}
	return newConnector(ctr, d), nil
}

func newConnector(ctr driver.Connector, drv *commentDriver) *connector {
	return &connector{
		ctr: ctr,
		drv: drv,
	}
}

type connector struct {
	ctr driver.Connector
	drv *commentDriver
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := c.ctr.Connect(ctx)
	if err != nil {
		return nil, err
	}
	return newConn(conn, c.drv.cmt), nil
}

func (c *connector) Driver() driver.Driver {
	return c.drv
}

func newDSNConnector(dsn string, drv *commentDriver) *dsnConnector {
	return &dsnConnector{
		dsn: dsn,
		drv: drv,
	}
}

type dsnConnector struct {
	dsn string
	drv *commentDriver
}

func (c *dsnConnector) Connect(context.Context) (driver.Conn, error) {
	return c.drv.Open(c.dsn)
}

func (c *dsnConnector) Driver() driver.Driver {
	return c.drv
}
