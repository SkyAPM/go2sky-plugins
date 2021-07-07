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
	"database/sql/driver"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	TagDbType          = "db.type"
	TagDbInstance      = "db.instance"
	TagDbStatement     = "db.statement"
	TagDbSqlParameters = "db.sql.parameters"
)

var ErrUnsupportedOp = errors.New("operation unsupported by the underlying driver")

// emptyInjectFunc defines a empty injector for propagation.Injector function
func emptyInjectFunc(key, value string) error { return nil }

// parseAddr parse dsn to a endpoint addr string (host:port)
func parseAddr(dsn string, dbType DBType) (addr string) {
	switch dbType {
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
	return
}

func genOpName(dbType DBType, opName string) string {
	switch dbType {
	case MYSQL:
		return "Mysql/Go2Sky/" + opName
	default:
		return "Sql/Go2Sky/" + opName
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
