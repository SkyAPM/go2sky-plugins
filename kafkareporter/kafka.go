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
	"sync"
	"time"

	"github.com/Shopify/sarama"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky-plugins/kafkareporter/internal/tool"
	"google.golang.org/protobuf/proto"
	commonv3 "skywalking.apache.org/repo/goapi/collect/common/v3"
	agentv3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"
	managementv3 "skywalking.apache.org/repo/goapi/collect/management/v3"
)

const (
	defaultCheckInterval   = 20 * time.Second
	defaultKafkaLogPrefix  = "go2sky-kafka"
	topicKeyRegister       = "register-"
	defaultTopicManagement = "skywalking-managements"
	defaultTopicSegment    = "skywalking-segments"
)

type kafkaReporter struct {
	c               *sarama.Config
	producer        sarama.AsyncProducer
	service         string
	serviceInstance string
	wg              sync.WaitGroup
	ctx             context.Context
	cancel          context.CancelFunc
	instanceProps   map[string]string
	logger          *log.Logger
	topicManagement string
	topicSegment    string
	checkInterval   time.Duration
	// cdsInterval      time.Duration
	// cdsService       *go2sky.ConfigDiscoveryService
	// cdsClient        configuration.ConfigurationDiscoveryServiceClient
}

// New create a new reporter to send data to kafka.
func New(addrs []string, opts ...Option) (go2sky.Reporter, error) {
	r := &kafkaReporter{
		c:               sarama.NewConfig(),
		logger:          log.New(os.Stderr, defaultKafkaLogPrefix, log.LstdFlags),
		checkInterval:   defaultCheckInterval,
		topicManagement: defaultTopicManagement,
		topicSegment:    defaultTopicSegment,
	}

	for _, o := range opts {
		o(r)
	}

	p, err := sarama.NewAsyncProducer(addrs, r.c)
	if err != nil {
		return nil, err
	}
	r.producer = p

	if r.c.Producer.Return.Errors {
		go func() {
			for e := range p.Errors() {
				r.logger.Printf("send kafka err: %v", e.Err)
			}
		}()
	}

	return r, nil
}

// Option allows for functional options to adjust behaviour
// of a kafka reporter to be created by New
type Option func(r *kafkaReporter)

// WithConfig setup sarama.Config for kafka reporter
func WithConfig(c *sarama.Config) Option {
	return func(r *kafkaReporter) {
		r.c = c
	}
}

// WithCheckInterval setup service and endpoint registry check interval
func WithCheckInterval(interval time.Duration) Option {
	return func(r *kafkaReporter) {
		r.checkInterval = interval
	}
}

// WithInstanceProps setup service instance properties eg: org=SkyAPM
func WithInstanceProps(props map[string]string) Option {
	return func(r *kafkaReporter) {
		r.instanceProps = props
	}
}

// WithLogger setup logger for kafka reporter
func WithLogger(logger *log.Logger) Option {
	return func(r *kafkaReporter) {
		r.logger = logger
	}
}

// WithTopicManagement setup service management topic
func WithTopicManagement(topicManagement string) Option {
	return func(r *kafkaReporter) {
		r.topicManagement = topicManagement
	}
}

// WithTopicSegment setup service segment topic
func WithTopicSegment(topicSegment string) Option {
	return func(r *kafkaReporter) {
		r.topicSegment = topicSegment
	}
}

func (r *kafkaReporter) Boot(service string, serviceInstance string, cdsWatchers []go2sky.AgentConfigChangeWatcher) {
	r.service = service
	r.serviceInstance = serviceInstance
	r.ctx, r.cancel = context.WithCancel(context.Background())
	r.check()
}

func (r *kafkaReporter) reportInstanceProperties() error {
	props := buildOSInfo()
	if r.instanceProps != nil {
		for k, v := range r.instanceProps {
			props = append(props, &commonv3.KeyStringValuePair{
				Key:   k,
				Value: v,
			})
		}
	}
	instanceProperties := &managementv3.InstanceProperties{
		Service:         r.service,
		ServiceInstance: r.serviceInstance,
		Properties:      props,
	}
	b, err := proto.Marshal(instanceProperties)
	if err != nil {
		return err
	}

	r.producer.Input() <- &sarama.ProducerMessage{
		Topic: r.topicManagement,
		Key:   sarama.StringEncoder(topicKeyRegister + instanceProperties.ServiceInstance),
		Value: sarama.ByteEncoder(b),
	}
	return nil
}

