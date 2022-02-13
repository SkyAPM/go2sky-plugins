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
	"fmt"
	"strconv"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/emicklei/go-restful/v3"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const componentIDGOHttpServer = 5004

// NewTraceFilterFunction return go-restful FilterFunction with tracing.
func NewTraceFilterFunction(tracer *go2sky.Tracer) restful.FilterFunction {
	if tracer == nil {
		return func(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
			chain.ProcessFilter(request, response)
		}
	}

	return func(request *restful.Request, response *restful.Response, chain *restful.FilterChain) {
		span, ctx, err := tracer.CreateEntrySpan(request.Request.Context(),
			fmt.Sprintf("/%s%s", request.Request.Method, request.SelectedRoutePath()), func(key string) (string, error) {
				return request.HeaderParameter(key), nil
			})

		if err != nil {
			chain.ProcessFilter(request, response)
			return
		}

		span.SetComponent(componentIDGOHttpServer)
		span.Tag(go2sky.TagHTTPMethod, request.Request.Method)
		span.Tag(go2sky.TagURL, request.Request.Host+request.Request.URL.Path)
		span.SetSpanLayer(agentv3.SpanLayer_Http)
		request.Request = request.Request.WithContext(ctx)
		defer func() {
			code := response.StatusCode()
			if code >= 400 {
				span.Error(time.Now(), "Error on handling request")
			}
			span.Tag(go2sky.TagStatusCode, strconv.Itoa(code))
			span.End()
		}()
		chain.ProcessFilter(request, response)
	}
}
