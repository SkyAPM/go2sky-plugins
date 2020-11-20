# Go2sky with [go-restful](https://github.com/emicklei/go-restful) (v3+)

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/go-restful
```

## Usage
```go
package main

import (
    "io"
	"log"

    "github.com/SkyAPM/go2sky"
    tracerestful "github.com/SkyAPM/go2sky-plugins/go-restful"
    "github.com/SkyAPM/go2sky/reporter"
    "github.com/emicklei/go-restful/v3"
)

func main() {
    // Use gRPC reporter for production
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}

	defer re.Close()

	tracer, err := go2sky.NewTracer("go-restful", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	ws := new(restful.WebService)
    	ws.Filter(tracerestful.NewTraceFilterFunction(tracer))
    
    	ws.Route(ws.GET("/hello").To(func(req *restful.Request, resp *restful.Response) {
    		io.WriteString(resp, "go-restful")
    	}))
    restful.Add(ws)
    go func() {
        http.ListenAndServe(":8080", nil)
    }()
	// do something
}
```

[See more](example_go_restful_test.go).