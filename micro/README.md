# Go2sky with go-micro (v1.18)

## Applicable version
<= v1.18

go-micro v2.0 had changed a lot and is not compatible with go2sky, so only 1.x version is supported.

## Installation
```go
go get -u github.com/micro/go-micro@v1.18
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
	sw "github.com/SkyAPM/go2sky-plugins/go_micro"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
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

	//create test server
	service := micro.NewService(
		micro.Name("greeter"),
		micro.WrapHandler(sw.NewHandlerWrapper(tracer)),
	)
	go func() {
		// initialise command line
		service.Init()
		// set the handler
		micro.RegisterHandler(service.Server(), new(Greeter))

		// run service
		service.Run()
	}()
	time.Sleep(time.Second * 5)

	cli := micro.NewService(
		micro.Name("micro_client"),
		micro.WrapClient(sw.NewClientWrapper(tracer)),
	)
	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		cli.Init()
		c := cli.Client()

		request := c.NewRequest("greeter", "Greeter.Hello", "john", client.WithContentType("application/json"))
		var response string
		if err := c.Call(context.TODO(), request, &response); err != nil {
			log.Fatalf("call service err %v \n", err)
		}
		log.Fatalf("reseponse: %v \n", response)
		wg.Done()
	}()
	wg.Wait()
}

```