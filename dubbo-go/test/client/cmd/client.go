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
	"time"

	_ "dubbo.apache.org/dubbo-go/v3/cluster/cluster_impl"
	_ "dubbo.apache.org/dubbo-go/v3/cluster/loadbalance"
	_ "dubbo.apache.org/dubbo-go/v3/common/proxy/proxy_factory"
	"dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/filter/filter_impl"
	_ "dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	_ "dubbo.apache.org/dubbo-go/v3/registry/protocol"
	_ "dubbo.apache.org/dubbo-go/v3/registry/zookeeper"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	hessian "github.com/apache/dubbo-go-hessian2"

	dubbogo "github.com/SkyAPM/go2sky-plugins/dubbo-go"
	"github.com/SkyAPM/go2sky-plugins/dubbo-go/test/client/pkg"
	httpplugin "github.com/SkyAPM/go2sky/plugins/http"
)

var userProvider = new(pkg.UserProvider)

func init() {
	config.SetConsumerService(userProvider)
	hessian.RegisterPOJO(&pkg.User{})
}

const (
	oap         = "mockoap:19876"
	serviceName = "dubbo-go-client"
)

// need to setup environment variable "CONF_CONSUMER_FILE_PATH" to "conf/client.yml" before run
func main() {
	hessian.RegisterPOJO(&pkg.User{})
	config.Load()
	time.Sleep(3 * time.Second)

	report, err := reporter.NewGRPCReporter(oap)
	if err != nil {
		log.Fatalf("crate grpc reporter error: %v \n", err)
	}

	tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(report))
	if err != nil {
		log.Fatalf("crate tracer error: %v \n", err)
	}

	err = dubbogo.SetClientTracer(tracer)
	if err != nil {
		log.Fatalf("set tracer error: %v \n", err)
	}
	dubbogo.SetClientExtraTags("extra-tags", "client")
	dubbogo.SetClientReportTags("release")
	dubbogo.SetClientReportTags("organization")

	route := http.NewServeMux()

	route.HandleFunc("/hello", func(writer http.ResponseWriter, req *http.Request) {
		user := &pkg.User{}
		err1 := userProvider.GetUser(req.Context(), []interface{}{"A001"}, user)
		if err1 != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			_, _ = writer.Write([]byte(fmt.Sprintf("call service err %v \n", err1)))
			return
		}

		_, _ = writer.Write([]byte(fmt.Sprintf("%v", *user)))
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
