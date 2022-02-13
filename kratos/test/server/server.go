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
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/metadata"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

const (
	oap         = "mockoap:19876"
	serviceName = "kratos-server"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	helloworld.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: fmt.Sprintf("Hello %s", in.Name)}, nil
}

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
	httpSvr := http.NewServer(
		http.Address(":8000"),
		http.Middleware(
			recovery.Recovery(),
			metadata.Server(metadata.WithPropagatedPrefix("")),
			kratosplugin.Server(tracer, kratosplugin.WithReportTags("User-Agent")),
			logging.Server(logger),
		),
	)

	grpcSvr := grpc.NewServer(
		grpc.Address(":9000"),
		grpc.Middleware(
			middleware.Chain(
				recovery.Recovery(),
				kratosplugin.Server(tracer),
				logging.Server(logger),
			),
		))

	s := &server{}
	helloworld.RegisterGreeterServer(grpcSvr, s)
	helloworld.RegisterGreeterHTTPServer(httpSvr, s)

	app := kratos.New(
		kratos.Name(serviceName),
		kratos.Server(
			httpSvr,
			grpcSvr,
		),
	)

	if err := app.Run(); err != nil {
		panic(err)
	}
}
