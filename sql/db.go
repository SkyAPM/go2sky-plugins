package sql

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"time"

	"github.com/SkyAPM/go2sky"
)

type DB struct {
	*sql.DB

	tracer *go2sky.Tracer
	opts   *options
}

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

	return &DB{
		DB:     db,
		tracer: tracer,
		opts:   options,
	}, nil
}

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

func (db *DB) PrepareContext(ctx context.Context, query string) (*Stmt, error) {
	stmt, err := db.DB.PrepareContext(ctx, query)
	if err != nil {
		return nil, err
	}
	return &Stmt{
		Stmt:  stmt,
		query: query,
	}, nil
}

func (db *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	span, err := createSpan(ctx, db.tracer, db.opts, "exec")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	res, err := db.DB.ExecContext(ctx, query, args)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return res, err
}

func (db *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	span, err := createSpan(ctx, db.tracer, db.opts, "query")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	rows, err := db.DB.QueryContext(ctx, query, args)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return rows, err
}

func (db *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	span, err := createSpan(ctx, db.tracer, db.opts, "query")
	if err != nil {
		return nil
	}
	defer span.End()

	if db.opts.reportQuery {
		span.Tag(tagDbStatement, query)
	}
	if db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	return db.DB.QueryRowContext(ctx, query, args)
}

func (db *DB) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	span, err := createSpan(ctx, db.tracer, db.opts, "transaction")
	if err != nil {
		return nil, err
	}

	tx, err := db.DB.BeginTx(ctx, opts)
	if err != nil {
		span.Error(time.Now(), err.Error())
		span.End()
		return nil, err
	}
	return &Tx{
		Tx:   tx,
		db:   db,
		span: span,
	}, nil
}

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
