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

	sql.Register("skywalking-sql", NewTracerDriver(mysqlDriver, tracer, WithSqlDBType(MYSQL), WithQueryReport()))
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
