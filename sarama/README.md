# Go2sky with sarama reporter

## Installation

```go
go get -u github.com/SkyAPM/go2sky-plugins/sarama
```

## Usage

```go
r, err := sarama.NewKafkaReporter([]string{"localhost:9092"})
if err != nil {
    log.Fatalf("new kafka reporter error %v \n", err)
}
defer r.Close()
tracer, err := go2sky.NewTracer("example", go2sky.WithReporter(r))
```
