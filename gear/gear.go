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
	"strconv"
	"time"

	"github.com/SkyAPM/go2sky"
	"github.com/teambition/gear"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const componentIDGearServer = 5007

//Middleware gear middleware return HandlerFunc  with tracing.
func Middleware(tracer *go2sky.Tracer) gear.Middleware {
	return func(ctx *gear.Context) error {
		if tracer == nil {
			return nil
		}

		span, _, err := tracer.CreateEntrySpan(ctx, operationName(ctx), func(key string) (string, error) {
			return ctx.GetHeader(key), nil
		})
		if err != nil {
			return nil
		}

		span.SetComponent(componentIDGearServer)
		span.Tag(go2sky.TagHTTPMethod, ctx.Method)
		span.Tag(go2sky.TagURL, ctx.Host+ctx.Path)
		span.SetSpanLayer(agentv3.SpanLayer_Http)

		ctx.OnEnd(func() {
			code := ctx.Res.Status()
			span.Tag(go2sky.TagStatusCode, strconv.Itoa(code))
			if code >= 400 {
				span.Error(time.Now(), string(ctx.Res.Body()))
			}
			span.End()
		})
		return nil
	}
}

func operationName(ctx *gear.Context) string {
	return fmt.Sprintf("/%s%s", ctx.Method, ctx.Path)
}
