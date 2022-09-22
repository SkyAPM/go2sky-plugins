# Go2sky with Ent.

## Installation

```bash
1. Define an schema type.
2. Generate the specified Ent Model files.
3. Apply it to your application.
```

## Steps
```go
go install entgo.io/ent/cmd/ent@v0.11.2
ent generate --feature sql/upsert --target ./gen/entschema ./schema
```

## Usage
```go
package main

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"github.com/SkyAPM/go2sky"
	sqlPlugin "github.com/SkyAPM/go2sky-plugins/sql"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/pkg/errors"
	v3 "skywalking.apache.org/repo/goapi/collect/language/agent/v3"

	"github.com/SkyAPM/go2sky-plugins/ent/gen/entschema"
	entuser "github.com/SkyAPM/go2sky-plugins/ent/gen/entschema/user"
)

func main() {
	r, err := reporter.NewGRPCReporter("127.0.0.1:11800", reporter.WithCheckInterval(time.Second*10))
	if err != nil {
		panic(err)
	}
	tracer, err := go2sky.NewTracer("example", go2sky.WithReporter(r))
	if err != nil {
		panic(err)
	}

	entClient01 := NewEntClient01(tracer)

	go func() {
		doWork(tracer, entClient01)
	}()

	entClient02 := NewEntClient02(tracer)

	go func() {
		doWork(tracer, entClient02)
	}()
	
	select {}
}
```