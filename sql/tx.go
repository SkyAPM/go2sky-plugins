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

package sql

import (
	"context"
	"database/sql"
	"time"

	"github.com/SkyAPM/go2sky"
)

type Tx struct {
	*sql.Tx

	db   *DB
	span go2sky.Span
}

func (tx *Tx) Commit() error {
	if tx.span != nil {
		tx.span.Tag(tagDbStatement, "commit")
		defer tx.span.End()
	}
	return tx.Tx.Commit()
}

func (tx *Tx) Rollback() error {
	if tx.span != nil {
		tx.span.Tag(tagDbStatement, "rollback")
		defer tx.span.End()
	}
	return tx.Tx.Rollback()
}

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

func (tx *Tx) StmtContext(ctx context.Context, stmt *Stmt) *Stmt {
	st := tx.Tx.StmtContext(ctx, stmt.Stmt)
	return &Stmt{
		Stmt:  st,
		db:    tx.db,
		query: stmt.query,
	}
}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	span, err := createSpan(ctx, tx.db.tracer, tx.db.opts, "execute")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if tx.db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if tx.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	res, err := tx.Tx.ExecContext(ctx, query, args...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return res, err
}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	span, err := createSpan(ctx, tx.db.tracer, tx.db.opts, "query")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if tx.db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if tx.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	rows, err := tx.Tx.QueryContext(ctx, query, args...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return rows, err
}

func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	span, err := createSpan(ctx, tx.db.tracer, tx.db.opts, "query")
	if err != nil {
		return nil
	}
	defer span.End()

	if tx.db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if tx.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	return tx.Tx.QueryRowContext(ctx, query, args...)
}
