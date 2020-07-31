package main

import (
	"context"
	"fmt"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/micro/go-micro/v2/server"
	"github.com/micro/go-micro/v2"
	"log"
	"sw_micro"
	protos "sw_micro/proto"
	"github.com/SkyAPM/go2sky"
)

type Greeter struct {
}

func (g *Greeter) Hello(context context.Context, req *protos.Request, rsp *protos.Response) error {
	rsp.Greeting = "Hello " + req.Name
	return nil
}

func lowWrapper(fn server.HandlerFunc) server.HandlerFunc {
	return func(ctx context.Context, req server.Request, rsp interface{}) error {
		log.Printf("[wrapper] server request: %v", req.Endpoint())
		err := fn(ctx, req, rsp)
		return err
	}
}



func main() {
	r, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("Create report err: %v", err)
	}
	tracer, err := go2sky.NewTracer("example", go2sky.WithReporter(r))

	sv := micro.NewService(
		micro.Name("greeter"),
		micro.WrapHandler(lowWrapper),
		micro.WrapHandler(sw_micro.NewHnadlerWrapper(tracer)),
	)
	sv.Init()

	err = protos.RegisterGreeterHandler(sv.Server(), new(Greeter))
	if err != nil {
		fmt.Println(err)
	}

	if err := sv.Run(); err != nil {
		fmt.Println(err)
	}
}
