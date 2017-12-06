package trace

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
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
	eventRequest event = "request"
	eventError   event = "error"
	// eventResponse = "response"
)

var defaultContentTypes = []string{"application/json"}

// once the request is received, it'll be invoked before entering the next middleware.
// You can customize here to record some of the span information.
type OnRequest func(span opentracing.Span, req *http.Request)

// type OnResponse func(span opentracing.Span, req *http.Request)

// Config is trace middleware configuration.
type Config struct {
	// You Can call 'NewDefaultTracerClient' get a default configuration tracer client.
	// Or use the 'github.com/uber/jaeger-client-go' custom configurations.
	Tracer opentracing.Tracer
	// The middleware will record the reqeust body and the response body by default.
	// if Disable, the middleware will not record it.
	DisableRecordFull bool
	// Need to record the 'content type' of the full request. By default only 'application/json'
	RecordContentTypes []string
	OnRequest          OnRequest
	// OnResponse         OnResponse
}

// New created trace middlewares.
func (c *Config) New() func(context.Context, router.RoutingChain) error {
	if len(c.RecordContentTypes) == 0 {
		c.RecordContentTypes = defaultContentTypes
	}

	return func(ctx context.Context, next router.RoutingChain) error {
		req := web.HTTPRequest(ctx)

		// extract span context from HTTP Headers
		spanContext, _ := c.Tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(req.Header)) // nolint: errcheck

		// TODO(yejiayu): abstract path
		span := c.Tracer.StartSpan(req.URL.Path, ext.RPCServerOption(spanContext))
		defer span.Finish()

		// set standard tags
		ext.HTTPUrl.Set(span, req.URL.String())
		ext.HTTPMethod.Set(span, req.Method)
		ext.Component.Set(span, "nirvana/middlewares/trace")
		span.SetTag("Request-Id", req.Header.Get("Request-Id"))

		req, err := c.logsRequest(span, req)
		if err != nil {
			return err
		}
		if c.OnRequest != nil {
			c.OnRequest(span, req)
		}

		ctx = opentracing.ContextWithSpan(ctx, span)
		if err := next.Continue(ctx); err != nil {
			ext.HTTPStatusCode.Set(span, 500)
			ext.Error.Set(span, true)
			span.LogFields(
				log.String("event", string(eventError)),
				log.Error(err),
			)
		}

		resp := web.HTTPResponseWriter(ctx)
		code := resp.StatusCode()
		ext.HTTPStatusCode.Set(span, uint16(code))
		if code >= 400 {
			ext.Error.Set(span, true)
		}

		// TODO(yejiayu): logs response

		return nil
	}
}

func (c *Config) logsRequest(span opentracing.Span, req *http.Request) (*http.Request, error) {
	if req.URL.Query().Encode() != "" {
		span.LogFields(
			log.String("event", string(eventRequest)),
			log.String("query", req.URL.Query().Encode()),
		)
	}

	if c.DisableRecordFull {
		return req, nil
	}

	method := req.Method
	if method != http.MethodPost && method != http.MethodPut && method != http.MethodPatch {
		return req, nil
	}

	contentType := req.Header.Get("Content-Type")
	for _, ct := range c.RecordContentTypes {
		if contentType == ct {
			defer req.Body.Close() // nolint: errcheck
			b, err := ioutil.ReadAll(req.Body)
			if err != nil {
				return nil, err
			}

			span.LogFields(
				log.String("event", string(eventRequest)),
				log.String("body", string(b)),
			)
			req.Body = ioutil.NopCloser(bytes.NewReader(b))
			break
		}
	}

	return req, nil
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
