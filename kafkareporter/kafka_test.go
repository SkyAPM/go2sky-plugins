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

package kafkareporter

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/Shopify/sarama"
	"github.com/Shopify/sarama/mocks"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/propagation"
	"google.golang.org/protobuf/proto"
	commonv3 "skywalking.apache.org/repo/goapi/collect/common/v3"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	managementv3 "skywalking.apache.org/repo/goapi/collect/management/v3"
)

const (
	sample                = 1
	traceID               = "1f2d4bf47bf711eab794acde48001122"
	parentSegmentID       = "1e7c204a7bf711eab858acde48001122"
	parentSpanID          = 0
	parentService         = "service"
	parentServiceInstance = "serviceInstance"
	parentEndpoint        = "/foo/bar"
	addressUsedAtClient   = "foo.svc:8787"

	mockService         = "service"
	mockServiceInstance = "serviceInstance"
)

var header string

func init() {
	scx := propagation.SpanContext{
		Sample:                sample,
		TraceID:               traceID,
		ParentSegmentID:       parentSegmentID,
		ParentSpanID:          parentSpanID,
		ParentService:         parentService,
		ParentServiceInstance: parentServiceInstance,
		ParentEndpoint:        parentEndpoint,
		AddressUsedAtClient:   addressUsedAtClient,
	}
	header = scx.EncodeSW8()
}

func TestKafkaReporterE2E(t *testing.T) {
	r := createKafkaReporter()
	tracer, err := go2sky.NewTracer(mockService, go2sky.WithReporter(r), go2sky.WithInstance(mockServiceInstance))
	if err != nil {
		t.Error(err)
	}

	c := mocks.NewTestConfig()
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = false
	mp := mocks.NewAsyncProducer(t, c)
	mp.ExpectInputAndSucceed()
	r.producer = mp

	entrySpan, ctx, err := tracer.CreateEntrySpan(context.Background(), "/rest/api", func(key string) (string, error) {
		return header, nil
	})
	if err != nil {
		t.Error(err)
	}
	exitSpan, err := tracer.CreateExitSpan(ctx, "/foo/bar", "foo.svc:8787", func(key, value string) error {
		scx := propagation.SpanContext{}
		if key == propagation.Header {
			err = scx.DecodeSW8(value)
			if err != nil {
				t.Fatal(err)
			}
		}
		return nil
	})
	if err != nil {
		t.Error(err)
	}
	exitSpan.End()
	entrySpan.End()
	for msg := range r.producer.Successes() {
		r.Close()
		if msg.Topic != r.topicSegment {
			t.Errorf("Excepted kafka topic is %s not %s", r.topicSegment, msg.Topic)
		}
		v, _ := msg.Value.Encode()
		var s agentv3.SegmentObject
		if err := proto.Unmarshal(v, &s); err != nil {
			t.Fatal(err)
		}
		if s.TraceId != traceID {
			t.Errorf("trace id parse error")
		}
		if len(s.Spans) == 0 {
			t.Error("empty spans")
		}
		if s.Service != mockService {
			t.Error("error are not set service")
		}
		if s.ServiceInstance != mockServiceInstance {
			t.Error("error are not set service instance")
		}
	}
}

func TestKafkaReporter_Close(t *testing.T) {
	r := createKafkaReporter()
	tracer, err := go2sky.NewTracer(mockService, go2sky.WithReporter(r), go2sky.WithInstance(mockServiceInstance))
	if err != nil {
		t.Error(err)
	}
	c := mocks.NewTestConfig()
	c.Producer.Return.Errors = false
	mp := mocks.NewAsyncProducer(t, c)
	r.producer = mp

	entry, _, err := tracer.CreateEntrySpan(context.Background(), "/close", func(key string) (s string, err error) {
		return header, nil
	})
	if err != nil {
		t.Error(err)
	}
	r.Close()
	entry.End()
}

