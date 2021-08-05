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

package kratos

import (
	"context"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/go-kratos/kratos/v2/metadata"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/transport"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	componentIDGoKratosServer = 5010
	componentIDGoKratosClient = 5011
)

func Server(tracer go2sky.Tracer) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				if md, ok := metadata.FromServerContext(ctx); ok {
					span, _, err := tracer.CreateEntrySpan(ctx, tr.Operation(), func(key string) (string, error) {
						return md.Get(key), nil
					})

					if err != nil {
						return nil, err
					}

					defer func() { span.End() }()

					span.SetComponent(componentIDGoKratosServer)
					span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

					reply, err := handler(ctx, req)
					if err != nil {
						span.Error(time.Now(), err.Error())
					}
					return reply, err
				}
			}
			return handler(ctx, req)
		}
	}
}

func Client(tracer go2sky.Tracer) middleware.Middleware {
	return func(handler middleware.Handler) middleware.Handler {
		return func(ctx context.Context, req interface{}) (interface{}, error) {
			if tr, ok := transport.FromServerContext(ctx); ok {
				md, ok := metadata.FromClientContext(ctx)
				if !ok { // if no metadata, create it
					ctx = metadata.NewClientContext(ctx, md)
				}
				span, err := tracer.CreateExitSpan(ctx, tr.Operation(), tr.Endpoint(), func(key, value string) error {
					md.Set(key, value)
					return nil
				})

				if err != nil {
					return nil, err
				}

				defer func() { span.End() }()

				span.SetComponent(componentIDGoKratosClient)
				span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

				reply, err := handler(ctx, req)
				if err != nil {
					span.Error(time.Now(), err.Error())
				}
				return reply, err
			}
			return handler(ctx, req)
		}
	}
}
