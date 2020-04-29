package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/metrics"
	"github.com/caicloud/nirvana/service"
)

func Default() definition.Middleware {
	metrics.Install("")
	return func(ctx context.Context, next definition.Chain) error {
		startTime := time.Now()
		err := next.Continue(ctx)
		metrics.RecordRequest(startTime, service.HTTPContextFrom(ctx))
		return err
	}
}

func Namespace(namespace string) definition.Middleware {
	metrics.Install(namespace)
	return func(ctx context.Context, next definition.Chain) error {
		startTime := time.Now()
		err := next.Continue(ctx)
		metrics.RecordRequest(startTime, service.HTTPContextFrom(ctx))
		return err
	}
}

func Descriptor(path string) definition.Descriptor {
	if path == "" {
		path = "/metrics"
	}
	return definition.SimpleDescriptor(definition.Get, path, service.WrapHTTPHandler(promhttp.Handler()))
}
