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

type connector struct {
	connector driver.Connector
	tracer    *go2sky.Tracer

	// attr include some attributes need to report to OAP server
	attr attribute
}

func (ct *connector) Connect(ctx context.Context) (driver.Conn, error) {
	span, err := createSpan(ctx, ct.tracer, ct.attr, "connect")
	if err != nil {
		return nil, err
	}
	defer span.End()
	span.Tag(tagDbType, string(ct.attr.dbType))
	span.Tag(tagDbInstance, ct.attr.peer)

	c, err := ct.connector.Connect(ctx)
	return &conn{
		conn:   c,
		tracer: ct.tracer,
		attr:   ct.attr,
	}, nil
}

func (ct *connector) Driver() driver.Driver {
	return ct.connector.Driver()
}

type fallbackConnector struct {
	driver driver.Driver
	name   string
	attr   attribute
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
func (d *swSQLDriver) OpenConnector(name string) (driver.Connector, error) {
	attr := newAttribute(name, d.dbType)
	if dc, ok := d.driver.(driver.DriverContext); ok {
		c, err := dc.OpenConnector(name)
		if err != nil {
			return nil, err
		}
		return &connector{
			connector: c,
			tracer:    d.tracer,
			attr:      attr,
		}, nil
	}

	// given driver does not implement driver.DriverContext interface
	return &connector{
		connector: &fallbackConnector{
			driver: d.driver,
			name:   name,
			attr:   attr,
		},
		tracer: d.tracer,
		attr:   attr,
	}, nil
}
