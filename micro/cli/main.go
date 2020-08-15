package main

import (
	"context"
	"fmt"
	"log"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/micro/go-micro"
	"sw_micro"
	protos "sw_micro/proto"
)

/*type logWrapper struct {
	client.Client
}

func (l *logWrapper) Call(ctx context.Context, req client.Request, rsp interface{}, opts...client.CallOption) error {
	md, _ := metadata.FromContext(ctx)
	hdr := make(metadata.Metadata)
	hdr = metadata.Copy(md)
	hdr["Test"] = "test"
	ctx = metadata.NewContext(ctx, hdr)
	fmt.Printf("[log Wrapper] ctx: %v serviece: %s method: %s \n string: %s", md, req.Service(), req.Endpoint(), req.Endpoint())
	return l.Client.Call(ctx, req, rsp)
}

func NewLogWrapper(c client.Client) client.Client{
	return &logWrapper{c}
}*/





func main() {
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer re.Close()

	tracer, err := go2sky.NewTracer("micro-Client", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	service := micro.NewService(
		micro.Name("greeter.client"),
/*		micro.WrapClient(NewLogWrapper),*/
		micro.WrapClient(sw_micro.NewClientWrapper(tracer)),
		)
	service.Init()

	greeter := protos.NewGreeterService("greeter", service.Client())
	rsp, err := greeter.Hello(context.TODO(), &protos.Request{Name: "Zaun pianist"})
	if err != nil {
		fmt.Println(err)
	}
	str := rsp.Greeting
	fmt.Println(str)
}

