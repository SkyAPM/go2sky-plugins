package sql

import (
	"context"
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

type connector struct {
	connector driver.Connector
	tracer    *go2sky.Tracer

	// dsn defines the destination of sql server, format in host:port
	dsn string
}

func (c connector) Connect(ctx context.Context) (driver.Conn, error) {
	panic("implement me")
}

func (c connector) Driver() driver.Driver {
	panic("implement me")
}
