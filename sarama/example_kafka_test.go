package sarama

import (
	"log"

	"github.com/SkyAPM/go2sky"
)

func ExampleNewKafkaReporter() {
	r, err := NewKafkaReporter([]string{"localhost:9092"})
	if err != nil {
		log.Fatalf("new kafka reporter error %v \n", err)
	}
	defer r.Close()

	_, err = go2sky.NewTracer("example", go2sky.WithReporter(r))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	// Output:
}
