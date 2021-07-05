package sql

import (
	"context"
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

// conn is a tracing wrapper for driver.Conn
type conn struct {
	conn driver.Conn
	tracer *go2sky.Tracer
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	panic("implement me")
}

func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	panic("implement me")
}

func (c *conn) Close() error {
	panic("implement me")
}

func (c *conn) Begin() (driver.Tx, error) {
	panic("implement me")
}

func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	panic("implement me")
}

