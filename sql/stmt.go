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

// Stmt wrap sql.Stmt and support trace
type Stmt struct {
	*sql.Stmt

	db    *DB
	query string
}

// ExecContext support trace
func (s *Stmt) ExecContext(ctx context.Context, args ...interface{}) (sql.Result, error) {
	span, err := createSpan(ctx, s.db.tracer, s.db.opts, "execute")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if s.db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, s.query)
	}
	if s.db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	res, err := s.Stmt.ExecContext(ctx, args...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return res, err
}

// QueryContext support trace
func (s *Stmt) QueryContext(ctx context.Context, args ...interface{}) (*sql.Rows, error) {
	span, err := createSpan(ctx, s.db.tracer, s.db.opts, "query")
	if err != nil {
		return nil, err
	}
	defer span.End()

	if s.db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, s.query)
	}
	if s.db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	rows, err := s.Stmt.QueryContext(ctx, args...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return rows, err
}

// QueryRowContext support trace
func (s *Stmt) QueryRowContext(ctx context.Context, args ...interface{}) *sql.Row {
	span, err := createSpan(ctx, s.db.tracer, s.db.opts, "query")
	if err != nil {
		return nil
	}
	defer span.End()

	if s.db.opts.reportQuery {
		span.Tag(go2sky.TagDBStatement, s.query)
	}
	if s.db.opts.reportParam {
		span.Tag(go2sky.TagDBSqlParameters, argsToString(args))
	}

	return s.Stmt.QueryRowContext(ctx, args...)
}
