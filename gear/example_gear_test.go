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

package gear

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/SkyAPM/go2sky"
	h "github.com/SkyAPM/go2sky/plugins/http"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/teambition/gear"
)

func ExampleMiddleware() {
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}

	defer re.Close()

	tracer, err := go2sky.NewTracer("gear", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	app := gear.New()
	app.Use(Middleware(tracer))

	router := gear.NewRouter()
	router.Get("/user", func(ctx *gear.Context) error {
		name := ctx.Param("name")
		return ctx.End(http.StatusOK, []byte(fmt.Sprintf("Hello %s", name)))
	})

	app.UseHandler(router)

	go func() {
		app.Error(app.Listen(":8080"))
	}()
	// Wait for the server to start
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

	request, err := http.NewRequest("GET", "http://127.0.0.1:8080/user?name=gear", nil)
	if err != nil {
		log.Fatalf("unable to create http request: %+v\n", err)
	}

	res, err := client.Do(request)
	if err != nil {
		log.Fatalf("unable to do http request: %+v\n", err)
	}

	_ = res.Body.Close()
}
