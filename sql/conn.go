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
	"database/sql/driver"

	"github.com/SkyAPM/go2sky"
)

// conn is a tracing wrapper for driver.Conn
type conn struct {
	conn   driver.Conn
	tracer *go2sky.Tracer

	// addr defines the address of sql server, format in host:port
	addr string
	// dbType defines the sql server type
	dbType DBType
}

// Ping implements driver.Pinger interface.
// If the underlying Conn does not implement Pinger,
// Ping will return a ErrUnsupportedOp error
func (c *conn) Ping(ctx context.Context) error {
	if pinger, ok := c.conn.(driver.Pinger); ok {
		span, err := c.tracer.CreateExitSpan(ctx, genOpName(c.dbType, "ping"), c.addr, emptyInjectFunc)
		if err != nil {
			return err
		}
		defer span.End()
		return pinger.Ping(ctx)
	}
	return ErrUnsupportedOp
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	st, err := c.conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &stmt{
		stmt:   st,
		tracer: c.tracer,
		query:  query,
		addr:   c.addr,
		dbType: c.dbType,
	}, nil
}

// PrepareContext implements driver.ConnPrepareContext PrepareContext
// If the underlying Conn does not implements
// driver.ConnPrepareContext interface, this method
// will use Prepare instead.
func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if ConnPrepareContext, ok := c.conn.(driver.ConnPrepareContext); ok {
		st, err := ConnPrepareContext.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &stmt{
			stmt:   st,
			tracer: c.tracer,
			query:  query,
			addr:   c.addr,
			dbType: c.dbType,
		}, nil
	}
	return c.Prepare(query)
}

// Close implements driver.Conn Close
func (c *conn) Close() error {
	return c.conn.Close()
}

// Begin implements driver.Conn Begin
func (c *conn) Begin() (driver.Tx, error) {
	t, err := c.conn.Begin()
	if err != nil {
		return nil, err
	}
	return &tx{
		tx: t,
	}, nil
}

// BeginTx implements driver.ConnBeginTx BeginTx.
// If the underlying Conn does not implements
// driver.ConnBeginTx interface, this method
// will use Begin instead.
func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	span, ctx, err := c.tracer.CreateLocalSpan(ctx, go2sky.WithOperationName(genOpName(c.dbType, "beginTransaction")))
	if err != nil {
		return nil, err
	}
	if connBeginTx, ok := c.conn.(driver.ConnBeginTx); ok {
		t, err := connBeginTx.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &tx{
			tx:   t,
			span: span,
		}, nil
	}
	return c.Begin()
}

// Exec implements driver.Execer Exec
func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.conn.(driver.Execer); ok {
		return execer.Exec(query, args)
	}
	return nil, ErrUnsupportedOp
}

// ExecContext implements driver.ExecerContext ExecContext.
// If the underlying Conn does not implements
// driver.ExecerContext interface, this method
// will use Exec instead.
func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	span, err := c.tracer.CreateExitSpan(ctx, genOpName(c.dbType, "exec"), c.addr, emptyInjectFunc)
	if err != nil {
		return nil, err
	}
	defer span.End()
	span.Tag(TagDbType, string(c.dbType))
	span.Tag(TagDbInstance, c.addr)
	span.Tag(TagDbStatement, query)

	if execerContext, ok := c.conn.(driver.ExecerContext); ok {
		return execerContext.ExecContext(ctx, query, args)
	}

	values, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	return c.Exec(query, values)
}

// Query implements driver.Queryer Query
func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.conn.(driver.Queryer); ok {
		return queryer.Query(query, args)
	}
	return nil, ErrUnsupportedOp
}

// QueryContext implements driver.QueryerContext QueryContext
// If the underlying Conn does not implements
// driver.QueryerContext interface, this method
// will use Query instead.
func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	span, err := c.tracer.CreateExitSpan(ctx, genOpName(c.dbType, "query"), c.addr, emptyInjectFunc)
	if err != nil {
		return nil, err
	}
	defer span.End()
	span.Tag(TagDbType, string(c.dbType))
	span.Tag(TagDbInstance, c.addr)
	span.Tag(TagDbStatement, query)

	if queryerContext, ok := c.conn.(driver.QueryerContext); ok {
		return queryerContext.QueryContext(ctx, query, args)
	}

	values, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	return c.Query(query, values)
}
