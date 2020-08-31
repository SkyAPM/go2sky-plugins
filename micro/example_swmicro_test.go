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

// Package micro (sw_micro) is a plugin that can be used to trace Go-micro framework.

package swmicro

import (
	"context"
	"log"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/micro/go-micro"
	"github.com/micro/go-micro/client"
)

type Greeter struct{}

func (g *Greeter) Hello(ctx context.Context, name *string, msg *string) error {
	*msg = "Hello " + *name
	return nil
}

func ExampleNewHandlerWrapper() {
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer re.Close()

	tracer, err := go2sky.NewTracer("micor-server", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	service := micro.NewService(
		micro.Name("greeter"),
		micro.WrapHandler(NewHandlerWrapper(tracer)),
	)
	service.Init()

	if err := service.Run(); err != nil {
		log.Fatalln(err)
	}
	log.Fatalln("HandleWrapper test pass")

}

func ExampleNewClientWrapper() {
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer re.Close()

	tracer, err := go2sky.NewTracer("micor-client", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	service := micro.NewService(
		micro.WrapClient(NewClientWrapper(tracer)),
	)
	service.Init()
	c := service.Client()

	request := c.NewRequest("greeter", "Greeter.Hello", "john", client.WithContentType("application/json"))
	var response string
	if err := c.Call(context.TODO(), request, &response); err != nil {
		log.Fatalln(err)
	}
	log.Fatalln(response + "\n ClientWrapper test pass")
}
