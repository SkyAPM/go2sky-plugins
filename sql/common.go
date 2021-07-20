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
	"errors"
	"fmt"
	"strings"

	"github.com/SkyAPM/go2sky"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	componentIDUnknown = 0
	componentIDMysql   = 5010
	componentIDSqlite  = 5011
)

const (
	tagDbType          = "db.type"
	tagDbInstance      = "db.instance"
	tagDbStatement     = "db.statement"
	tagDbSqlParameters = "db.sql.parameters"
)

var ErrUnsupportedOp = errors.New("operation unsupported by the underlying driver")

// emptyInjectFunc defines a empty injector for propagation.Injector function
func emptyInjectFunc(key, value string) error { return nil }

// namedValueToValueString converts driver arguments of NamedValue format to Value string format.
func namedValueToValueString(named []driver.NamedValue) string {
	b := make([]string, 0, len(named))
	for _, param := range named {
		b = append(b, fmt.Sprintf("%v", param.Value))
	}
	return strings.Join(b, ",")
}

// namedValueToValue converts driver arguments of NamedValue format to Value format.
// Implemented in the same way as in database/sql/ctxutil.go.
func namedValueToValue(named []driver.NamedValue) ([]driver.Value, error) {
	dargs := make([]driver.Value, len(named))
	for n, param := range named {
		if len(param.Name) > 0 {
			return nil, errors.New("sql: driver does not support the use of Named Parameters")
		}
		dargs[n] = param.Value
	}
	return dargs, nil
}

func createSpan(ctx context.Context, tracer *go2sky.Tracer, opts *options, operation string) (go2sky.Span, error) {
	s, _, err := tracer.CreateLocalSpan(ctx,
		go2sky.WithSpanType(go2sky.SpanTypeExit),
		go2sky.WithOperationName(opts.getOpName(operation)),
	)
	if err != nil {
		return nil, err
	}
	s.SetPeer(opts.peer)
	s.SetComponent(opts.componentID)
	s.SetSpanLayer(agentv3.SpanLayer_Database)
	return s, nil
}
