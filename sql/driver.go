package sql

import (
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

type Driver struct {
	driver driver.Driver
	tracer *go2sky.Tracer
}

func NewTracerDriver(driver driver.Driver, tracer *go2sky.Tracer) driver.Driver {
	return &Driver{
		driver: driver,
		tracer: tracer,
	}
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	// TODO: parse destination
	// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
	c, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &conn{conn: c, tracer: d.tracer}, nil
}

// OpenConnector implements driver.DriverContext OpenConnector
func (d *Driver) OpenConnector(name string) (driver.Connector, error) {
	panic("implement me")
}

// emptyInjectFunc defines a empty injector for propagation.Injector function
func emptyInjectFunc(key, value string) error { return nil }
