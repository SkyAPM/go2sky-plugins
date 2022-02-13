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

package micro

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	microv3 "github.com/asim/go-micro/v3"
	"github.com/asim/go-micro/v3/client"
	"github.com/asim/go-micro/v3/logger"
)

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, name *string, msg *string) error {
	*msg = "Hello " + *name
	return nil
}

func ExampleNewHandlerWrapper() {
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
		_ = logger.DefaultLogger.Init(logger.WithLevel(logger.ErrorLevel))
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
	// Output:
}
