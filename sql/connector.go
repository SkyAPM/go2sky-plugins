//
// Copyright 2021 SkyAPM org
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package sql

import (
	"context"
	"database/sql/driver"
	"time"

	"github.com/SkyAPM/go2sky"
)

type connector struct {
	connector driver.Connector
	tracer    *go2sky.Tracer

	opts *options
}

func (ct *connector) Connect(ctx context.Context) (driver.Conn, error) {
	span, err := createSpan(ctx, ct.tracer, ct.opts, "connect")
	if err != nil {
		return nil, err
	}
	defer span.End()

	span.Tag(tagDbType, string(ct.opts.dbType))
	span.Tag(tagDbInstance, ct.opts.peer)

	c, err := ct.connector.Connect(ctx)
	if err != nil {
		span.Error(time.Now(), err.Error())
		return nil, err
	}
	return &conn{
		conn:   c,
		tracer: ct.tracer,
		opts:   ct.opts,
	}, nil
}

func (ct *connector) Driver() driver.Driver {
	return ct.connector.Driver()
}

type fallbackConnector struct {
	driver driver.Driver
	name   string

	opts *options
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
func (d *go2SkySQLDriver) OpenConnector(name string) (driver.Connector, error) {
	if d.opts.peer == "" {
		d.opts.setPeerWithDsn(name)
	}
	if dc, ok := d.driver.(driver.DriverContext); ok {
		c, err := dc.OpenConnector(name)
		if err != nil {
			return nil, err
		}
		return &connector{
			connector: c,
			tracer:    d.tracer,
			opts:      d.opts,
		}, nil
	}

	// given driver does not implement driver.DriverContext interface
	return &connector{
		connector: &fallbackConnector{
			driver: d.driver,
			name:   name,
			opts:   d.opts,
		},
		tracer: d.tracer,
		opts:   d.opts,
	}, nil
}
