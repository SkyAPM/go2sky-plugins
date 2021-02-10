# Go2sky with [grpc(unary)](https://grpc.io)

## Note

This plugin works just for gRPC server and client in unary mode, streaming has not been supported.

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/grpc
```

## Usage
```go

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
```
[See more](example/example_grpc_test.go).