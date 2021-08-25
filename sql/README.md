# Go2Sky with database/sql

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/sql
```

## Usage

```go
import (
	sqlPlugin "github.com/SkyAPM/go2sky-plugins/sql"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	_ "github.com/go-sql-driver/mysql"
)

// init reporter
re, err := reporter.NewLogReporter()
defer re.Close()

// init tracer
tracer, err := go2sky.NewTracer("service-name", go2sky.WithReporter(re))
if err != nil {
    log.Fatalf("init tracer error: %v", err)
}

// use sql plugin to open db with tracer
db, err := sqlPlugin.Open("mysql", dsn, tracer,
    sqlPlugin.WithSqlDBType(sqlPlugin.MYSQL),
    sqlPlugin.WithQueryReport(),
    sqlPlugin.WithParamReport(),
    sqlPlugin.WithPeerAddr("127.0.0.1:3306"),
)
if err != nil {
	log.Fatalf("open db error: %v \n", err)
}

// use db handler as usual.
```