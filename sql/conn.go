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
)

type Conn struct {
	*sql.Conn

	db *DB
}

func (c *Conn) PingContext(ctx context.Context) error {
	span, err := createSpan(ctx, c.db.tracer, c.db.opts, "ping")
	if err != nil {
		return err
	}
	defer span.End()
	err = c.Conn.PingContext(ctx)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

func (c *Conn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	span, err := createSpan(ctx, c.db.tracer, c.db.opts, "exec")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if c.db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if c.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	res, err := c.Conn.ExecContext(ctx, query, args)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return res, err
}

func (c *Conn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	span, err := createSpan(ctx, c.db.tracer, c.db.opts, "query")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if c.db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if c.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	rows, err := c.Conn.QueryContext(ctx, query, args)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return rows, err
}

func (c *Conn) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	span, err := createSpan(ctx, c.db.tracer, c.db.opts, "query")
	if err != nil {
		return nil
	}
	defer span.End()

	if c.db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if c.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	return c.Conn.QueryRowContext(ctx, query, args)
}

func (c *Conn) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := c.Conn.PrepareContext(ctx, query)
	return &Stmt{
		Stmt:  stmt,
		db:    c.db,
		query: query,
	}, err
}

func (c *Conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	span, err := createSpan(ctx, c.db.tracer, c.db.opts, "transaction")
	if err != nil {
		return nil, err
	}

	tx, err := c.Conn.BeginTx(ctx, opts)
	if err != nil {
		span.Error(time.Now(), err.Error())
		span.End()
		return nil, err
	}

	return &Tx{
		Tx:   tx,
		span: span,
	}, nil

}
