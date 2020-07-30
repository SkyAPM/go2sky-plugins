package skywalking

import (
	"context"
	"fmt"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"github.com/micro/go-micro/v2/client"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2/metadata"
	"time"
)

type swWrapper struct {
	sw *go2sky.Tracer
	client.Client
}

func NewHnadlerWrapper(sw *go2sky.Tracer) server.HandlerWrapper {
	return func(fn server.HandlerFunc) server.HandlerFunc {
		return func(ctx context.Context, req server.Request, rsp interface{}) error {
			name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
			span, ctx, err := sw.CreateEntrySpan(ctx, name, func() (string, error) {
				return getHeader(req.Header(),propagation.Header), nil
			})
			if err != nil {
				return err
			}
			span.Tag(go2sky.TagHTTPMethod, req.Method())
			if err = fn(ctx, req, rsp); err != nil {
				span.Error(time.Now(), err.Error())
			}
			return err
		}
	}
}

func getHeader(head map[string]string, str string) string {
	return head[str]
}

func setHeader(m map[string]string, header string) {
	m[propagation.Header] = header
}

func (s *swWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts ...client.CallOption) error {
	name := fmt.Sprintf("%s.%s", req.Service(), req.Endpoint())
	span, err:= s.sw.CreateExitSpan(ctx,name, name, func(header string) error {
		return nil    // TODO
	})
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

func NewClientWrapper (sw *go2sky.Tracer) client.Wrapper {
	return func(c client.Client) client.Client {
		return &swWrapper{sw: sw, Client: c}
	}
}
