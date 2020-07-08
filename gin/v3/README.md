# Go2sky with gin (v1.5.0)

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/gin/v3
```

## Usage
```go
package main

import (
	"log"

	"github.com/SkyAPM/go2sky"
	v3 "github.com/SkyAPM/go2sky-plugins/gin/v3"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/gin-gonic/gin"
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

	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	//Use go2sky middleware with tracing
	r.Use(v3.Middleware(r, tracer))

	// do something
}
```

[See more](example_gin_test.go).