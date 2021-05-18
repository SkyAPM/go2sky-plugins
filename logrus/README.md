# Go2sky with logrus (v1.8.1)

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/logrus
```

## Usage

```go
package main

import (
	"context"
	"github.com/sirupsen/logrus"
	logrusplugin "github.com/SkyAPM/go2sky-plugins/logrus"
)

func main() {
	// init format with custom trace context key
	// SW_CTX format: [$serviceName,$instanceName,$traceId,$traceSegmentId,$spanId]
	logrus.SetFormatter(logrusplugin.Wrap(&logrus.JSONFormatter{}, "SW_CTX"}))

	// init tracer

	// log with context
	ctx := context.Background()
	logrus.WithContext(ctx).Info("test1")
}
```

[See more](example_logrus_test.go).