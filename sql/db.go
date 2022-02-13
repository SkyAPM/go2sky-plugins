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

package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/SkyAPM/go2sky"
)

// DB wrap sql.DB and support trace
type DB struct {
	*sql.DB

	tracer *go2sky.Tracer
	opts   *options
}

// OpenDB support trace
func OpenDB(c driver.Connector, tracer *go2sky.Tracer, opts ...Option) *DB {
	db := sql.OpenDB(c)

	options := &options{
		dbType:      UNKNOWN,
		componentID: componentIDUnknown,
		reportQuery: false,
		reportParam: false,
	}
	for _, o := range opts {
		o(options)
	}

	return &DB{
		DB:     db,
		tracer: tracer,
		opts:   options,
	}
}

// Open support trace
func Open(driverName, dataSourceName string, tracer *go2sky.Tracer, opts ...Option) (*DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}

	options := &options{
		dbType:      UNKNOWN,
		componentID: componentIDUnknown,
		reportQuery: false,
		reportParam: false,
	}
	for _, o := range opts {
		o(options)
	}

	if options.peer == "" {
		options.peer = parseDsn(options.dbType, dataSourceName)
	}

	return &DB{
		DB:     db,
		tracer: tracer,
		opts:   options,
	}, nil
}

// PingContext support trace
func (db *DB) PingContext(ctx context.Context) error {
	span, err := createSpan(ctx, db.tracer, db.opts, "ping")
	if err != nil {
		return err
	}
	defer span.End()
	err = db.DB.PingContext(ctx)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

// PrepareContext support trace
func (db *DB) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &Stmt{
		Stmt:  stmt,
		db:    db,
		query: query,
	}, nil
}

// ExecContext support trace
func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	span, err := createSpan(ctx, db.tracer, db.opts, "execute")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, query)
	}
	if db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	res, err := db.DB.ExecContext(ctx, query, args...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return res, err
}

// QueryContext support trace
func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	span, err := createSpan(ctx, db.tracer, db.opts, "query")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, query)
	}
	if db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	rows, err := db.DB.QueryContext(ctx, query, args...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return rows, err
}

// QueryRowContext support trace
func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	span, err := createSpan(ctx, db.tracer, db.opts, "query")
	if err != nil {
		return nil
	}
	defer span.End()

	if db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, query)
	}
	if db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	return db.DB.QueryRowContext(ctx, query, args...)
}

// BeginTx support trace
func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	span, err := createSpan(ctx, db.tracer, db.opts, "begin")
	if err != nil {
		return nil, err
	}
	defer span.End()

	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		span.Error(time.Now(), err.Error())
		return nil, err
	}

	return &Tx{
		Tx:  tx,
		db:  db,
		ctx: ctx,
	}, nil
}

// Conn support trace
func (db *DB) Conn(ctx context.Context) (*Conn, error) {
	conn, err := db.DB.Conn(ctx)
	if err != nil {
		return nil, err
	}
	return &Conn{
		Conn: conn,
		db:   db,
	}, nil
}
