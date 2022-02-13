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
	"fmt"
	"log"
	"net/http"

	"github.com/SkyAPM/go2sky"
	httpplugin "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/asim/go-micro/v3/client"

	microv3 "github.com/asim/go-micro/v3"

	microv3plugin "github.com/SkyAPM/go2sky-plugins/micro"
)

const (
	oap         = "mockoap:19876"
	serviceName = "micro-client"
)

func main() {
	report, err := reporter.NewGRPCReporter(oap)
	if err != nil {
		log.Fatalf("crate grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(report))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	cli := microv3.NewService(
		microv3.Name(serviceName),
		microv3.WrapClient(microv3plugin.NewClientWrapper(tracer, microv3plugin.WithClientWrapperReportTags("Micro-From-Service"))),
	)

	route := http.NewServeMux()

	c := cli.Client()
	route.HandleFunc("/hello", func(writer http.ResponseWriter, req *http.Request) {
		request := c.NewRequest("greeter", "Greeter.Hello", "john", client.WithContentType("application/json"))
		var response string
		if err1 := c.Call(req.Context(), request, &response); err1 != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(fmt.Sprintf("call service err %v \n", err1)))
			return
		}
		_, _ = writer.Write([]byte(response))
	})

	sm, err := httpplugin.NewServerMiddleware(tracer)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}
	err = http.ListenAndServe(":8080", sm(route))
	if err != nil {
		log.Fatal(err)
	}

}
