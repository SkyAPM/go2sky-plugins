package sql

import (
	"context"
	"database/sql"
	"time"
)

type Stmt struct {
	*sql.Stmt

	db    *DB
	query string
}

func (s *Stmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	span, err := createSpan(ctx, s.db.tracer, s.db.opts, "exec")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if s.db.opts.reportQuery {
		span.Tag(tagDbStatement, s.query)
	}
	if s.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	res, err := s.Stmt.ExecContext(ctx, args)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return res, err
}

func (s *Stmt) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	span, err := createSpan(ctx, s.db.tracer, s.db.opts, "query")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if s.db.opts.reportQuery {
		span.Tag(tagDbStatement, s.query)
	}
	if s.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	rows, err := s.Stmt.QueryContext(ctx, args)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return rows, err
}

func (s *Stmt) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {
	span, err := createSpan(ctx, s.db.tracer, s.db.opts, "query")
	if err != nil {
		return nil
	}
	defer span.End()

	if s.db.opts.reportQuery {
		span.Tag(tagDbStatement, s.query)
	}
	if s.db.opts.reportParam {
		span.Tag(tagDbSqlParameters, argsToString(args))
	}

	return s.Stmt.QueryRowContext(ctx, args)
}
