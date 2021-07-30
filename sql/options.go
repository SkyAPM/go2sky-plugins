//
// Copyright 2021 SkyAPM org
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
	"regexp"
)

type DBType string

const (
	MYSQL DBType = "mysql"
	IPV4  DBType = "others"
)

type Option func(*options)

type options struct {
	dbType      DBType
	peer        string
	componentID int32

	reportQuery bool
	reportParam bool
}

// WithSqlDBType set dbType option,
// dbType is used for parsing dsn string to peer address
// and setting componentID, if DB type is not support in DBType
// list, please use WithPeerAddr to set peer address manually
func WithSqlDBType(t DBType) Option {
	return func(o *options) {
		o.dbType = t
	}
}

// WithPeerAddr set the peer address to report
func WithPeerAddr(addr string) Option {
	return func(o *options) {
		o.peer = addr
	}
}

func WithQueryReport() Option {
	return func(o *options) {
		o.reportQuery = true
	}
}

func WithParamReport() Option {
	return func(o *options) {
		o.reportParam = true
	}
}

func (o options) getOpName(op string) string {
	switch o.dbType {
	case MYSQL:
		return "Mysql/Go2Sky/" + op
	default:
		return "Sql/Go2Sky/" + op
	}
}

// setPeerWithDsn parse dsn to a endpoint addr string (host:port)
func (o *options) setPeerWithDsn(dsn string) {
	var addr string
	switch o.dbType {
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
	o.peer = addr
}

func (o *options) setComponentID() {
	switch o.dbType {
	case MYSQL:
		o.componentID = componentIDMysql
	default:
		o.componentID = componentIDUnknown
	}
}
