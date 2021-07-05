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
	return Driver{
		driver: driver,
		tracer: tracer,
	}
}

func (d Driver) Open(name string) (driver.Conn, error) {
	panic("implement me")
}