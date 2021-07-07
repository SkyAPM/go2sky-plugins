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

// stmt is a tracing wrapper for driver.Stmt
type stmt struct {
	stmt   driver.Stmt
	tracer *go2sky.Tracer

	// query defines the statement query
	query string
	// addr defines the address of sql server, format in host:port
	addr string
	// dbType defines the sql server type
	dbType DBType
}

func (s *stmt) Close() error {
	return s.stmt.Close()
}

func (s *stmt) NumInput() int {
	return s.stmt.NumInput()
}

func (s *stmt) Exec(args []driver.Value) (driver.Result, error) {
	return s.stmt.Exec(args)
}

func (s *stmt) ExecContext(ctx context.Context, args []driver.NamedValue) (driver.Result, error) {
	span, err := s.tracer.CreateExitSpan(ctx, genOpName(s.dbType, "exec"), s.addr, emptyInjectFunc)
	if err != nil {
		return nil, err
	}
	defer span.End()
	span.Tag(TagDbType, string(s.dbType))
	span.Tag(TagDbInstance, s.addr)
	span.Tag(TagDbStatement, s.query)

	if execerContext, ok := s.stmt.(driver.StmtExecContext); ok {
		return execerContext.ExecContext(ctx, args)
	}

	values, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	return s.Exec(values)
}

func (s *stmt) Query(args []driver.Value) (driver.Rows, error) {
	return s.stmt.Query(args)
}

func (s *stmt) QueryContext(ctx context.Context, args []driver.NamedValue) (driver.Rows, error) {
	span, err := s.tracer.CreateExitSpan(ctx, genOpName(s.dbType, "query"), s.addr, emptyInjectFunc)
	if err != nil {
		return nil, err
	}
	defer span.End()
	span.Tag(TagDbType, string(s.dbType))
	span.Tag(TagDbInstance, s.addr)
	span.Tag(TagDbStatement, s.query)

	if queryer, ok := s.stmt.(driver.StmtQueryContext); ok {
		return queryer.QueryContext(ctx, args)
	}

	values, err := namedValueToValue(args)
	if err != nil {
		return nil, err
	}
	return s.Query(values)
}
