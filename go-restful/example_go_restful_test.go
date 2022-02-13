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

package restful

import (
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/SkyAPM/go2sky"
	h "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/emicklei/go-restful/v3"
)

func ExampleFilterFunction() {
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}

	defer re.Close()

	tracer, err := go2sky.NewTracer("go-restful", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	ws := new(restful.WebService)
	ws.Filter(NewTraceFilterFunction(tracer))

	ws.Route(ws.GET("/hello").To(func(req *restful.Request, resp *restful.Response) {
		_, _ = io.WriteString(resp, "go-restful")
	}))
	restful.Add(ws)

	go func() {
		_ = http.ListenAndServe(":8080", nil)
	}()

	// // Wait for the server to start
	time.Sleep(time.Second)

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		request(tracer)
	}()
	wg.Wait()
	// Output:
}

func request(tracer *go2sky.Tracer) {
	client, err := h.NewClient(tracer)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}

	request, err := http.NewRequest("GET", "http://127.0.0.1:8080/hello", nil)
	if err != nil {
		log.Fatalf("unable to create http request: %+v\n", err)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Fatalf("unable to do http request: %+v\n", err)
	}

	_ = res.Body.Close()
}
