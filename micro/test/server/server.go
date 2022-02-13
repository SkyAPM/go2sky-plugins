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

package main

import (
	"context"
	"log"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	microv3 "github.com/asim/go-micro/v3"

	microv3plugin "github.com/SkyAPM/go2sky-plugins/micro"
)

const (
	oap         = "mockoap:19876"
	serviceName = "micro-server"
)

// Greeter example
type Greeter struct{}

// Hello example
func (g *Greeter) Hello(ctx context.Context, name *string, msg *string) error {
	*msg = "Hello " + *name
	return nil
}

func main() {
	report, err := reporter.NewGRPCReporter(oap)
	if err != nil {
		log.Fatalf("crate grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(report))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	service := microv3.NewService(
		microv3.Name("greeter"),
		microv3.WrapHandler(microv3plugin.NewHandlerWrapper(tracer, "User-Agent")),
		microv3.Address(":8081"))

	if err = microv3.RegisterHandler(service.Server(), new(Greeter)); err != nil {
		log.Fatalf("Register service error: %v \n", err)
	}

	if err = service.Run(); err != nil {
		log.Fatalf("Run server error: %v \n", err)
	}
}
