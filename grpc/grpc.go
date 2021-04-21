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

package grpc

import (
	"context"
	"log"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	language_agent "github.com/SkyAPM/go2sky/reporter/grpc/language-agent"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

const (
	componentIDGOGrpcServer = 23 //same with ComponentsDefine in java
	componentIDGOGrpcClient = 23 //same with ComponentsDefine in java
)

//NewUnaryServerTraceInterceptor return grpc unary server interceptor of skywalking trace
func NewUnaryServerTraceInterceptor(tracer *go2sky.Tracer) grpc.UnaryServerInterceptor {
	if tracer == nil {
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
		span, spanCtx, err := tracer.CreateEntrySpan(ctx, info.FullMethod, func() (string, error) {
			md, ok := metadata.FromIncomingContext(ctx)
			if !ok {
				return "", nil
			}

			values := md.Get(propagation.Header)
			if len(values) == 0 || len(values[0]) == 0 {
				return "", nil
			}
			return values[0], nil
		})

		if err != nil {
			log.Printf("create entry span error. %v\n", err)
			return handler(ctx, req)
		}
		ctx = spanCtx

		span.SetComponent(componentIDGOGrpcServer)
		span.Tag(go2sky.TagURL, info.FullMethod)
		span.SetSpanLayer(language_agent.SpanLayer_RPCFramework)

		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in server trace interceptor. %v\n", r)
				err = status.Error(codes.Unknown, "panic happened")
			}

			if err != nil {
				span.Error(time.Now(), err.Error())
				span.Tag(go2sky.TagStatusCode, status.Code(err).String())
			}

			span.End()
		}()

		resp, err = handler(ctx, req)

		return
	}
}

//NewUnaryClientTraceInterceptor return grpc unary client interceptor of skywalking trace
func NewUnaryClientTraceInterceptor(tracer *go2sky.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) (err error) {
		span, err := tracer.CreateExitSpan(ctx, method, cc.Target(), func(header string) error {
			md := metadata.New(map[string]string{propagation.Header: header})
			ctx = metadata.NewOutgoingContext(ctx, md)
			return nil
		})
		if err != nil {
			log.Printf("create exit span error. %v\n", err)
			err = invoker(ctx, method, req, reply, cc, opts...)
			return
		}

		span.SetComponent(componentIDGOGrpcClient)
		span.Tag(go2sky.TagURL, method)
		span.SetSpanLayer(language_agent.SpanLayer_RPCFramework)

		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic in client trace interceptor. %v\n", r)
				err = status.Error(codes.Unknown, "panic happened")
			}

			if err != nil {
				span.Error(time.Now(), err.Error())
				span.Tag(go2sky.TagStatusCode, status.Code(err).String())
			}

			span.End()
		}()

		err = invoker(ctx, method, req, reply, cc, opts...)

		return
	}
}
