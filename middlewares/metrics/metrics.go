package metrics

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/metrics"
	"github.com/caicloud/nirvana/service"
)

// Default returns a metric Middleware under the namespace "nirvana".
//
// Once called, the namespace is set and can not be changed. Future attempt to build more Middleware
// will result in ones with the same namespace as the first one.
//
// Unlike the metrics plugin which takes care of everything, you must call Descriptor() to build a
// Descriptor and configure it to a server yourself.
func Default() definition.Middleware {
	metrics.Install("")
	return func(ctx context.Context, next definition.Chain) error {
		startTime := time.Now()
		err := next.Continue(ctx)
		metrics.RecordRequest(startTime, service.HTTPContextFrom(ctx))
		return err
	}
}

// Namespace returns a metric Middleware under the given namespace.
//
// Once called, the namespace is set and can not be changed. Future attempt to build more Middleware
// will result in ones with the same namespace as the first one.
//
// Unlike the metrics plugin which takes care of everything, you must call Descriptor() to build a
// Descriptor and configure it to a server yourself.
func Namespace(namespace string) definition.Middleware {
	metrics.Install(namespace)
	return func(ctx context.Context, next definition.Chain) error {
		startTime := time.Now()
		err := next.Continue(ctx)
		metrics.RecordRequest(startTime, service.HTTPContextFrom(ctx))
		return err
	}
}

// Descriptor returns a descriptor for the API; it must be configured to a server in order to serve the
// metric API.
func Descriptor(path string) definition.Descriptor {
	if path == "" {
		path = "/metrics"
	}
	return definition.SimpleDescriptor(definition.Get, path, service.WrapHTTPHandler(promhttp.Handler()))
}
