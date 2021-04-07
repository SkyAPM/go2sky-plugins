# Go2sky with go-micro (v3.5.0)

## Installation
```go
go get -u github.com/SkyAPM/go2sky-plugins/micro
```

## Usage
```go
package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	microv3 "github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/client"
)

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, name *string, msg *string) error {
	*msg = "Hello " + *name
	return nil
}

func main() {
	//Use gRPC reporter for production
	r, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer r.Close()

	tracer, err := go2sky.NewTracer("example", go2sky.WithReporter(r))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	go func() {
		//create test server
		service := microv3.NewService(
			microv3.Name("greeter"),
			//Use go2sky middleware with tracing
			microv3.WrapHandler(NewHandlerWrapper(tracer, "User-Agent")),
		)
		// initialise command line
		// set the handler
		if err := microv3.RegisterHandler(service.Server(), new(Greeter)); err != nil {
			log.Fatalf("Registe service error: %v \n", err)
		}

		// run service
		if err := service.Run(); err != nil {
			log.Fatalf("Run server error: %v \n", err)
		}
	}()
	// wait server to start
	time.Sleep(time.Second * 5)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		cli := microv3.NewService(
			microv3.Name("micro_client"),
			//Use go2sky middleware with tracing
			microv3.WrapClient(NewClientWrapper(tracer, WithClientWrapperReportTags("Micro-From-Service"))),
		)
		c := cli.Client()
		request := c.NewRequest("greeter", "Greeter.Hello", "john", client.WithContentType("application/json"))
		var response string
		if err := c.Call(context.TODO(), request, &response); err != nil {
			log.Fatalf("call service err %v \n", err)
		}
		log.Printf("reseponse: %v \n", response)
	}()
	wg.Wait()
}
```

[See more](example_micro_handler_test.go).