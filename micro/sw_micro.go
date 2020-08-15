package sw_micro

import (
	"context"
	"errors"
	"fmt"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/metadata"
	"github.com/micro/go-micro/v2/registry"
	"time"
)

type swWrapper struct {
	sw go2sky.Tracer
	client.Client
}

func (s *swWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	span, err:= s.sw.CreateExitSpan(ctx,name, req.Service(), func(header string) error {
		swHeader := make(metadata.Metadata)
		swHeader[propagation.Header] = header
		ctx = metadata.NewContext(ctx, swHeader)
		return nil
	})
	if err != nil {
		return err
	}
	defer span.End()
	span.Tag(go2sky.TagHTTPMethod, req.Method())
	span.Tag(go2sky.TagURL, req.Service()+req.Endpoint())
	if err= s.Client.Call(ctx, req, rsp, opts ...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}
/*
func (s *swWrapper) Stream(ctx context.Context, req client.Request, opts ...client.CallOption) (client.Stream, error) {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	span, err := s.sw.CreateExitSpan(ctx, name, req.Service(), func(header string) error {
		swHeader := make(metadata.Metadata)
		swHeader[propagation.Header] = header
		ctx = metadata.NewContext(ctx, swHeader)
		return nil
	})
	defer span.End()
	span.Tag(go2sky.TagHTTPMethod, req.Method())
	span.Tag(go2sky.TagURL, req.Service()+req.Endpoint())
	stream, err := s.Client.Stream(ctx, req, opts...)
	if err != nil {
		span.Error(time.Now(), err.Error())
	}
	return stream, err
}

func (s *swWrapper) Publish(ctx context.Context, p client.Message, opts ...client.PublishOption) error {
	name := fmt.Sprintf("Pub to %s", p.Topic())
	span, err := s.sw.CreateExitSpan(ctx, name, p.ContentType(), func(header string) error {
		swHeader := make(metadata.Metadata)
		swHeader[propagation.Header] = header
		ctx = metadata.NewContext(ctx, swHeader)
		return nil
	})
	span.Tag(go2sky.TagHTTPMethod, p.ContentType())
	span.Tag(go2sky.TagURL, p.Topic())
	if err != nil {
		return err
	}
	defer span.End()
	if err = s.Client.Publish(ctx, p, opts...); err != nil {
		span.Error(time.Now(), err.Error())
	}
	return err
}*/

func NewClientWrapper (sw *go2sky.Tracer) client.Wrapper {
	return func(c client.Client) client.Client {
		return &swWrapper{sw: *sw, Client: c}
	}
}

func NewCallWrapper(sw *go2sky.Tracer) client.CallWrapper {
	return func(cf client.CallFunc) client.CallFunc {
		return func(ctx context.Context, node *registry.Node, req client.Request, rsp interface{}, opts client.CallOptions) error {
			if sw == nil {
				return errors.New("tracer is nil")
			}
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			span, err := sw.CreateExitSpan(ctx, name, req.Service(), func(header string) error {
				swHeader := make(metadata.Metadata)
				swHeader[propagation.Header] = header
				metadata.MergeContext(ctx, swHeader, true)
				return nil
			})
			if err != nil {
				return err
			}
			span.Tag(go2sky.TagHTTPMethod, req.Method())
			span.Tag(go2sky.TagURL, req.Service()+req.Endpoint())
			defer span.End()
			if err = cf(ctx, node, req, rsp, opts); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}
/*
func NewSubscriberWrapper(sw *go2sky.Tracer) server.SubscriberWrapper {
	return func(next server.SubscriberFunc) server.SubscriberFunc {
		return func(ctx context.Context, msg server.Message) error {
			name := "Sub from " + msg.Topic()
			if sw == nil {
				return errors.New("tracer is nil")
			}
			span, err := sw.CreateExitSpan(ctx, name, msg.ContentType(), func(header string) error {
				swHeader := make(metadata.Metadata)
				swHeader[propagation.Header] = header
				ctx = metadata.NewContext(ctx, swHeader)
				return nil
			})
			if err != nil {
				return err
			}
			defer span.End()
			span.Tag(go2sky.TagHTTPMethod, msg.ContentType())
			span.Tag(go2sky.TagURL, msg.Topic())
			if err = next(ctx, msg); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

func NewHnadlerWrapper(sw *go2sky.Tracer) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			span, ctx, err := sw.CreateEntrySpan(ctx, name, func() (string, error) {
				str, ok := metadata.Get(ctx, propagation.Header)
				if !ok {
					return "no key", nil
				}
				return str, nil
			})
			if err != nil {
				return err
			}
			span.Tag(go2sky.TagHTTPMethod, req.Method())
			span.Tag(go2sky.TagURL, req.Service()+req.Endpoint())
			if err = fn(ctx, req, rsp); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}*/