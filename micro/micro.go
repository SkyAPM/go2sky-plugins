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
package micro

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/micro/go-micro/registry"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/micro/go-micro/client"
	"github.com/micro/go-micro/metadata"
	"github.com/micro/go-micro/server"
)

type ClientWrapper struct {
	sw *go2sky.Tracer
	client.Client
	extraTags metadata.Metadata
}

type handlerWrapper struct {
	sw        *go2sky.Tracer
	extraTags metadata.Metadata
}

type callWrapper struct {
	sw        *go2sky.Tracer
	extraTags metadata.Metadata
}

type subscriberWrapper struct {
	sw        *go2sky.Tracer
	extraTags metadata.Metadata
}

// ClientOption allow optional configuration of Client
type ClientOption func(*ClientWrapper)

// HandlerOption allow optional configuration of Handler
type HandlerOption func(*handlerWrapper)

// CallOption allow optional configuration of Call
type CallOption func(*callWrapper)

// SubscriberOption allow optional configuration of Subscriber
type SubscriberOption func(*subscriberWrapper)

func WithClientTag(key string) ClientOption {
	return func(c *ClientWrapper) {
		if c.extraTags == nil {
			c.extraTags = make(metadata.Metadata)
		}
		c.extraTags[key] = ""
	}
}

func WithHandlerTag(key string) HandlerOption {
	return func(h *handlerWrapper) {
		if h.extraTags == nil {
			h.extraTags = make(metadata.Metadata)
		}
		h.extraTags[key] = ""
	}
}

func WithCallTag(key string) CallOption {
	return func(c *callWrapper) {
		if c.extraTags == nil {
			c.extraTags = make(metadata.Metadata)
		}
		c.extraTags[key] = ""
	}
}

func WithSubscriberTag(key string) SubscriberOption {
	return func(s *subscriberWrapper) {
		if s.extraTags == nil {
			s.extraTags = make(metadata.Metadata)
		}
		s.extraTags[key] = ""
	}
}

func (s *ClientWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
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
	for k := range s.extraTags {
		v, ok := metadata.Get(ctx, k)
		if !ok {
			log.Fatalf("set tag %s failed \n", k)
		}
		span.Tag(go2sky.Tag(k), v)
	}
	if err = s.Client.Call(ctx, req, rsp, opts...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

func (s *ClientWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
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
	for k := range s.extraTags {
		v, ok := metadata.Get(ctx, k)
		if !ok {
			log.Fatalf("set tag %s failed \n", k)
		}
		span.Tag(go2sky.Tag(k), v)
	}
	stream, err := s.Client.Stream(ctx, req, opts...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return stream, err
}

func (s *ClientWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
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
	for k := range s.extraTags {
		v, ok := metadata.Get(ctx, k)
		if !ok {
			log.Fatalf("set tag %s failed \n", k)
		}
		span.Tag(go2sky.Tag(k), v)
	}
	if err = s.Client.Publish(ctx, p, opts...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}

//NewClientWrapper return a client with wrapper
func NewClientWrapper(sw *go2sky.Tracer, options ...ClientOption) client.Wrapper {
	return func(c client.Client) client.Client {
		co := &ClientWrapper{
			sw:     sw,
			Client: c,
		}
		for _, option := range options {
			option(co)
		}
		return co
	}
}

//NewCallWrapper return call with wrapper
func NewCallWrapper(sw *go2sky.Tracer, options ...CallOption) client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			if sw == nil {
				return errors.New("tracer is nil")
			}
			cl := &callWrapper{
				sw: sw,
			}
			for _, option := range options {
				option(cl)
			}
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
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
			for k := range cl.extraTags {
				v, ok := metadata.Get(ctx, k)
				if !ok {
					log.Fatalf("set tag %s failed \n", k)
				}
				span.Tag(go2sky.Tag(k), v)
			}
			if err = cf(ctx, node, req, rsp, opts); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

//NewSubscriberWrapper return a subscriber with wrapper
func NewSubscriberWrapper(sw *go2sky.Tracer, options ...SubscriberOption) server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			so := &subscriberWrapper{
				sw: sw,
			}
			for _, option := range options {
				option(so)
			}
			name := "Sub from " + msg.Topic()
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
			for k := range so.extraTags {
				v, ok := metadata.Get(ctx, k)
				if !ok {
					log.Fatalf("set tag %s failed \n", k)
				}
				span.Tag(go2sky.Tag(k), v)
			}
			if err = next(ctx, msg); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

//NewHandlerWrapper return a server.Hanler with wrapper
func NewHandlerWrapper(sw *go2sky.Tracer, options ...HandlerOption) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			co := &handlerWrapper{
				sw: sw,
			}
			for _, option := range options {
				option(co)
			}
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			span, ctx, err := sw.CreateEntrySpan(ctx, name, func() (string, error) {
				str, _ := metadata.Get(ctx, strings.Title(propagation.Header))
				return str, nil
			})
			if err != nil {
				return err
			}
			defer span.End()
			for k := range co.extraTags {
				v, ok := metadata.Get(ctx, k)
				if !ok {
					log.Fatalf("set tag %s failed \n", k)
				}
				span.Tag(go2sky.Tag(k), v)
			}
			if err = fn(ctx, req, rsp); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

