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
	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/metrics"
	metricsmiddleware "github.com/caicloud/nirvana/middlewares/metrics"
	"github.com/caicloud/nirvana/service"
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
			Middlewares: []definition.Middleware{metricsmiddleware.Restful(&metrics.Options{NamespaceValue: c.namespace})},
		}
		err = builder.AddDescriptor(monitorMiddleware, metricsmiddleware.Descriptor(c.path))
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
