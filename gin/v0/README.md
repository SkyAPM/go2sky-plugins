# Go2sky with gin(1.3.0)

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/gin
```

## Usage
```go
package main

import (
	"log"

	"github.com/SkyAPM/go2sky"
	gearplugin "github.com/SkyAPM/go2sky-plugins/gear"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/teambition/gear"
)

func main() {
	re, err := reporter.NewLogReporter()
	if err != nil {
		log.Fatalf("new reporter error %v \n", err)
	}

	defer re.Close()

	tracer, err := go2sky.NewTracer("gear", go2sky.WithReporter(re))
	if err != nil {
		log.Fatalf("create tracer error %v \n", err)
	}

	app := gear.New()
	app.Use(gearplugin.Middleware(tracer))

	// do something
}
```