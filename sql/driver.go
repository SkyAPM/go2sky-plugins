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

// go2SkySQLDriver is a tracing wrapper for driver.Driver
type go2SkySQLDriver struct {
	driver driver.Driver
	tracer *go2sky.Tracer

	opts *options
}

func NewTracerDriver(driver driver.Driver, tracer *go2sky.Tracer, opts ...Option) driver.Driver {
	options := &options{
		dbType: IPV4,
	}
	for _, o := range opts {
		o(options)
	}
	if options.componentID == 0 {
		options.setComponentID()
	}
	return &go2SkySQLDriver{
		driver: driver,
		tracer: tracer,
		opts:   options,
	}
}

func (d *go2SkySQLDriver) Open(name string) (driver.Conn, error) {
	if d.opts.peer == "" {
		d.opts.setPeerWithDsn(name)
	}
	span, err := createSpan(context.Background(), d.tracer, d.opts, "open")
	if err != nil {
		return nil, err
	}
	defer span.End()
	span.Tag(tagDbType, string(d.opts.dbType))
	span.Tag(tagDbInstance, d.opts.peer)

	c, err := d.driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &conn{
		conn:   c,
		tracer: d.tracer,
		opts:   d.opts,
	}, nil
}
