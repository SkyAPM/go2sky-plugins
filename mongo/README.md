# Go2Sky with Mongo

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/mongo
```

## Usage

```go
import (
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	
	mongoPlugin "go2sky-plugins/mongo"
)

// init reporter
re, err := reporter.NewLogReporter()
defer re.Close()

// init tracer
tracer, err := go2sky.NewTracer("service-name", go2sky.WithReporter(re))
if err != nil {
    log.Fatalf("init tracer error: %v", err)
}

// init connect mongodb.
client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(dsn).SetMonitor(mongoPlugin.Middleware(tracer)))
if err != nil {
    log.Fatalf("connect mongodb error %v \n", err)
}

...

```
