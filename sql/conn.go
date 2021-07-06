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

// Ping implements driver.Pinger interface,
// If the underlying Conn does not implement Pinger,
// Ping will return a ErrUnsupportedOp error
func (c *conn) Ping(ctx context.Context) error {
	if pinger, ok := c.conn.(driver.Pinger); ok {
		s, err := c.tracer.CreateExitSpan(ctx, genOpName(c.dbType, "ping"), c.addr, emptyInjectFunc)
		if err != nil {
			return err
		}
		defer s.End()
		return pinger.Ping(ctx)
	}
	return ErrUnsupportedOp
}

func (c *conn) Prepare(query string) (driver.Stmt, error) {
	st, err := c.conn.Prepare(query)
	if err != nil {
		return nil, err
	}
	return stmt{
		stmt:   st,
		tracer: c.tracer,
		addr:   c.addr,
	}, nil
}

// PrepareContext implements driver.ConnPrepareContext PrepareContext
func (c *conn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if ConnPrepareContext, ok := c.conn.(driver.ConnPrepareContext); ok {
		st, err := ConnPrepareContext.PrepareContext(ctx, query)
		if err != nil {
			return nil, err
		}
		return &stmt{
			stmt:   st,
			tracer: c.tracer,
			addr:   c.addr,
		}, nil
	}
	return nil, ErrUnsupportedOp
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
		tx:     t,
		tracer: c.tracer,
	}, nil
}

// BeginTx implements driver.ConnBeginTx BeginTx
func (c *conn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	s, ctx, err := c.tracer.CreateLocalSpan(ctx, go2sky.WithOperationName(genOpName(c.dbType, "beginTransaction")))
	if err != nil {
		return nil, err
	}
	if connBeginTx, ok := c.conn.(driver.ConnBeginTx); ok {
		t, err := connBeginTx.BeginTx(ctx, opts)
		if err != nil {
			return nil, err
		}
		return &tx{
			tx:     t,
			tracer: c.tracer,
			span:   s,
		}, nil
	}
	return nil, ErrUnsupportedOp
}

func (c *conn) Exec(query string, args []driver.Value) (driver.Result, error) {
	if execer, ok := c.conn.(driver.Execer); ok {
		return execer.Exec(query, args)
	}
	return nil, ErrUnsupportedOp
}

func (c *conn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	s, err := c.tracer.CreateExitSpan(ctx, genOpName(c.dbType, "exec"), c.addr, emptyInjectFunc)
	if err != nil {
		return nil, err
	}
	s.Tag(TagDbType, string(c.dbType))
	s.Tag(TagDbInstance, c.addr)
	s.Tag(TagDbStatement, query)
	defer s.End()
	if execerContext, ok := c.conn.(driver.ExecerContext); ok {
		return execerContext.ExecContext(ctx, query, args)
	}
	return nil, ErrUnsupportedOp
}

// Query implements driver.Queryer Query
func (c *conn) Query(query string, args []driver.Value) (driver.Rows, error) {
	if queryer, ok := c.conn.(driver.Queryer); ok {
		return queryer.Query(query, args)
	}
	return nil, ErrUnsupportedOp
}

// QueryContext implements driver.QueryerContext QueryContext
func (c *conn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	s, err := c.tracer.CreateExitSpan(ctx, genOpName(c.dbType, "query"), c.addr, emptyInjectFunc)
	if err != nil {
		return nil, err
	}
	s.Tag(TagDbType, string(c.dbType))
	s.Tag(TagDbInstance, c.addr)
	s.Tag(TagDbStatement, query)
	defer s.End()
	if queryerContext, ok := c.conn.(driver.QueryerContext); ok {
		return queryerContext.QueryContext(ctx, query, args)
	}
	return nil, ErrUnsupportedOp
}
