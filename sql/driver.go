package sql

import (
	"context"
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

type DBType string

const (
	MYSQL DBType = "mysql"
	IPV4  DBType = "others"
)

type Driver struct {
	driver driver.Driver
	tracer *go2sky.Tracer

	dbType DBType
}

func NewTracerDriver(driver driver.Driver, tracer *go2sky.Tracer, dbType DBType) driver.Driver {
	return &Driver{
		driver: driver,
		tracer: tracer,
		dbType: dbType,
	}
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	addr := parseAddr(name, d.dbType)
	s, err := d.tracer.CreateExitSpan(context.Background(), genOpName(d.dbType, "open"), addr, emptyInjectFunc)
	if err != nil {
		return nil, err
	}
	defer s.End()
	c, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &conn{
		conn:   c,
		tracer: d.tracer,
		addr:   addr,
	}, nil
}
