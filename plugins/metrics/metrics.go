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

package metrics

import (
	"context"
	"strconv"
	"time"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	nirvana.RegisterConfigInstaller(&metricsInstaller{})
}

// ExternalConfigName is the external config name of metrics.
const ExternalConfigName = "metrics"

// config is metrics config.
type config struct {
	path      string
	namespace string
}

type metricsInstaller struct{}

func newMetricsMiddleware(namespace string) definition.Middleware {
	constLabel := prometheus.Labels{"component": namespace}
	requestCounter := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace:   namespace,
			Name:        "request_total",
			Help:        "Counter of server requests broken out for each verb, API resource, and HTTP response code.",
			ConstLabels: constLabel,
		},
		[]string{"method", "path", "code"},
	)
	requestLatencies := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace:   namespace,
			Name:        "request_duration_seconds",
			Help:        "Response latency distribution in seconds for each verb, resource and client.",
			ConstLabels: constLabel,
			Buckets:     prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	prometheus.MustRegister(requestCounter)
	prometheus.MustRegister(requestLatencies)

	return func(ctx context.Context, next definition.Chain) error {
		startTime := time.Now()
		err := next.Continue(ctx)

		httpCtx := service.HTTPContextFrom(ctx)
		req := httpCtx.Request()
		resp := httpCtx.ResponseWriter()
		path := httpCtx.RoutePath()

		requestCounter.WithLabelValues(req.Method, path, strconv.Itoa(resp.StatusCode())).Inc()
		requestLatencies.WithLabelValues(req.Method, path).Observe(float64((time.Since(startTime)) / time.Second))

		return err
	}
}

// Name is the external config name.
func (i *metricsInstaller) Name() string {
	return ExternalConfigName
}

// Install installs stuffs before server starting.
func (i *metricsInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {

		monitorMiddleware := definition.Descriptor{
			Path:        "/",
			Middlewares: []definition.Middleware{newMetricsMiddleware(c.namespace)},
		}
		metricsEndpoint := definition.SimpleDescriptor(definition.Get, c.path, service.WrapHTTPHandler(promhttp.Handler()))
		err = builder.AddDescriptor(monitorMiddleware, metricsEndpoint)
	})
	return err
}

// Uninstall uninstalls stuffs after server terminating.
func (i *metricsInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	return nil
}

// Disable returns a configurer to disable metrics.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}

// Default Configurer does nothing but ensure default config was set.
func Default() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
		})
		return nil
	}
}

func wrapper(c *nirvana.Config, f func(c *config)) {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{
			path:      "/metrics",
			namespace: "nirvana",
		}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// Path returns a configurer to set metrics path.
func Path(path string) nirvana.Configurer {
	if path == "" {
		path = "/metrics"
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.path = path
		})
		return nil
	}
}

// Namespace returns a configure to set metrics namespace.
func Namespace(ns string) nirvana.Configurer {
	if ns == "" {
		ns = "nirvana"
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.namespace = ns
		})
		return nil
	}
}

// Option contains basic configurations of metrics.
type Option struct {
	// Namespace is metrics namespace.
	Namespace string `desc:"Metrics namespace"`
	// Path is metrics path.
	Path string `desc:"Metrics path"`
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		Namespace: "nirvana",
		Path:      "/metrics",
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	cfg.Configure(
		Namespace(p.Namespace),
		Path(p.Path),
	)
	return nil
}
