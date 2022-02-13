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
	"time"

	"github.com/SkyAPM/go2sky"
)

// Tx wrap sql.Tx and support trace
type Tx struct {
	*sql.Tx

	db  *DB
	ctx context.Context
}

// Commit support trace
func (tx *Tx) Commit() (err error) {
	span, err := createSpan(tx.ctx, tx.db.tracer, tx.db.opts, "commit")
	if err != nil {
		return err
	}
	defer span.End()

	err = tx.Tx.Commit()
	if err != nil {
		span.Error(time.Now(), err.Error())
		return err
	}

	return nil
}

// Rollback support trace
func (tx *Tx) Rollback() (err error) {
	span, err := createSpan(tx.ctx, tx.db.tracer, tx.db.opts, "rollback")
	if err != nil {
		return err
	}
	defer span.End()

	err = tx.Tx.Rollback()
	if err != nil {
		span.Error(time.Now(), err.Error())
		return err
	}

	return nil
}

// Prepare support trace
func (tx *Tx) Prepare(query string) (*Stmt, error) {
	stmt, err := tx.Tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &Stmt{
		Stmt:  stmt,
		db:    tx.db,
		query: query,
	}, nil
}

// PrepareContext support trace
func (tx *Tx) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := tx.Tx.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &Stmt{
		Stmt:  stmt,
		db:    tx.db,
		query: query,
	}, nil
}

// StmtContext support trace
func (tx *Tx) StmtContext(ctx context.Context, stmt *Stmt) *Stmt {
	st := tx.Tx.StmtContext(ctx, stmt.Stmt)
	return &Stmt{
		Stmt:  st,
		db:    tx.db,
		query: stmt.query,
	}
}

// Exec support trace
func (tx *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return tx.ExecContext(tx.ctx, query, args...)
}

// ExecContext support trace
func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	if id := go2sky.SpanID(ctx); id == go2sky.EmptySpanID {
		// if ctx do not contain parent span, use transaction ctx instead
		ctx = tx.ctx
	}
	span, err := createSpan(ctx, tx.db.tracer, tx.db.opts, "execute")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if tx.db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, query)
	}
	if tx.db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	res, err := tx.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return res, err
}

// Query support trace
func (tx *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return tx.QueryContext(tx.ctx, query, args)
}

// QueryContext support trace
func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	if id := go2sky.SpanID(ctx); id == go2sky.EmptySpanID {
		// if ctx do not contain parent span, use transaction ctx instead
		ctx = tx.ctx
	}
	span, err := createSpan(ctx, tx.db.tracer, tx.db.opts, "query")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if tx.db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, query)
	}
	if tx.db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	rows, err := tx.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return rows, err
}

// QueryRow support trace
func (tx *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return tx.QueryRowContext(tx.ctx, query, args)
}

// QueryRowContext support trace
func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	if id := go2sky.SpanID(ctx); id == go2sky.EmptySpanID {
		// if ctx do not contain parent span, use transaction ctx instead
		ctx = tx.ctx
	}
	span, err := createSpan(ctx, tx.db.tracer, tx.db.opts, "query")
	if err != nil {
		return nil
	}
	defer span.End()

	if tx.db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, query)
	}
	if tx.db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	return tx.Tx.QueryRowContext(ctx, query, args...)
}
