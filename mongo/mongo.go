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

package mongo

import (
	"context"
	"sync"

	"github.com/SkyAPM/go2sky"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/event"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
)

const (
	// ComponentMongo ComponentID.
	ComponentMongo int32 = 42

	// ComponentMongoDB db.type.
	ComponentMongoDB string = "MongoDB"
)

// Option custom option.
type Option func(span go2sky.Span, evt *event.CommandStartedEvent)

// Middleware mongo monitor.
func Middleware(tracer *go2sky.Tracer, peer string, opts ...Option) *event.CommandMonitor {
	spanMap := sync.Map{}
	apmMonitor := &event.CommandMonitor{
		Started: func(ctx context.Context, evt *event.CommandStartedEvent) {
			span, _, err := tracer.CreateLocalSpan(ctx,
				go2sky.WithSpanType(go2sky.SpanTypeExit),
				go2sky.WithOperationName(GetOpName(evt.CommandName)),
			)
			if err != nil {
				return
			}
			span.SetPeer(peer)
			span.SetComponent(ComponentMongo)
			span.SetSpanLayer(agentv3.SpanLayer_Database)
			span.Tag(go2sky.TagDBType, ComponentMongoDB)
			span.Tag(go2sky.TagDBStatement, GetMongoDBStatement(evt))
			for _, opt := range opts {
				opt(span, evt)
			}
			spanMap.Store(evt.RequestID, span)
		},
		Succeeded: func(ctx context.Context, evt *event.CommandSucceededEvent) {
			if span, ok := spanMap.Load(evt.RequestID); ok {
				span.(go2sky.Span).End()
				spanMap.Delete(evt.RequestID)
			}
		},
		Failed: func(ctx context.Context, evt *event.CommandFailedEvent) {
			if span, ok := spanMap.Load(evt.RequestID); ok {
				span.(go2sky.Span).End()
				spanMap.Delete(evt.RequestID)
			}
		},
	}
	return apmMonitor
}

// GetOpName get operation name.
func GetOpName(operation string) string {
	return "MongoDB/Go2Sky/" + operation
}

// GetMongoDBStatement get statement.
func GetMongoDBStatement(evt *event.CommandStartedEvent) string {
	removeFields := map[string]*struct{}{
		"lsid":         nil,
		"$clusterTime": nil,
		"txnNumber":    nil,
	}
	rows := make(bson.RawElement, 0)
	elements, err := evt.Command.Elements()
	if err != nil {
		return ""
	}
	for _, element := range elements {
		if _, ok := removeFields[element.Key()]; !ok {
			rows = append(rows, element...)
		}
	}
	return rows.String()
}
