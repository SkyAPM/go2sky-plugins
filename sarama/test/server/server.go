//
// Copyright 2021 SkyAPM org
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
	"net/http"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky-plugins/sarama"
)

var kafka = []string{"kafka:9092"}

const (
	service = "kafka-reporter"
	addr    = ":8081"
)

func main() {
	// init tracer
	re, err := sarama.NewKafkaReporter(kafka)
	if err != nil {
		log.Fatalf("create kafka reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(service, go2sky.WithReporter(re), go2sky.WithInstance("provider1"))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	route := http.NewServeMux()
	route.HandleFunc("/healthCheck", func(res http.ResponseWriter, req *http.Request) {
		_, _ = res.Write([]byte("success"))
	})
	route.HandleFunc("/info", func(res http.ResponseWriter, req *http.Request) {
		span, _, err := tracer.CreateLocalSpan(
			context.Background(),
			go2sky.WithOperationName("info"),
		)
		if err != nil {
			log.Fatalf("create span error: %v \n", err)
		}
		defer span.End()

		_, _ = res.Write([]byte("info success"))
	})

	log.Println("start server")
	err = http.ListenAndServe(addr, route)
	if err != nil {
		log.Fatalf("server start error: %v \n", err)
	}
}
