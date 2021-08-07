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

package main

import (
	"context"
	"fmt"
	stdhttp "net/http"
	"os"

	"github.com/SkyAPM/go2sky"
	kratosplugin "github.com/SkyAPM/go2sky-plugins/kratos"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/go-kratos/kratos/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/logging"
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

	route := stdhttp.NewServeMux()
	route.HandleFunc("/hello", func(writer stdhttp.ResponseWriter, request *stdhttp.Request) {
		reply, err := httpClient.SayHello(request.Context(), &helloworld.HelloRequest{Name: "http-kratos"})
		if err != nil {
			_, _ = writer.Write([]byte(fmt.Sprintf("[http] error: %v", err)))
			return
		}
		_, _ = writer.Write([]byte(fmt.Sprintf("[http] Say hello: %s\n", reply)))

		reply, err = grpcClient.SayHello(request.Context(), &helloworld.HelloRequest{Name: "grpc-kratos"})
		if err != nil {
			_, _ = writer.Write([]byte(fmt.Sprintf("[grpc] error: %v", err)))
			return
		}
		_, _ = writer.Write([]byte(fmt.Sprintf("[grpc] Say hello: %s\n", reply)))
	})
	route.HandleFunc("/healthCheck", func(writer stdhttp.ResponseWriter, request *stdhttp.Request) {
		_, _ = writer.Write([]byte("Success"))
	})

	err = stdhttp.ListenAndServe(":8080", route)
	if err != nil {
		panic(err)
	}
}
