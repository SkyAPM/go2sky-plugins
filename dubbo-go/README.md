### Go2Sky with Dubbo-go

##### Installation

```
go get -u github.com/SkyAMP/go2sky-plugins/dubbo-go
```

##### Usage

Server:

```go
import (
    _ "dubbo.apache.org/dubbo-go/v3/cluster/cluster_impl"
	_ "dubbo.apache.org/dubbo-go/v3/cluster/loadbalance"
	"dubbo.apache.org/dubbo-go/v3/common/logger"
	_ "dubbo.apache.org/dubbo-go/v3/common/proxy/proxy_factory"
	"dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/filter/filter_impl"
	_ "dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	_ "dubbo.apache.org/dubbo-go/v3/registry/protocol"
	_ "dubbo.apache.org/dubbo-go/v3/registry/zookeeper"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	hessian "github.com/apache/dubbo-go-hessian2"

	dubbo_go "github.com/SkyAPM/go2sky-plugins/dubbo-go"
	"github.com/SkyAPM/go2sky-plugins/dubbo-go/test/server/pkg"
)

// set dubbogo configs ...

// setup reporter, use gRPC reporter for production
report, err := reporter.NewLogReporter()
if err != nil {
    log.Fatalf("new reporter error: %v \n", err)
}

// setup tracer
tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(report))
if err != nil {
    log.Fatalf("crate tracer error: %v \n", err)
}

// set dubbogo plugin server tracer
err = dubbo_go.SetServerTracer(tracer)
if err != nil {
    log.Fatalf("set tracer error: %v \n", err)
}

// set extra tags and report tags
dubbo_go.SetServerExtraTags("extra-tags", "server")
dubbo_go.SetServerReportTags("release")
```

Client:

```go
import (
	_ "dubbo.apache.org/dubbo-go/v3/cluster/cluster_impl"
	_ "dubbo.apache.org/dubbo-go/v3/cluster/loadbalance"
	_ "dubbo.apache.org/dubbo-go/v3/common/proxy/proxy_factory"
	"dubbo.apache.org/dubbo-go/v3/config"
	_ "dubbo.apache.org/dubbo-go/v3/filter/filter_impl"
	_ "dubbo.apache.org/dubbo-go/v3/protocol/dubbo"
	_ "dubbo.apache.org/dubbo-go/v3/registry/protocol"
	_ "dubbo.apache.org/dubbo-go/v3/registry/zookeeper"
	"github.com/SkyAPM/go2sky"
	"github.com/SkyAPM/go2sky/reporter"
	hessian "github.com/apache/dubbo-go-hessian2"

	dubbo_go "github.com/SkyAPM/go2sky-plugins/dubbo-go"
	"github.com/SkyAPM/go2sky-plugins/dubbo-go/test/client/pkg"
)

// set dubbogo configs ...

// setup reporter, use gRPC reporter for production
report, err := reporter.NewLogReporter()
if err != nil {
    log.Fatalf("new reporter error: %v \n", err)
}

// setup tracer
tracer, err := go2sky.NewTracer(serviceName, go2sky.WithReporter(report))
if err != nil {
    log.Fatalf("crate tracer error: %v \n", err)
}

// set dubbogo plugin client tracer
err = dubbo_go.SetClientTracer(tracer)
if err != nil {
    log.Fatalf("set tracer error: %v \n", err)
}

// set extra tags and report tags
dubbo_go.SetClientExtraTags("extra-tags", "client")
dubbo_go.SetClientReportTags("release")
```

