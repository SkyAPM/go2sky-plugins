# Go2sky with kafka reporter

## Installation

```go
go get -u github.com/SkyAPM/go2sky-plugins/kafkareporter
```

## Usage

```go
r, err := kafkareporter.New([]string{"localhost:9092"})
if err != nil {
    log.Fatalf("new kafka reporter error %v \n", err)
}
defer r.Close()
tracer, err := go2sky.NewTracer("example", go2sky.WithReporter(r))
```
