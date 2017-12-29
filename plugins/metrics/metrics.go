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
	"github.com/caicloud/nirvana/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	nirvana.RegisterConfigInstaller(&metricsInstaller{})
}

// ExternalConfigName is the external config name of metrics.
const ExternalConfigName = "metrics"

// config is metrics config.
type config struct {
	path string
}

type metricsInstaller struct{}

// Name is the external config name.
func (i *metricsInstaller) Name() string {
	return ExternalConfigName
}

// Install installs stuffs before server starting.
func (i *metricsInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wapper(cfg, func(c *config) {
		err = builder.AddDescriptor(definition.SimpleDescriptor(definition.Get, c.path, service.WarpHTTPHandler(promhttp.Handler())))
	})
	return err

}

// Uninstall uninstalls stuffs after server terminating.
func (i *metricsInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	return nil
}

// Disable disables metrics.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}

func wapper(c *nirvana.Config, f func(c *config)) {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{
			path: "/metrics",
		}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// Path sets metrics path. Empty path means /metrics.
func Path(path string) nirvana.Configurer {
	if path == "" {
		path = "/metrics"
	}
	return func(c *nirvana.Config) error {
		wapper(c, func(c *config) {
			c.path = path
		})
		return nil
	}
}
