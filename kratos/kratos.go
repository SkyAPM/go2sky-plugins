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

package kratos

import (
	"context"
	"fmt"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	componentIDKratos = 5010
)

// Option allows for functional options to adjust behaviour
type Option func(*options)

type options struct {
	reportTags []string
}

// WithReportTags will set tags that need to report in metadata
func WithReportTags(tags ...string) Option {
	return func(o *options) {
		o.reportTags = append(o.reportTags, tags...)
	}
}

// Server go2sky middleware for kratos server
func Server(tracer *go2sky.Tracer, opts ...Option) middleware.Middleware {
	options := &options{
		reportTags: []string{},
	}
	for _, o := range opts {
		o(options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(c context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(c); ok {
				span, ctx, err := tracer.CreateEntrySpan(c, tr.Operation(), func(key string) (string, error) {
					return tr.RequestHeader().Get(key), nil
				})
				if err != nil {
					return nil, err
				}
				defer func() { span.End() }()

				span.SetComponent(componentIDKratos)
				span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

				if md, ok := metadata.FromServerContext(ctx); ok {
					for _, k := range options.reportTags {
						span.Tag(go2sky.Tag(k), md.Get(k))
					}
				}

				reply, err := handler(ctx, req)
				if err != nil {
					span.Error(time.Now(), err.Error())
				}
				return reply, err
			}
			return handler(c, req)
		}
	}
}

// Client go2sky middleware for kratos client
func Client(tracer *go2sky.Tracer, opts ...Option) middleware.Middleware {
	options := &options{
		reportTags: []string{},
	}
	for _, o := range opts {
		o(options)
	}
	return func(handler middleware.Handler) middleware.Handler {
		return func(c context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromClientContext(c); ok {
				span, ctx, err := tracer.CreateExitSpanWithContext(c, tr.Operation(), tr.Endpoint(), func(key, value string) error {
					tr.RequestHeader().Set(key, value)
					return nil
				})
				if err != nil {
					return nil, err
				}
				defer func() { span.End() }()

				span.SetComponent(componentIDKratos)
				span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

				if md, ok := metadata.FromClientContext(ctx); ok {
					for _, k := range options.reportTags {
						span.Tag(go2sky.Tag(k), md.Get(k))
					}
				}

				reply, err := handler(ctx, req)
				if err != nil {
					span.Error(time.Now(), err.Error())
				}
				return reply, err
			}
			return handler(c, req)
		}
	}
}

// TraceID inject the current traceId into the kratos log
func TraceID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if id := go2sky.TraceID(ctx); id != go2sky.EmptyTraceID {
			return id
		}
		return ""
	}
}

// SpanID inject the current spanId into the kratos log
func SpanID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if id := go2sky.SpanID(ctx); id != go2sky.EmptySpanID {
			return fmt.Sprintf("%d", id)
		}
		return ""
	}
}

// SegmentID inject the current segmentId into the kratos log
func SegmentID() log.Valuer {
	return func(ctx context.Context) interface{} {
		if id := go2sky.TraceSegmentID(ctx); id != go2sky.EmptyTraceSegmentID {
			return id
		}
		return ""
	}
}
