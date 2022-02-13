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
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/SkyAPM/go2sky"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	componentIDUnknown = 0
	componentIDMysql   = 5012
)

// ErrUnsupportedOp operation unsupported by the underlying driver
var ErrUnsupportedOp = errors.New("operation unsupported by the underlying driver")

func argsToString(args []interface{}) string {
	sb := strings.Builder{}
	for _, arg := range args {
		sb.WriteString(fmt.Sprintf("%v, ", arg))
	}
	return sb.String()
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
	s.Tag(go2sky.TagDBType, string(opts.dbType))
	s.Tag(go2sky.TagDBInstance, opts.peer)
	return s, nil
}

// parseDsn parse dsn to a endpoint addr string (host:port)
func parseDsn(dbType DBType, dsn string) string {
	var addr string
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
	return addr
}
