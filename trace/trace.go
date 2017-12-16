/*
Copyright 2017 Caicloud Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package trace

import (
	"context"
	"io"
	"time"

	"github.com/caicloud/nirvana/router"
	"github.com/caicloud/nirvana/web"
	"github.com/golang/glog"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"github.com/opentracing/opentracing-go/log"
	"github.com/uber/jaeger-client-go/config"
)

type event string

const (
	eventRequest  event = "request"
	eventResponse event = "response"
)

// Config is trace middleware configuration.
type Config struct {
	// You Can call 'NewDefaultTracerClient' get a default configuration tracer client.
	// Or use the 'github.com/uber/jaeger-client-go' custom configurations.
	Tracer opentracing.Tracer
}

// New created trace middlewares.
func New(c *Config) func(context.Context, router.RoutingChain) error {
	return func(ctx context.Context, next router.RoutingChain) error {
		req := web.HTTPRequest(ctx)

		// extract span context from HTTP Headers
		spanContext, err := c.Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header))
		if err != nil {
			glog.Error(err)
		}

		// TODO(yejiayu): abstract path
		span := c.Tracer.StartSpan(req.URL.Path, ext.RPCServerOption(spanContext))
		defer span.Finish()

		// set standard tags
		ext.HTTPUrl.Set(span, req.URL.String())
		ext.HTTPMethod.Set(span, req.Method)
		ext.Component.Set(span, "nirvana/middlewares/trace")
		span.SetTag("Request-Id", req.Header.Get("Request-Id"))

		span.LogFields(
			log.String("event", string(eventRequest)),
		)

		ctx = opentracing.ContextWithSpan(ctx, span)

		defer func() {
			span.LogFields(
				log.String("event", string(eventResponse)),
			)
		}()
		if err := next.Continue(ctx); err != nil {
			ext.HTTPStatusCode.Set(span, 500)
			ext.Error.Set(span, true)
			return err
		}

		resp := web.HTTPResponseWriter(ctx)
		code := resp.StatusCode()
		ext.HTTPStatusCode.Set(span, uint16(code))
		if code >= 400 {
			ext.Error.Set(span, true)
		}

		return nil
	}
}

// NewDefaultTracerClient created a default configuration tracer client.
func NewDefaultTracerClient(serviceName string, agentHostPort string) (opentracing.Tracer, io.Closer) {
	cfg := config.Configuration{
		Sampler: &config.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &config.ReporterConfig{
			LogSpans:            false,
			BufferFlushInterval: 1 * time.Second,
			LocalAgentHostPort:  agentHostPort,
		},
	}

	tracer, ioClose, err := cfg.New(
		serviceName,
		config.Logger(&loggerAdapter{}),
	)
	if err != nil {
		glog.Fatalf("cannot initialize Jaeger Tracer\n%s", err.Error())
	}
	return tracer, ioClose
}

type loggerAdapter struct{}

// Error logs a message at error priority
func (logger *loggerAdapter) Error(msg string) {
	glog.Error(msg)
}

// Infof logs a message at info priority
func (logger *loggerAdapter) Infof(msg string, args ...interface{}) {
	glog.Infof(msg, args)
}
