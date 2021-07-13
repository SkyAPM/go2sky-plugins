// Licensed to SkyAPM org under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. SkyAPM org licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

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

// swSQLDriver is a tracing wrapper for driver.Driver
type swSQLDriver struct {
	driver driver.Driver
	tracer *go2sky.Tracer

	dbType DBType
}

func NewTracerDriver(driver driver.Driver, tracer *go2sky.Tracer, dbType DBType) driver.Driver {
	return &swSQLDriver{
		driver: driver,
		tracer: tracer,
		dbType: dbType,
	}
}

func (d *swSQLDriver) Open(name string) (driver.Conn, error) {
	attr := newAttribute(name, d.dbType)
	span, err := createSpan(context.Background(), d.tracer, attr, "open")
	if err != nil {
		return nil, err
	}
	defer span.End()
	span.Tag(tagDbType, string(attr.dbType))
	span.Tag(tagDbInstance, attr.peer)

	c, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &conn{
		conn:   c,
		tracer: d.tracer,
		attr:   attr,
	}, nil
}
