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

package micro

import (
	"context"
	"fmt"
	"github.com/SkyAPM/go2sky"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/server"
	"time"
)


type swWrapper struct {
	sw	go2sky.Tracer
	client.Client
}


func (s *swWrapper) Call (ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	span, err:= s.sw.CreateExitSpan(ctx, req.Endpoint(), req.Service(), nil)
	if err != nil {
		return err
	}
	md, ok := metadata.FromContext(ctx)
	if !ok {
		md = make(metadata.Metadata)
	}
	defer span.End()
	//TODO
	span.Tag(go2sky.TagHTTPMethod, req.Method())
	for k, v := range md {
		span.Tag(go2sky.Tag(k), v)
	}
	if err= s.Client.Call(ctx, req, rsp, opts ...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

/*func (s *swWrapper) Stream (ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	span, err := s.sw.CreateExitSpan(ctx, req.Endpoint(), req.Service(), nil)
	if err != nil {
		return nil, err
	}
	defer span.End()
	stream, err := s.Client.Stream(ctx, req, opts ...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return stream, err
}*/

/*func (s *swWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
	span, err := s.sw.CreateExitSpan(ctx, name, "", nil)
	if err != nil {
		return err
	}
	defer span.End()
	if err = s.Client.Publish(ctx, p, opts...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}*/

//NewClientWrapper accepts a go2sky Tracer and returns a Client Wrapper
func NewClientWrapper (sw *go2sky.Tracer) client.Wrapper {
	return func(c client.Client) client.Client {
		return &swWrapper{sw:*sw, Client: c}
	}
}

//NewHandlerWrapper accepts a go2sky Tracer and returns a Server Wrapper
func NewHandlerWrapper(sw *go2sky.Tracer) server.HandlerWrapper {
	return func(h server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			span, ctx, err := sw.CreateEntrySpan(ctx, name, nil)
			if err != nil {
				return err
			}
			//TODO
			span.Tag(go2sky.TagHTTPMethod, req.Method())
			if err = h(ctx, req, rsp); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}
