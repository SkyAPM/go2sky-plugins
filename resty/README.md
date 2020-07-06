# Go2sky with go-resty(v2.2.0)

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/resty
```

## Usage
```go
package main

import (
	"log"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky-plugins/resty"
	"github.com/SkyAPM/go2sky/reporter"
)

func main() {
	// Use gRPC reporter for production
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}
	defer re.Close()

	tracer, err := go2sky.NewTracer("gin-server", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	// create resty client
	client := resty.NewGoResty(tracer)
	// do something
}
```

[See more](example_go_resty_test.go)