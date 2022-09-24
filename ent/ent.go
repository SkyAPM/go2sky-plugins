//
// Copyright 2022 SkyAPM org
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

package ent

import (
	"context"
	"database/sql"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky-plugins/ent/gen/entschema"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	sqlPlugin "github.com/SkyAPM/go2sky-plugins/sql"
)

// Driver custom wrapper ent driver, Don't use native.
type Driver struct {
	dialect.Driver
	db *sqlPlugin.DB
}

// NewDriver EntDriver.
func NewDriver(drv dialect.Driver, db *sqlPlugin.DB) *Driver {
	return &Driver{drv, db}
}

// BeginTx custom.
func (e *Driver) BeginTx(ctx context.Context, option *sql.TxOptions) (dialect.Tx, error) {
	tx, err := e.db.BeginTx(ctx, option)
	if err != nil {
		return nil, err
	}
	return &entsql.Tx{
		Conn: entsql.Conn{ExecQuerier: tx},
		Tx:   tx,
	}, nil
}

// NewEntClient ent client.
func NewEntClient(dsn string, tracer *go2sky.Tracer) *entschema.Client {
	driver, url := "mysql", dsn
	/*
		db, err := sql.Open(driver, url)
		if err != nil {
			panic(err)
		}
		drv := entsql.OpenDB(driver, db)
	*/
	drv, err := entsql.Open(driver, url)
	if err != nil {
		panic(err)
	}

	apmDB, err := sqlPlugin.Open(
		driver,
		url,
		tracer,
		sqlPlugin.WithSQLDBType(sqlPlugin.MYSQL),
		sqlPlugin.WithQueryReport(),
		sqlPlugin.WithParamReport(),
	)
	if err != nil {
		panic(err)
	}
	apmDB.DB.SetMaxIdleConns(50)
	apmDB.DB.SetMaxOpenConns(200)

	drv.Conn = entsql.Conn{ExecQuerier: apmDB}

	EntClient := entschema.NewClient(
		entschema.Driver(
			NewDriver(drv, apmDB),
		),
	)

	return EntClient
}
