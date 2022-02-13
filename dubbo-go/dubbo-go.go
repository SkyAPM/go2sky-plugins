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

package dubbogo

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/SkyAPM/go2sky"

	"dubbo.apache.org/dubbo-go/v3/common/extension"
	"dubbo.apache.org/dubbo-go/v3/filter"
	"dubbo.apache.org/dubbo-go/v3/protocol"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

// side record the server or client
type side int

const (
	componentID = 3

	serverSide side = iota
	clientSide side = iota
)

var (
	errInvalidTracer = errors.New("invalid tracer")
)

// this should be executed before users set their own Tracer
func init() {
	// init default tracer
	defaultTracer, _ = go2sky.NewTracer("default-dubbo-go-tracer")

	// set filter
	// in dubbo go conf:
	//     client side: filter: "go2sky-tracing-client"
	//     server side: filter: "go2sky-tracing-server"
	// maybe move to constant
	extension.SetFilter("go2sky-tracing-client", GetClientTracingFilterSingleton)
	extension.SetFilter("go2sky-tracing-server", GetServerTracingFilterSingleton)
}

// default go2sky tracer
var defaultTracer *go2sky.Tracer

type tracingFilter struct {
	tracer     *go2sky.Tracer
	extraTags  map[string]string
	reportTags []string
	side       side
}

// SetClientTracer set client tracer with user's tracer.
func SetClientTracer(tracer *go2sky.Tracer) error {
	if tracer == nil {
		return errInvalidTracer
	}

	clientFilter.tracer = tracer

	return nil
}

// SetServerTracer set server tracer with user's tracer.
func SetServerTracer(tracer *go2sky.Tracer) error {
	if tracer == nil {
		return errInvalidTracer
	}

	serverFilter.tracer = tracer

	return nil
}

// SetClientExtraTags adds extra tag to client tracer spans.
func SetClientExtraTags(key string, value string) {
	if clientFilter.extraTags == nil {
		clientFilter.extraTags = make(map[string]string)
	}

	clientFilter.extraTags[key] = value
}

// SetServerExtraTags adds extra tag to server tracer spans.
func SetServerExtraTags(key string, value string) {
	if serverFilter.extraTags == nil {
		serverFilter.extraTags = make(map[string]string)
	}

	serverFilter.extraTags[key] = value
}

// SetClientReportTags adds report tags to client tracer spans.
func SetClientReportTags(tags ...string) {
	clientFilter.reportTags = append(clientFilter.reportTags, tags...)
}

// SetServerReportTags adds report tags to server tracer spans.
func SetServerReportTags(tags ...string) {
	serverFilter.reportTags = append(serverFilter.reportTags, tags...)
}

// Invoke implements dubbo-go filter interface.
func (cf tracingFilter) Invoke(ctx context.Context, invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	operationName := invoker.GetURL().ServiceKey() + "#" + invocation.MethodName()
	var span go2sky.Span

	if cf.side == clientSide {
		span, _ = cf.tracer.CreateExitSpan(ctx, operationName, invoker.GetURL().Location, func(key, value string) error {
			invocation.SetAttachments(key, value)
			return nil
		})

		span.SetComponent(componentID)
	} else {
		// componentIDGo2SkyServer
		span, ctx, _ = cf.tracer.CreateEntrySpan(ctx, operationName, func(key string) (string, error) {
			return invocation.AttachmentsByKey(key, ""), nil
		})

		span.SetComponent(componentID)
	}
	defer span.End()

	// add extra tags
	for k, v := range cf.extraTags {
		span.Tag(go2sky.Tag(k), v)
	}

	// add report tags
	for _, tag := range cf.reportTags {
		// from attachments
		if v, ok := invocation.Attachments()[tag].(string); ok {
			if ok {
				span.Tag(go2sky.Tag(tag), v)
			}
		}
		// or from url
		if v := invoker.GetURL().GetParam(tag, ""); v != "" {
			span.Tag(go2sky.Tag(tag), v)
		}
	}

	// other tags ...
	span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

	result := invoker.Invoke(ctx, invocation)

	// finish span
	// defer span.End()
	if result.Error() != nil {
		// tag error
		span.Error(time.Now(), result.Error().Error())
	}
	return result
}

// OnResponse implements dubbo-go filter interface.
func (cf tracingFilter) OnResponse(ctx context.Context, result protocol.Result, invoker protocol.Invoker, invocation protocol.Invocation) protocol.Result {
	return result
}

var (
	serverFilterOnce sync.Once
	serverFilter     *tracingFilter

	clientFilterOnce sync.Once
	clientFilter     *tracingFilter
)

// GetServerTracingFilterSingleton returns global server filter for server side.
func GetServerTracingFilterSingleton() filter.Filter {
	serverFilterOnce.Do(func() {
		serverFilter = &tracingFilter{
			tracer: defaultTracer,
			side:   serverSide,
		}
	})
	return serverFilter
}

// GetClientTracingFilterSingleton returns global filter for client side.
func GetClientTracingFilterSingleton() filter.Filter {
	clientFilterOnce.Do(func() {
		clientFilter = &tracingFilter{
			tracer: defaultTracer,
			side:   clientSide,
		}
	})
	return clientFilter
}
