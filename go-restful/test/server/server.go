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
	"io"
	"log"
	"net/http"

	"github.com/SkyAPM/go2sky"
	go_restfile_plugin "github.com/SkyAPM/go2sky-plugins/go-restful"
	"github.com/SkyAPM/go2sky/reporter"

	"github.com/emicklei/go-restful/v3"
)

const (
	oap     = "mockoap:19876"
	service = "go-restful"
)

func main() {
	report, err := reporter.NewGRPCReporter(oap)
	if err != nil {
		log.Fatalf("crate grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(service, go2sky.WithReporter(report))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	ws := &restful.WebService{}
	ws.Filter(go_restfile_plugin.NewTraceFilterFunction(tracer))

	ws.Route(ws.GET("/hello").To(func(request *restful.Request, response *restful.Response) {
		_, _ = io.WriteString(response, "Hello World!")
	}))

	restful.Add(ws)

	if err = http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
