# Go2sky with go-kratos v2

## Installation
```go
go get -u github.com/SkyAPM/go2sky-plugins/kratos
```

## Usage

Server:
```go
    //Use gRPC reporter for production
	r, err := reporter.NewLogReporter()
	if err != nil {
		panic(err)
	}
	defer r.Close()

	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "segment_id", TraceSegmentID())
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
```

Client:
```go
    //Use gRPC reporter for production
	r, err := reporter.NewLogReporter()
	if err != nil {
		panic(err)
	}
	defer r.Close()

	logger := log.NewStdLogger(os.Stdout)
	logger = log.With(logger, "segment_id", TraceSegmentID())
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
		http.WithEndpoint("127.0.0.1:8000"),
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
		grpc.WithEndpoint("127.0.0.1:9000"),
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
```

