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
	"regexp"
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

type attribute struct {
	dbType      DBType
	peer        string
	componentID int32
}

var ErrUnsupportedOp = errors.New("operation unsupported by the underlying driver")

// emptyInjectFunc defines a empty injector for propagation.Injector function
func emptyInjectFunc(key, value string) error { return nil }

func newAttribute(name string, dbType DBType) attribute {
	a := attribute{
		dbType: dbType,
	}
	a.setPeer(name)
	a.setComponentID()
	return a
}

// setPeer parse dsn to a endpoint addr string (host:port)
func (a attribute) setPeer(dsn string) {
	var addr string
	switch a.dbType {
	case MYSQL:
		// [user[:password]@][net[(addr)]]/dbname[?param1=value1&paramN=valueN]
		re := regexp.MustCompile(`\(.+\)`)
		addr = re.FindString(dsn)
		addr = addr[1 : len(addr)-1]
	case IPV4:
		// ipv4 addr
		re := regexp.MustCompile(`((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})(\.((2(5[0-5]|[0-4]\d))|[0-1]?\d{1,2})){3}:\d{1,5}`)
		addr = re.FindString(dsn)
	}
	a.peer = addr
	return
}

func (a attribute) getOpName(op string) string {
	switch a.dbType {
	case MYSQL:
		return "Mysql/Go2Sky/" + op
	default:
		return "Sql/Go2Sky/" + op
	}
}

func (a attribute) setComponentID() {
	switch a.dbType {
	case MYSQL:
		a.componentID = componentIDMysql
	default:
		a.componentID = componentIDUnknown
	}
}

// namedValueToValueString converts driver arguments of NamedValue format to Value string format.
func namedValueToValueString(named []driver.NamedValue) string {
	b := strings.Builder{}
	for _, param := range named {
		b.WriteString(fmt.Sprintf("%v,", param.Value))
	}
	return b.String()
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

func createSpan(ctx context.Context, tracer *go2sky.Tracer, attr attribute, operation string) (go2sky.Span, error) {
	s, _, err := tracer.CreateLocalSpan(ctx,
		go2sky.WithSpanType(go2sky.SpanTypeExit),
		go2sky.WithOperationName(attr.getOpName(operation)),
	)
	if err != nil {
		return nil, err
	}
	s.SetPeer(attr.peer)
	s.SetComponent(attr.componentID)
	s.SetSpanLayer(agentv3.SpanLayer_Database)
	return s, nil
}
