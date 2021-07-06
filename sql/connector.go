package sql

import (
	"context"
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

type connector struct {
	connector driver.Connector
	tracer    *go2sky.Tracer

	// addr defines the address of sql server, format in host:port
	addr string
}

func (c *connector) Connect(ctx context.Context) (driver.Conn, error) {
	panic("implement me")
}

func (c *connector) Driver() driver.Driver {
	return c.connector.Driver()
}

type fallbackConnector struct {
	driver driver.Driver
	name   string
}

func (fc *fallbackConnector) Connect(ctx context.Context) (driver.Conn, error) {
	conn, err := fc.driver.Open(fc.name)
	if err != nil {
		return nil, err
	}
	if ctx.Err() != nil { // ctx done closed
		conn.Close()
		return nil, ctx.Err()
	}
	return conn, nil
}

func (fc *fallbackConnector) Driver() driver.Driver {
	return fc.driver
}

// OpenConnector implements driver.DriverContext OpenConnector
func (d *Driver) OpenConnector(name string) (driver.Connector, error) {
	if dc, ok := d.driver.(driver.DriverContext); ok {
		c, err := dc.OpenConnector(name)
		if err != nil {
			return nil, err
		}
		return &connector{
			connector: c,
			tracer:    d.tracer,
			addr:      parseAddr(name, d.dbtype),
		}, nil
	}

	// given driver does not implement driver.DriverContext interface
	return &connector{
		connector: nil,
		tracer:    d.tracer,
		addr:      parseAddr(name, d.dbtype),
	}, nil
}
