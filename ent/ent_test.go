//nolint
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

package ent_test

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SkyAPM/go2sky"
	sqlPlugin "github.com/SkyAPM/go2sky-plugins/sql"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/pkg/errors"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"

	"github.com/SkyAPM/go2sky-plugins/ent/gen/entschema"
	entuser "github.com/SkyAPM/go2sky-plugins/ent/gen/entschema/user"
)

func TestEntTrace(t *testing.T) {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800", reporter.WithCheckInterval(time.Second*10))
	if err != nil {
		panic(err)
	}
	tracer, err := go2sky.NewTracer("example", go2sky.WithReporter(r))
	if err != nil {
		panic(err)
	}

	entClient01 := NewEntClient01(tracer)

	go func() {
		doWork(tracer, entClient01)
	}()

	entClient02 := NewEntClient02(tracer)

	go func() {
		doWork(tracer, entClient02)
	}()

	select {}
}

func doWork(tracer *go2sky.Tracer, entClient *entschema.Client) {
	span, ctx, err := tracer.CreateLocalSpan(context.Background(), go2sky.WithOperationName("test_ent"))
	if err != nil {
		return
	}
	defer span.End()
	span.SetComponent(5012)
	span.SetSpanLayer(v3.SpanLayer_Database)

	user, err := entClient.User.Query().Where(entuser.ID(1)).First(ctx)
	if err != nil {
		return
	}
	fmt.Printf("%+v \n", user)

	if err = entClient.User.Update().
		SetName("test01").
		Where(entuser.ID(1)).
		Exec(ctx); err != nil {
		return
	}

	user, err = entClient.User.Query().Where(entuser.ID(1)).First(ctx)
	if err != nil {
		return
	}
	fmt.Printf("%+v \n", user)

	if err = WithTx(ctx, entClient, func(tx *entschema.Tx) error {
		if err = tx.User.Create().
			SetName("test02").
			SetAge(12).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return
	}

	if _, err = entClient.User.Delete().
		Where(entuser.Name("test02")).
		Exec(ctx); err != nil {
		return
	}
}

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

// NewEntClient01 ent option. (Advice)
func NewEntClient01(tracer *go2sky.Tracer) *entschema.Client {
	driver, url := "mysql", "root:12345678@(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true&loc=UTC&interpolateParams=true"

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

// NewEntClient02 ent option.
func NewEntClient02(tracer *go2sky.Tracer) *entschema.Client {
	driver, url := "mysql", "root:12345678@(127.0.0.1:3306)/test?charset=utf8mb4&parseTime=true&loc=UTC&interpolateParams=true"

	db, err := sql.Open(driver, url)
	if err != nil {
		panic(err)
	}

	drv := entsql.OpenDB(driver, db)

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

// WithTx 事务快捷方式.
// https://entgo.io/docs/transactions/#best-practices
func WithTx(ctx context.Context, client *entschema.Client, fn func(tx *entschema.Tx) error) error {
	// not use. throw error panic: interface conversion: sql.ExecQuerier is *sql.DB, not *sql.DB
	// tx, err := client.Tx(ctx, nil)
	tx, err := client.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if v := recover(); v != nil {
			err = tx.Rollback()
			panic(v)
		}
	}()
	if err = fn(tx); err != nil {
		if r := tx.Rollback(); r != nil {
			err = errors.Wrapf(err, "rolling back transaction: %v", r)
		}
		return err
	}
	if err = tx.Commit(); err != nil {
		return errors.Wrapf(err, "committing transaction: %v", err)
	}
	return nil
}
