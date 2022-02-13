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

package main

import (
	"context"
	"fmt"
	"os"

	"github.com/SkyAPM/go2sky"
	kratosplugin "github.com/SkyAPM/go2sky-plugins/kratos"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/go-kratos/kratos/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

const (
	oap         = "mockoap:19876"
	serviceName = "kratos-client"
	host        = "kratosserver"
)

func main() {
	r, err := reporter.NewGRPCReporter(oap)
	//r, err := reporter.NewLogReporter()
	if err != nil {
		panic(err)
	}
	defer r.Close()

	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "segment_id", kratosplugin.SegmentID())
	logger = log.With(logger, "trace_id", kratosplugin.TraceID())
	logger = log.With(logger, "span_id", kratosplugin.SpanID())

	tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(r))
	if err != nil {
		panic(err)
	}

	httpCli, err := http.NewClient(
		context.Background(),
		http.WithMiddleware(
			kratosplugin.Client(tracer),
			logging.Client(logger),
		),
		http.WithEndpoint(host+":8000"),
	)
	if err != nil {
		panic(err)
	}
	defer httpCli.Close()

	grpcCli, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithMiddleware(
			kratosplugin.Client(tracer),
			logging.Client(logger),
		),
		grpc.WithEndpoint(host+":9000"),
	)
	if err != nil {
		panic(err)
	}
	defer grpcCli.Close()

	httpClient := helloworld.NewGreeterHTTPClient(httpCli)
	grpcClient := helloworld.NewGreeterClient(grpcCli)

	server := http.NewServer(
		http.Address(":8080"),
		http.Middleware(
			recovery.Recovery(),
			kratosplugin.Server(tracer),
			logging.Server(logger),
		),
	)

	route := server.Route("/")
	route.GET("/hello", func(ctx http.Context) error {
		var in interface{}
		http.SetOperation(ctx, "/hello")
		h := ctx.Middleware(func(ctx context.Context, req interface{}) (interface{}, error) {
			hreply, err := httpClient.SayHello(ctx, &helloworld.HelloRequest{Name: "http-kratos"})
			if err != nil {
				return fmt.Sprintf("[http] error: %v", err), err
			}
			greply, err := grpcClient.SayHello(ctx, &helloworld.HelloRequest{Name: "grpc-kratos"})
			if err != nil {
				return fmt.Sprintf("[grpc] error: %v", err), err
			}
			return fmt.Sprintf("[http] Say hello: %s, [grpc] Say hello: %s", hreply, greply), nil
		})
		return ctx.Returns(h(ctx, &in))
	})

	app := kratos.New(
		kratos.Name(serviceName),
		kratos.Server(server),
	)
	if err := app.Run(); err != nil {
		panic(err)
	}
}
