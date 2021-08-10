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

type Conn struct {
	*sql.Conn

	db *DB
}

func (c *Conn) PingContext(ctx context.Context) error {

}

func (c *Conn) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {

}

func (c *Conn) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {

}

func (c *Conn) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sqlRow {

}

func (c *Conn) PrepareContext(ctx context.Context, query string) (*Stmt, error) {

}

func (c *Conn) BeginTx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {

}

type Tx struct {
	*sql.Tx

	db   *DB
	span go2sky.Span
}

func (tx *Tx) Commit() error {

}

func (tx *Tx) Rollback() error {

}

func (tx *Tx) PrepareContext(ctx context.Context, query string) (*Stmt, error) {

}

func (tx *Tx) StmtContext(ctx context.Context, stmt *Stmt) *Stmt {

}

func (tx *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {

}

func (tx *Tx) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {

}

func (tx *Tx) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {

}

type Stmt struct {
	*sql.Stmt

	db    *DB
	query string
}

func (s *Stmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {

}

func (s *Stmt) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {

}

func (s *Stmt) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {

}