func TestKafkaReporterOption(t *testing.T) {
	// props
	instanceProps := make(map[string]string)
	instanceProps["org"] = "SkyAPM"

	// log
	logger := log.New(os.Stderr, "WithLogger", log.LstdFlags)

	// kafka config
	c := sarama.NewConfig()

	tests := []struct {
		name       string
		option     Option
		verifyFunc func(t *testing.T, reporter *kafkaReporter)
	}{
		{
			name:   "with kafka config",
			option: WithConfig(c),
			verifyFunc: func(t *testing.T, reporter *kafkaReporter) {
				if reporter.c != c {
					t.Error("error are not set WithKafkaConfig")
				}
			},
		},
		{
			name:   "with check interval",
			option: WithCheckInterval(time.Second),
			verifyFunc: func(t *testing.T, reporter *kafkaReporter) {
				if reporter.checkInterval != time.Second {
					t.Error("error are not set checkInterval")
				}
			},
		},
		{
			name:   "with serviceInstance props",
			option: WithInstanceProps(instanceProps),
			verifyFunc: func(t *testing.T, reporter *kafkaReporter) {
				var value string
				var ok bool
				if value, ok = reporter.instanceProps["org"]; !ok {
					t.Error("error are not set instanceProps")
				}
				if value != "SkyAPM" {
					t.Error("error are not set instanceProps")
				}
			},
		},
		{
			name:   "with logger",
			option: WithLogger(logger),
			verifyFunc: func(t *testing.T, reporter *kafkaReporter) {
				if reporter.logger != logger {
					t.Error("error are not set logger")
				}
			},
		},
		{
			name:   "with topic management",
			option: WithTopicManagement("test_management"),
			verifyFunc: func(t *testing.T, reporter *kafkaReporter) {
				if reporter.topicManagement != "test_management" {
					t.Error("error are not set WithKafkaTopicManagement")
				}
			},
		},
		{
			name:   "with topic segment",
			option: WithTopicSegment("test_segment"),
			verifyFunc: func(t *testing.T, reporter *kafkaReporter) {
				if reporter.topicSegment != "test_segment" {
					t.Error("error are not set WithKafkaTopicSegment")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := createKafkaReporter()
			tt.option(reporter)
			tt.verifyFunc(t, reporter)
		})
	}
}

func TestKafkaReporter_reportInstanceProperties(t *testing.T) {
	customProps := make(map[string]string)
	customProps["org"] = "SkyAPM"
	osProps := buildOSInfo()
	for k, v := range customProps {
		osProps = append(osProps, &commonv3.KeyStringValuePair{
			Key:   k,
			Value: v,
		})
	}

	reporter := createKafkaReporter()
	reporter.service = mockService
	reporter.serviceInstance = mockServiceInstance
	reporter.instanceProps = customProps

	c := mocks.NewTestConfig()
	c.Producer.Return.Successes = true
	c.Producer.Return.Errors = false
	mp := mocks.NewAsyncProducer(t, c)
	mp.ExpectInputAndSucceed()
	reporter.producer = mp

	err := reporter.reportInstanceProperties()
	if err != nil {
		t.Error()
	}
	for msg := range reporter.producer.Successes() {
		_ = reporter.producer.Close()
		if msg.Topic != reporter.topicManagement {
			t.Errorf("Excepted kafka topic is %s not %s", reporter.topicManagement, msg.Topic)
		}
		v, _ := msg.Value.Encode()
		var s managementv3.InstanceProperties
		if err := proto.Unmarshal(v, &s); err != nil {
			t.Fatal(err)
		}
		if s.Service != mockService {
			t.Error("error are not set service")
		}
		if s.ServiceInstance != mockServiceInstance {
			t.Error("error are not set service instance")
		}
		if len(s.Properties) != len(osProps) {
			t.Error("error are not set service Properties")
		}
	}
}

func createKafkaReporter() *kafkaReporter {
	reporter := &kafkaReporter{
		logger:          log.New(os.Stderr, "go2sky", log.LstdFlags),
		topicManagement: defaultTopicManagement,
		topicSegment:    defaultTopicSegment,
	}
	return reporter
}
