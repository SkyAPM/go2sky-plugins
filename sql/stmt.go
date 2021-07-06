package sql

import (
	"context"
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

// stmt is a tracing wrapper for driver.Stmt
type stmt struct {
	stmt   driver.Stmt
	tracer *go2sky.Tracer

	// addr defines the address of sql server, format in host:port
	addr string
}

func (s stmt) Close() error {
	panic("implement me")
}

func (s stmt) NumInput() int {
	panic("implement me")
}

func (s stmt) Exec(args []driver.Value) (driver.Result, error) {
	panic("implement me")
}

func (s stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	panic("implement me")
}

func (s stmt) Query(args []driver.Value) (driver.Rows, error) {
	panic("implement me")
}

func (s stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	panic("implement me")
}
