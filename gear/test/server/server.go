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
	gear_plugin "github.com/SkyAPM/go2sky-plugins/gear"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/teambition/gear"
)

const (
	oap     = "mockoap:19876"
	service = "gear"
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

	app := gear.New()
	app.Use(gear_plugin.Middleware(tracer))

	router := gear.NewRouter()
	router.Get("/hello", func(ctx *gear.Context) error {
		return ctx.End(http.StatusOK, []byte("Hello World!"))
	})

	app.UseHandler(router)

	app.Error(app.Listen(":8080"))
}
