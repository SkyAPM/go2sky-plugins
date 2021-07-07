# Go2Sky with database/sql

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/sql
```

## Usage

```go
import (
    "database/sql"

	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	"github.com/go-sql-driver/mysql"
)

// init reporter
re, err := reporter.NewLogReporter()
defer re.Close()

// init tracer
tracer, err := go2sky.NewTracer("service-name", go2sky.WithReporter(re))

// register go2sky sql wrapper
sql.Register("skywalking-sql", NewTracerDriver(&mysql.MySQLDriver{}, tracer, MYSQL))
db, err := sql.Open("skywalking-sql", "user:password@tcp(127.0.0.1:3306)/dbname")

// use db handle as usual.
```