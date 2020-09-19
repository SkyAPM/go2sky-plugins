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

// Package micro (sw_micro) is a plugin that can be used to trace Go-micro framework.

package go_micro

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/micro/go-micro/registry"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
)

type swWrapper struct {
	sw go2sky.Tracer
	client.Client
}

func (s *swWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	mm, _ := metadata.FromContext(ctx)
	span, err := s.sw.CreateExitSpan(ctx, name, req.Service(), func(header string) error {
		mda, _ := metadata.FromContext(ctx)
		md := metadata.Copy(mda)
		md[propagation.Header] = header
		ctx = metadata.NewContext(ctx, md)
		return nil
	})
	if err != nil {
		return err
	}

	defer span.End()
	for k, v := range mm {
		span.Tag(go2sky.Tag(mm[k]), v)
	}
	if err = s.Client.Call(ctx, req, rsp, opts...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

func (s *swWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	mm, _ := metadata.FromContext(ctx)
	span, err := s.sw.CreateExitSpan(ctx, name, req.Service(), func(header string) error {
		mda, _ := metadata.FromContext(ctx)
		md := metadata.Copy(mda)
		md[propagation.Header] = header
		ctx = metadata.NewContext(ctx, md)
		return nil
	})
	if err != nil {
		return nil, err
	}
	defer span.End()
	for k, v := range mm {
		span.Tag(go2sky.Tag(mm[k]), v)
	}
	stream, err := s.Client.Stream(ctx, req, opts...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return stream, err
}

func (s *swWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
	mm, _ := metadata.FromContext(ctx)
	span, err := s.sw.CreateExitSpan(ctx, name, p.ContentType(), func(header string) error {
		mda, _ := metadata.FromContext(ctx)
		md := metadata.Copy(mda)
		md[propagation.Header] = header
		ctx = metadata.NewContext(ctx, md)
		return nil
	})
	if err != nil {
		return err
	}
	defer span.End()
	for k, v := range mm {
		span.Tag(go2sky.Tag(mm[k]), v)
	}
	if err = s.Client.Publish(ctx, p, opts...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

//NewClientWrapper return a client with wrapper
func NewClientWrapper(sw *go2sky.Tracer) client.Wrapper {
	return func(c client.Client) client.Client {
		return &swWrapper{sw: *sw, Client: c}
	}
}

//NewCallWrapper return call with wrapper
func NewCallWrapper(sw *go2sky.Tracer) client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			if sw == nil {
				return errors.New("tracer is nil")
			}
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			mm, _ := metadata.FromContext(ctx)
			span, err := sw.CreateExitSpan(ctx, name, req.Service(), func(header string) error {
				mda, _ := metadata.FromContext(ctx)
				md := metadata.Copy(mda)
				md[propagation.Header] = header
				ctx = metadata.NewContext(ctx, md)
				return nil
			})
			if err != nil {
				return err
			}
			defer span.End()
			for k, v := range mm {
				span.Tag(go2sky.Tag(mm[k]), v)
			}
			if err = cf(ctx, node, req, rsp, opts); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

//NewSubscriberWrapper return a subscriber with wrapper
func NewSubscriberWrapper(sw *go2sky.Tracer) server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			name := "Sub from " + msg.Topic()
			mm, _ := metadata.FromContext(ctx)
			if sw == nil {
				return errors.New("tracer is nil")
			}
			span, err := sw.CreateExitSpan(ctx, name, msg.ContentType(), func(header string) error {
				mda, _ := metadata.FromContext(ctx)
				md := metadata.Copy(mda)
				md[propagation.Header] = header
				ctx = metadata.NewContext(ctx, md)
				return nil
			})
			if err != nil {
				return err
			}
			defer span.End()
			for k, v := range mm {
				span.Tag(go2sky.Tag(mm[k]), v)
			}
			if err = next(ctx, msg); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

//NewHandlerWrapper return a server.Hanler with wrapper
func NewHandlerWrapper(sw *go2sky.Tracer) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			mm, _ := metadata.FromContext(ctx)
			span, ctx, err := sw.CreateEntrySpan(ctx, name, func() (string, error) {
				str, _ := metadata.Get(ctx, "Sw8")
				return str, nil
			})
			if err != nil {
				return err
			}
			defer span.End()
			for k, v := range mm {
				span.Tag(go2sky.Tag(mm[k]), v)
			}
			if err = fn(ctx, req, rsp); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}
