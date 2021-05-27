# Go2sky with zap (v1.16.0)

## Installation

```bash
go get -u github.com/SkyAPM/go2sky-plugins/zap
```

## Usage

```go
package main

import (
	"context"
	
	zapplugin "github.com/SkyAPM/go2sky-plugins/zap"
	"go.uber.org/zap"
)

func main() {
	ctx := context.Background()
	logger := zap.NewExample()
	
	// You have two way to adopt
	// 1. Addition fields before logging
	logger.With(zapplugin.TraceContext(ctx)...).Info("test")
	
	// 2. Wrap logger and correlate context at logging
	logger = zapplugin.WrapWithContext(logger)
	logger.Info(ctx, "test")
}
```

[See more](example_zap_test.go).