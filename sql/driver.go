package sql

import (
	"database/sql/driver"
	"regexp"

	"github.com/SkyAPM/go2sky"
)

type DBType int

const (
	MYSQL DBType = iota
	IPV4
)

type Driver struct {
	driver driver.Driver
	tracer *go2sky.Tracer

	dbtype DBType
}

func NewTracerDriver(driver driver.Driver, tracer *go2sky.Tracer, dbType DBType) driver.Driver {
	return &Driver{
		driver: driver,
		tracer: tracer,
		dbtype: dbType,
	}
}

func (d *Driver) Open(name string) (driver.Conn, error) {
	c, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &conn{
		conn:   c,
		tracer: d.tracer,
		addr:   parseAddr(name, d.dbtype),
	}, nil
}

// emptyInjectFunc defines a empty injector for propagation.Injector function
func emptyInjectFunc(key, value string) error { return nil }

// parseAddr parse dsn to a endpoint addr string (host:port)
func parseAddr(dsn string, dbType DBType) (addr string) {
	switch dbType {
	case MYSQL:
		// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
		re := regexp.MustCompile(`\(.+\)`)
		addr = re.FindString(dsn)
		addr = addr[1 : len(addr)-1]
	case IPV4:
		// ipv4 addr
		re := regexp.MustCompile(`((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}:\d{1,5}`)
		addr = re.FindString(dsn)
	}
	return
}
