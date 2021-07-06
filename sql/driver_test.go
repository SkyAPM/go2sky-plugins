package sql

import (
	"context"
	"database/sql"
	"testing"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/assert"
)

func TestDriver(t *testing.T) {
	// init driver
	mysqlDriver := &mysql.MySQLDriver{}

	// init tracer
	re, err := reporter.NewLogReporter()
	assert.NoError(t, err)
	defer re.Close()

	tracer, err := go2sky.NewTracer("test-driver", go2sky.WithReporter(re))
	assert.NoError(t, err)

	sql.Register("skywalking-sql", NewTracerDriver(mysqlDriver, tracer, MYSQL))
	db, err := sql.Open("skywalking-sql", "user:password@tcp(127.0.0.1:3306)/testdb")
	assert.NoError(t, err)

	span, ctx, err := tracer.CreateLocalSpan(context.Background())
	assert.NoError(t, err)
	defer span.End()

	_, err = db.ExecContext(ctx, `CREATE TABLE test (id char(255), name VARCHAR(255), age INTEGER)`)
	assert.NoError(t, err)

	// test insert
	_, err = db.ExecContext(ctx,
		`INSERT INTO test (id, name, age, datetime) VALUE ( ?, ?, ?)`,
		"0", "foo", 10)
	assert.NoError(t, err)

	var name string
	// test select
	err = db.QueryRowContext(ctx,
		`SELECT * FROM test WHERE id = ?`,
		"1").Scan(&name)
	assert.NoError(t, err)
	assert.Equal(t, "foo", name)

}
