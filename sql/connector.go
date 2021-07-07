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

	// addr defines the address of sql server, format in host:port
	addr string
	// dbType defines the sql server type
	dbType DBType
}

func (ct *connector) Connect(ctx context.Context) (driver.Conn, error) {
	span, err := ct.tracer.CreateExitSpan(ctx, genOpName(ct.dbType, "connect"), ct.addr, emptyInjectFunc)
	if err != nil {
		return nil, err
	}
	defer span.End()

	c, err := ct.connector.Connect(ctx)
	return &conn{
		conn:   c,
		tracer: ct.tracer,
		addr:   ct.addr,
	}, nil
}

func (ct *connector) Driver() driver.Driver {
	return ct.connector.Driver()
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
func (d *swDriver) OpenConnector(name string) (driver.Connector, error) {
	if dc, ok := d.driver.(driver.DriverContext); ok {
		c, err := dc.OpenConnector(name)
		if err != nil {
			return nil, err
		}
		return &connector{
			connector: c,
			tracer:    d.tracer,
			addr:      parseAddr(name, d.dbType),
		}, nil
	}

	// given driver does not implement driver.DriverContext interface
	return &connector{
		connector: nil,
		tracer:    d.tracer,
		addr:      parseAddr(name, d.dbType),
	}, nil
}
