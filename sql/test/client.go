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

package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"

	sqlPlugin "sql"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/go-sql-driver/mysql"
)

type testFunc func(context.Context, *sql.DB) error

const (
	oap     = "mockoap:19876"
	service = "sql-client"
	dsn     = "user:password@tcp(mysql:3306)/database"
	addr    = ":8080"
)

func main() {
	// init driver
	mysqlDriver := &mysql.MySQLDriver{}

	// init tracer
	re, err := reporter.NewGRPCReporter(oap)
	if err != nil {
		log.Fatalf("create grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(service, go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	sql.Register("skywalking-sql", sqlPlugin.NewTracerDriver(mysqlDriver, tracer,
		sqlPlugin.WithSqlDBType(sqlPlugin.MYSQL),
		sqlPlugin.WithQueryReport(),
		sqlPlugin.WithParamReport()))
	db, err := sql.Open("skywalking-sql", dsn)
	if err != nil {
		log.Fatalf("open db error: %v \n", err)
	}

	route := http.NewServeMux()
	route.HandleFunc("/healthCheck", func(res http.ResponseWriter, req *http.Request) {
		_, _ = res.Write([]byte("Success"))
	})
	route.HandleFunc("/execute", func(res http.ResponseWriter, req *http.Request) {
		tests := []struct {
			name string
			fn   testFunc
		}{
			{"exec", TestExec},
			{"stmt", TestStmt},
			{"commitTx", TestCommitTx},
			{"rollbackTx", TestRollbackTx},
		}

		span, ctx, err := tracer.CreateLocalSpan(context.Background())
		if err != nil {
			log.Fatalf("create span error: %v \n", err)
		}
		defer span.End()

		for _, test := range tests {
			if err := test.fn(ctx, db); err != nil {
				log.Fatalf("test case %s failed: %v", test.name, err)
			}
		}
		_, _ = res.Write([]byte("Execute sql success"))
	})

	log.Println("start client")
	err = http.ListenAndServe(addr, route)
	if err != nil {
		log.Fatalf("client start error: %v \n", err)
	}
}

func TestExec(ctx context.Context, db *sql.DB) error {
	if err := db.PingContext(ctx); err != nil {
		return err
	}

	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS users`); err != nil {
		return fmt.Errorf("exec drop error: %w", err)
	}

	if _, err := db.ExecContext(ctx, `CREATE TABLE IF NOT EXISTS users (id char(255), name VARCHAR(255), age INTEGER)`); err != nil {
		return fmt.Errorf("exec create error: %w", err)
	}

	// test insert
	if _, err := db.ExecContext(ctx, `INSERT INTO users (id, name, age) VALUE ( ?, ?, ?)`, "0", "foo", 10); err != nil {
		return fmt.Errorf("exec insert error: %w", err)
	}

	var name string
	// test select
	if err := db.QueryRowContext(ctx, `SELECT name FROM users WHERE id = ?`, "0").Scan(&name); err != nil {
		return fmt.Errorf("query select error: %w", err)
	}

	return nil
}

func TestStmt(ctx context.Context, db *sql.DB) error {
	stmt, err := db.PrepareContext(ctx, `INSERT INTO users (id, name, age) VALUE ( ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.ExecContext(ctx, "1", "bar", 11)
	if err != nil {
		return err
	}

	return nil
}

func TestCommitTx(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx error: %v \n", err)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE users SET name = ? WHERE id = ?`, "foobar", "0"); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}
	return nil
}

func TestRollbackTx(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx error: %v \n", err)
	}

	if _, err := tx.ExecContext(ctx, `UPDATE users SET name = ? WHERE id = ?`, "foobar", "1"); err != nil {
		return err
	}

	if err := tx.Rollback(); err != nil {
		return err
	}
	return nil
}
