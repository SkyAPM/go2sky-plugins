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

package micro

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/metadata"
	"github.com/asim/go-micro/v3/registry"
	"github.com/asim/go-micro/v3/server"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	componentIDGoMicroClient = 5008
	componentIDGoMicroServer = 5009
)

var errTracerIsNil = errors.New("tracer is nil")

type clientWrapper struct {
	client.Client

	sw         *go2sky.Tracer
	reportTags []string
}

// ClientOption allow optional configuration of Client
type ClientOption func(*clientWrapper)

// WithClientWrapperReportTags customize span tags
func WithClientWrapperReportTags(reportTags ...string) ClientOption {
	return func(c *clientWrapper) {
		c.reportTags = append(c.reportTags, reportTags...)
	}
}

// Call is used for client calls
func (s *clientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	span, err := s.sw.CreateExitSpan(ctx, name, req.Service(), func(key, value string) error {
		mda, _ := metadata.FromContext(ctx)
		md := metadata.Copy(mda)
		md[key] = value
		ctx = metadata.NewContext(ctx, md)
		return nil
	})
	if err != nil {
		return err
	}

	span.SetComponent(componentIDGoMicroClient)
	span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

	defer span.End()
	for _, k := range s.reportTags {
		if v, ok := metadata.Get(ctx, k); ok {
			span.Tag(go2sky.Tag(k), v)
		}
	}
	if err = s.Client.Call(ctx, req, rsp, opts...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

// Stream is used streaming
func (s *clientWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	span, err := s.sw.CreateExitSpan(ctx, name, req.Service(), func(key, value string) error {
		mda, _ := metadata.FromContext(ctx)
		md := metadata.Copy(mda)
		md[key] = value
		ctx = metadata.NewContext(ctx, md)
		return nil
	})
	if err != nil {
		return nil, err
	}

	span.SetComponent(componentIDGoMicroClient)
	span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

	defer span.End()
	for _, k := range s.reportTags {
		if v, ok := metadata.Get(ctx, k); ok {
			span.Tag(go2sky.Tag(k), v)
		}
	}
	stream, err := s.Client.Stream(ctx, req, opts...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return stream, err
}

// Publish is used publish message to subscriber
func (s *clientWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
	span, err := s.sw.CreateExitSpan(ctx, name, p.ContentType(), func(key, value string) error {
		mda, _ := metadata.FromContext(ctx)
		md := metadata.Copy(mda)
		md[key] = value
		ctx = metadata.NewContext(ctx, md)
		return nil
	})
	if err != nil {
		return err
	}

	span.SetComponent(componentIDGoMicroClient)
	span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

	defer span.End()
	for _, k := range s.reportTags {
		if v, ok := metadata.Get(ctx, k); ok {
			span.Tag(go2sky.Tag(k), v)
		}
	}
	if err = s.Client.Publish(ctx, p, opts...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

// NewClientWrapper accepts a go2sky Tracer and returns a Client Wrapper
func NewClientWrapper(sw *go2sky.Tracer, options ...ClientOption) client.Wrapper {
	return func(c client.Client) client.Client {
		co := &clientWrapper{
			sw:     sw,
			Client: c,
		}
		for _, option := range options {
			option(co)
		}
		return co
	}
}

// NewCallWrapper accepts an go2sky Tracer and returns a Call Wrapper
func NewCallWrapper(sw *go2sky.Tracer, reportTags ...string) client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			if sw == nil {
				return errTracerIsNil
			}

			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			span, err := sw.CreateExitSpan(ctx, name, req.Service(), func(key, value string) error {
				mda, _ := metadata.FromContext(ctx)
				md := metadata.Copy(mda)
				md[key] = value
				ctx = metadata.NewContext(ctx, md)
				return nil
			})
			if err != nil {
				return err
			}

			span.SetComponent(componentIDGoMicroClient)
			span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

			defer span.End()
			for _, k := range reportTags {
				if v, ok := metadata.Get(ctx, k); ok {
					span.Tag(go2sky.Tag(k), v)
				}
			}
			if err = cf(ctx, node, req, rsp, opts); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

// NewSubscriberWrapper accepts a go2sky Tracer and returns a Handler Wrapper
func NewSubscriberWrapper(sw *go2sky.Tracer, reportTags ...string) server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			if sw == nil {
				return errTracerIsNil
			}

			name := "Sub from " + msg.Topic()
			span, err := sw.CreateExitSpan(ctx, name, msg.ContentType(), func(key, value string) error {
				mda, _ := metadata.FromContext(ctx)
				md := metadata.Copy(mda)
				md[key] = value
				ctx = metadata.NewContext(ctx, md)
				return nil
			})
			if err != nil {
				return err
			}

			span.SetComponent(componentIDGoMicroClient)
			span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

			defer span.End()
			for _, k := range reportTags {
				if v, ok := metadata.Get(ctx, k); ok {
					span.Tag(go2sky.Tag(k), v)
				}
			}
			if err = next(ctx, msg); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

// NewHandlerWrapper accepts a go2sky Tracer and returns a Subscriber Wrapper
func NewHandlerWrapper(sw *go2sky.Tracer, reportTags ...string) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			if sw == nil {
				return errTracerIsNil
			}

			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			span, ctx, err := sw.CreateEntrySpan(ctx, name, func(key string) (string, error) {
				str, _ := metadata.Get(ctx, strings.Title(key))
				return str, nil
			})
			if err != nil {
				return err
			}

			span.SetComponent(componentIDGoMicroServer)
			span.SetSpanLayer(agentv3.SpanLayer_RPCFramework)

			defer span.End()
			for _, k := range reportTags {
				if v, ok := metadata.Get(ctx, k); ok {
					span.Tag(go2sky.Tag(k), v)
				}
			}
			if err = fn(ctx, req, rsp); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}
