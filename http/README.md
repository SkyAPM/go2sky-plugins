# Go2sky with net/http

## Installation

```bash
go get -u github.com/SkyAPM/go2sky
```

## Usage

### Server
```go
package main

import (
	"log"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/plugins/http"
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

	sm, err := http.NewServerMiddleware(tracer)
	if err != nil {
		log.Fatalf("create server middleware error %v \n", err)
	}
	// do something
}
```

### Client

```go
package main

import (
	"log"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/plugins/http"
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

	sm, err := http.NewClient(tracer)
	if err != nil {
		log.Fatalf("create client error %v \n", err)
	}
	// do something
}
```

[See more](https://github.com/SkyAPM/go2sky/blob/master/plugins/http/example_http_test.go)