func (r *kafkaReporter) check() {
	if r.checkInterval < 0 || r.producer == nil {
		return
	}
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()
		ticker := time.NewTicker(r.checkInterval)
		defer ticker.Stop()
		instancePropertiesSubmitted := false
		for {
			select {
			case <-r.ctx.Done():
				return
			case <-ticker.C:
				if !instancePropertiesSubmitted {
					err := r.reportInstanceProperties()
					if err != nil {
						r.logger.Printf("report serviceInstance properties error %v", err)
						continue
					}
					instancePropertiesSubmitted = true
				}

				instancePingPkg := &managementv3.InstancePingPkg{
					Service:         r.service,
					ServiceInstance: r.serviceInstance,
				}
				b, err := proto.Marshal(instancePingPkg)
				if err != nil {
					r.logger.Printf("send keep alive signal error %v", err)
					continue
				}

				r.producer.Input() <- &sarama.ProducerMessage{
					Topic: r.topicManagement,
					Key:   sarama.StringEncoder(instancePingPkg.ServiceInstance),
					Value: sarama.ByteEncoder(b),
				}
			}
		}
	}()
}

func (r *kafkaReporter) Send(spans []go2sky.ReportedSpan) {
	spanSize := len(spans)
	if spanSize < 1 {
		return
	}
	rootSpan := spans[spanSize-1]
	rootCtx := rootSpan.Context()
	segmentObject := &agentv3.SegmentObject{
		TraceId:         rootCtx.TraceID,
		TraceSegmentId:  rootCtx.SegmentID,
		Spans:           make([]*agentv3.SpanObject, spanSize),
		Service:         r.service,
		ServiceInstance: r.serviceInstance,
	}
	for i, s := range spans {
		spanCtx := s.Context()
		segmentObject.Spans[i] = &agentv3.SpanObject{
			SpanId:        spanCtx.SpanID,
			ParentSpanId:  spanCtx.ParentSpanID,
			StartTime:     s.StartTime(),
			EndTime:       s.EndTime(),
			OperationName: s.OperationName(),
			Peer:          s.Peer(),
			SpanType:      s.SpanType(),
			SpanLayer:     s.SpanLayer(),
			ComponentId:   s.ComponentID(),
			IsError:       s.IsError(),
			Tags:          s.Tags(),
			Logs:          s.Logs(),
		}
		srr := make([]*agentv3.SegmentReference, 0)
		if i == (spanSize-1) && spanCtx.ParentSpanID > -1 {
			srr = append(srr, &agentv3.SegmentReference{
				RefType:               agentv3.RefType_CrossThread,
				TraceId:               spanCtx.TraceID,
				ParentTraceSegmentId:  spanCtx.ParentSegmentID,
				ParentSpanId:          spanCtx.ParentSpanID,
				ParentService:         r.service,
				ParentServiceInstance: r.serviceInstance,
			})
		}
		if len(s.Refs()) > 0 {
			for _, tc := range s.Refs() {
				srr = append(srr, &agentv3.SegmentReference{
					RefType:                  agentv3.RefType_CrossProcess,
					TraceId:                  spanCtx.TraceID,
					ParentTraceSegmentId:     tc.ParentSegmentID,
					ParentSpanId:             tc.ParentSpanID,
					ParentService:            tc.ParentService,
					ParentServiceInstance:    tc.ParentServiceInstance,
					ParentEndpoint:           tc.ParentEndpoint,
					NetworkAddressUsedAtPeer: tc.AddressUsedAtClient,
				})
			}
		}
		segmentObject.Spans[i].Refs = srr
	}

	b, err := proto.Marshal(segmentObject)
	if err != nil {
		r.logger.Printf("reporter segment err %v", err)
		return
	}
	select {
	case <-r.ctx.Done():
		r.logger.Printf("reporter segment closed")
		return
	default:
	}
	r.producer.Input() <- &sarama.ProducerMessage{
		Topic: r.topicSegment,
		Key:   sarama.StringEncoder(segmentObject.TraceSegmentId),
		Value: sarama.ByteEncoder(b),
	}
}

func (r *kafkaReporter) Close() {
	r.cancel()
	r.wg.Wait()
	if err := r.producer.Close(); err != nil {
		r.logger.Print(err)
	}
}

func buildOSInfo() (props []*commonv3.KeyStringValuePair) {
	processNo := tool.ProcessNo()
	if processNo != "" {
		kv := &commonv3.KeyStringValuePair{
			Key:   "Process No.",
			Value: processNo,
		}
		props = append(props, kv)
	}

	hostname := &commonv3.KeyStringValuePair{
		Key:   "hostname",
		Value: tool.HostName(),
	}
	props = append(props, hostname)

	language := &commonv3.KeyStringValuePair{
		Key:   "language",
		Value: "go",
	}
	props = append(props, language)

	osName := &commonv3.KeyStringValuePair{
		Key:   "OS Name",
		Value: tool.OSName(),
	}
	props = append(props, osName)

	ipv4s := tool.AllIPV4()
	if len(ipv4s) > 0 {
		for _, ipv4 := range ipv4s {
			kv := &commonv3.KeyStringValuePair{
				Key:   "ipv4",
				Value: ipv4,
			}
			props = append(props, kv)
		}
	}
	return
}
