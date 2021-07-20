package main

import (
	"context"
	"database/sql"
	"log"

	swsql "sql"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/go-sql-driver/mysql"
)

const (
	oap     = "mockoap:19876"
	service = "sql-client"
	dsn     = "user:password@tcp(127.0.0.1:3306)/database"
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

	sql.Register("skywalking-sql", swsql.NewTracerDriver(mysqlDriver, tracer, swsql.WithSqlDBType(swsql.MYSQL), swsql.WithQueryReport()))
	db, err := sql.Open("skywalking-sql", dsn)
	if err != nil {
		log.Fatalf("open db error: %v \n", err)
	}

	span, ctx, err := tracer.CreateLocalSpan(context.Background())
	if err != nil {
		log.Fatalf("create span error: %v \n", err)
	}
	defer span.End()

	_, err = db.ExecContext(ctx, `CREATE TABLE test (id char(255), name VARCHAR(255), age INTEGER)`)

	// test insert
	_, err = db.ExecContext(ctx,
		`INSERT INTO test (id, name, age, datetime) VALUE ( ?, ?, ?)`,
		"0", "foo", 10)

	var name string
	// test select
	err = db.QueryRowContext(ctx,
		`SELECT name FROM test WHERE id = ?`,
		"1").Scan(&name)
}
