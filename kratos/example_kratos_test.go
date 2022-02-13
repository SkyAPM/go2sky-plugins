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

package kratos

import (
	"context"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/go-kratos/kratos/examples/helloworld/helloworld"
	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware"
	"github.com/go-kratos/kratos/v2/middleware/logging"
	"github.com/go-kratos/kratos/v2/middleware/recovery"
	"github.com/go-kratos/kratos/v2/transport/grpc"
	"github.com/go-kratos/kratos/v2/transport/http"
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	helloworld.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *helloworld.HelloRequest) (*helloworld.HelloReply, error) {
	return &helloworld.HelloReply{Message: fmt.Sprintf("Hello %s", in.Name)}, nil
}

func Test(t *testing.T) {
	//Use gRPC reporter for production
	r, err := reporter.NewLogReporter()
	if err != nil {
		panic(err)
	}
	defer r.Close()

	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "segment_id", SegmentID())
	logger = log.With(logger, "trace_id", TraceID())
	logger = log.With(logger, "span_id", SpanID())

	// run server
	go func() {
		serverName := "kratos-server"
		tracer, err := go2sky.NewTracer(serverName, go2sky.WithReporter(r))
		if err != nil {
			panic(err)
		}
		httpSvr := http.NewServer(
			http.Address(":8000"),
			http.Middleware(
				recovery.Recovery(),
				Server(tracer),
				logging.Server(logger),
			),
		)

		grpcSvr := grpc.NewServer(
			grpc.Address(":9000"),
			grpc.Middleware(
				middleware.Chain(
					recovery.Recovery(),
					Server(tracer),
					logging.Server(logger),
				),
			))

		s := &server{}
		helloworld.RegisterGreeterServer(grpcSvr, s)
		helloworld.RegisterGreeterHTTPServer(httpSvr, s)

		app := kratos.New(
			kratos.Name(serverName),
			kratos.Server(
				httpSvr,
				grpcSvr,
			),
		)

		if err := app.Run(); err != nil {
			panic(err)
		}
	}()

	// run client
	time.Sleep(5 * time.Second)
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		clientName := "kratos-client"
		defer wg.Done()
		tracer, err := go2sky.NewTracer(clientName, go2sky.WithReporter(r))
		if err != nil {
			panic(err)
		}

		httpCli, err := http.NewClient(
			context.Background(),
			http.WithMiddleware(
				Client(tracer),
				logging.Client(logger),
			),
			http.WithEndpoint("127.0.0.1:8000"),
		)
		if err != nil {
			panic(err)
		}

		grpcCli, err := grpc.DialInsecure(
			context.Background(),
			grpc.WithMiddleware(
				Client(tracer),
				logging.Client(logger),
			),
			grpc.WithEndpoint("127.0.0.1:9000"),
		)
		if err != nil {
			panic(err)
		}

		httpClient := helloworld.NewGreeterHTTPClient(httpCli)
		grpcClient := helloworld.NewGreeterClient(grpcCli)

		reply, err := httpClient.SayHello(context.Background(), &helloworld.HelloRequest{Name: "http-kratos"})
		if err != nil {
			panic(err)
		}
		t.Logf("[http] Say hello: %s\n", reply)

		reply, err = grpcClient.SayHello(context.Background(), &helloworld.HelloRequest{Name: "grpc-kratos"})
		if err != nil {
			panic(err)
		}
		t.Logf("[grpc] Say hello: %s\n", reply)
	}()
	wg.Wait()
}

func ExampleServer() {
	//Use gRPC reporter for production
	r, err := reporter.NewLogReporter()
	if err != nil {
		panic(err)
	}
	defer r.Close()

	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "segment_id", SegmentID())
	logger = log.With(logger, "trace_id", TraceID())
	logger = log.With(logger, "span_id", SpanID())

	serverName := "kratos-server"
	tracer, err := go2sky.NewTracer(serverName, go2sky.WithReporter(r))
	if err != nil {
		panic(err)
	}
	httpSvr := http.NewServer(
		http.Address(":8000"),
		http.Middleware(
			recovery.Recovery(),
			Server(tracer),
			logging.Server(logger),
		),
	)

	grpcSvr := grpc.NewServer(
		grpc.Address(":9000"),
		grpc.Middleware(
			middleware.Chain(
				recovery.Recovery(),
				Server(tracer),
				logging.Server(logger),
			),
		))

	s := &server{}
	helloworld.RegisterGreeterServer(grpcSvr, s)
	helloworld.RegisterGreeterHTTPServer(httpSvr, s)

	app := kratos.New(
		kratos.Name(serverName),
		kratos.Server(
			httpSvr,
			grpcSvr,
		),
	)

	if err := app.Run(); err != nil {
		panic(err)
	}
}

func ExampleClient() {
	//Use gRPC reporter for production
	r, err := reporter.NewLogReporter()
	if err != nil {
		panic(err)
	}
	defer r.Close()

	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "segment_id", SegmentID())
	logger = log.With(logger, "trace_id", TraceID())
	logger = log.With(logger, "span_id", SpanID())

	clientName := "kratos-client"
	tracer, err := go2sky.NewTracer(clientName, go2sky.WithReporter(r))
	if err != nil {
		panic(err)
	}

	httpCli, err := http.NewClient(
		context.Background(),
		http.WithMiddleware(
			Client(tracer),
			logging.Client(logger),
		),
		http.WithEndpoint("localhost:8000"),
	)
	if err != nil {
		panic(err)
	}
	defer httpCli.Close()

	grpcCli, err := grpc.DialInsecure(
		context.Background(),
		grpc.WithMiddleware(
			Client(tracer),
			logging.Client(logger),
		),
		grpc.WithEndpoint("localhost:9000"),
	)
	if err != nil {
		panic(err)
	}
	defer grpcCli.Close()

	httpClient := helloworld.NewGreeterHTTPClient(httpCli)
	grpcClient := helloworld.NewGreeterClient(grpcCli)

	reply, err := httpClient.SayHello(context.Background(), &helloworld.HelloRequest{Name: "http-kratos"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("[http] Say hello: %s\n", reply)

	reply, err = grpcClient.SayHello(context.Background(), &helloworld.HelloRequest{Name: "grpc-kratos"})
	if err != nil {
		panic(err)
	}
	fmt.Printf("[grpc] Say hello: %s\n", reply)
}
