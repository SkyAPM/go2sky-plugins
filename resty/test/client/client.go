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
	"log"
	"net/http"

	"github.com/SkyAPM/go2sky"
	restyplugin "github.com/SkyAPM/go2sky-plugins/resty"
	httpPlugin "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
)

const (
	oap         = "mockoap:19876"
	service     = "go-resty"
	upstreamURL = "http://httpserver:8080/helloserver"
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

	client := restyplugin.NewGoResty(tracer)

	route := http.NewServeMux()
	route.HandleFunc("/hello", func(writer http.ResponseWriter, request *http.Request) {
		resp, err1 := client.R().SetContext(request.Context()).Get(upstreamURL)
		if err1 != nil {
			log.Printf("unable to do http request: %+v\n", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		writer.WriteHeader(resp.StatusCode())
		_, _ = writer.Write(resp.Body())
	})

	sm, err := httpPlugin.NewServerMiddleware(tracer)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}
	err = http.ListenAndServe(":8080", sm(route))
	if err != nil {
		log.Fatal(err)
	}
}
