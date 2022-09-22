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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SkyAPM/go2sky"
	sqlPlugin "github.com/SkyAPM/go2sky-plugins/sql"
	httpplugin "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"

	"github.com/SkyAPM/go2sky-plugins/ent/gen/entschema"
	entuser "github.com/SkyAPM/go2sky-plugins/ent/gen/entschema/user"
)

type testFunc func(context.Context, *entschema.Client) error

const (
	oap     = "mockoap:19876"
	service = "ent-client"
	dsn     = "user:password@tcp(mysql:3306)/database"
	addr    = ":8080"
)

func main() {
	// init tracer
	re, err := reporter.NewGRPCReporter(oap)
	//re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("create grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(service, go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	client := NewEntClient(dsn, tracer)

	if err != nil {
		log.Fatalf("open db error: %v \n", err)
	}

	route := http.NewServeMux()
	route.HandleFunc("/execute", func(res http.ResponseWriter, req *http.Request) {
		tests := []struct {
			name string
			fn   testFunc
		}{
			{"raw", testRaw},
			{"create", testCreate},
			{"query", testQuery},
			{"update", testUpdate},
			{"delete", testDelete},
			{"tx", testTx},
		}

		for _, test := range tests {
			log.Printf("excute test case %s", test.name)
			if err1 := test.fn(req.Context(), client); err1 != nil {
				log.Fatalf("test case %s failed: %v", test.name, err1)
			}
		}
		_, _ = res.Write([]byte("execute sql success"))
	})

	sm, err := httpplugin.NewServerMiddleware(tracer)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}

	log.Println("start client")
	err = http.ListenAndServe(addr, sm(route))
	if err != nil {
		log.Fatalf("client start error: %v \n", err)
	}
}

// NewEntClient new Ent Client.
func NewEntClient(dsn string, tracer *go2sky.Tracer) *entschema.Client {
	driver, url := "mysql", dsn

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

func testRaw(ctx context.Context, client *entschema.Client) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	newClient := entschema.NewClient(entschema.Driver(entsql.OpenDB("mysql", db)))
	if err = newClient.Schema.Create(ctx); err != nil {
		return err
	}
	return nil
}

func testQuery(ctx context.Context, client *entschema.Client) error {
	user, err := client.User.Query().Where(entuser.ID(1)).First(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("%+v \n", user)

	return nil
}

func testCreate(ctx context.Context, client *entschema.Client) error {
	if err := client.User.Create().
		SetName("test02").
		SetAge(12).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func testDelete(ctx context.Context, client *entschema.Client) error {
	if _, err := client.User.Delete().
		Where(entuser.Name("test02")).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func testUpdate(ctx context.Context, client *entschema.Client) error {
	if err := client.User.Update().
		SetName("test01").
		Where(entuser.ID(1)).
		Exec(ctx); err != nil {
		return err
	}
	return nil
}

func testTx(ctx context.Context, client *entschema.Client) error {
	if err := WithTx(ctx, client, func(tx *entschema.Tx) error {
		if err := tx.User.Create().
			SetName("test02").
			SetAge(12).
			Exec(ctx); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}
	return nil
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

// WithTx 事务快捷方式.
// https://entgo.io/docs/transactions/#best-practices
func WithTx(ctx context.Context, client *entschema.Client, fn func(tx *entschema.Tx) error) error {
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
