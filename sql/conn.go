package sql

import (
	"context"
	"database/sql/driver"
	"fmt"

	"github.com/SkyAPM/go2sky"
)

// conn is a tracing wrapper for driver.Conn
type conn struct {
	conn   driver.Conn
	tracer *go2sky.Tracer

	// addr defines the address of sql server, format in host:port
	addr string
}

func (c *conn) Ping(ctx context.Context) error {
	if pinger, ok := c.conn.(driver.Pinger); ok {
		s, err := c.tracer.CreateExitSpan(ctx, "go2sky/sql/ping", c.addr, emptyInjectFunc)
		if err != nil {
			return err
		}
		defer s.End()
		return pinger.Ping(ctx)
	}
	return fmt.Errorf("driver not implements driver.Pinger interface")
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	return c.conn.Prepare(query)
}

// PrepareContext implements driver.ConnPrepareContext PrepareContext
func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	panic("")
}

// Close implements driver.Conn Close
func (c *conn) Close() error {
	return c.conn.Close()
}

// Begin implements driver.Conn Begin
func (c *conn) Begin() (driver.Tx, error) {
	panic("implement me")
}

// BeginTx implements driver.ConnBeginTx BeginTx
func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	panic("implement me")
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	panic("implement me")
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	panic("implement me")
}

func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	panic("implement me")
}

func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	panic("implement me")
}
