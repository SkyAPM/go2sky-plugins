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

package example

import (
	"context"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"google.golang.org/grpc"

	grpc2 "github.com/aoliang/go2sky-plugins/grpc"
	"github.com/aoliang/go2sky-plugins/grpc/example/demo"
)

type demoServer struct{}

func (s *demoServer) SayHello(ctx context.Context, in *demo.HelloRequest) (*demo.HelloReply, error) {
	log.Printf("Received: %v", in.Name)
	time.Sleep(time.Duration(200+rand.Intn(200)) * time.Millisecond)
	return &demo.HelloReply{Message: "Hello " + in.Name}, nil
}

func (s *demoServer) mustEmbedUnimplementedGreeterServer() {}

func ExampleInterceptorFunction() {
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}

	defer re.Close()

	tracer, err := go2sky.NewTracer("grpc", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	go func() {
		lis, err := net.Listen("tcp", "0.0.0.0:18088")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		traceInterceptor := grpc2.NewUnaryServerTraceInterceptor(tracer)
		s := grpc.NewServer(grpc.UnaryInterceptor(traceInterceptor))
		demo.RegisterGreeterServer(s, &demoServer{})
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// // Wait for the server to start
	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		request(tracer)
	}()
	wg.Wait()
	// Output:
}

func request(tracer *go2sky.Tracer) {
	traceInterceptor := grpc2.NewUnaryClientTraceInterceptor(tracer)
	conn, err := grpc.Dial("localhost:18088", grpc.WithInsecure(), grpc.WithUnaryInterceptor(traceInterceptor))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()

	client := demo.NewGreeterClient(conn)
	reply, err := client.SayHello(context.Background(), &demo.HelloRequest{Name: "aoliang"})
	if err != nil {
		log.Fatalf("error happened. %v", err)
	}
	log.Println(reply.Message)

}
