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

package nirvana

import (
	"runtime"

	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/metrics"
	"github.com/caicloud/nirvana/utils/profiling"
)

// Config describes configuration of service.Server
type Config struct {
	EnableProfiling bool
	// Requires generic profiling enabled
	EnableContentionProfiling bool
	EnableMetrics             bool
	// Path of prometheus metrics endpoint, default to '/metrics'
	MetricsPath     string
	Logger          log.Logger
	ConfigureServer func(service.Server)
}

// New creates a new service.Server
func New(c *Config) (service.Server, error) {
	if err := service.RegisterDefaultEnvironment(); err != nil {
		panic(err)
	}
	var s service.Server
	if c.ConfigureServer == nil {
		s = service.NewDefaultServer()
	} else {
		s = service.NewServer()
		c.ConfigureServer(s)
	}
	if c.Logger != nil {
		s.SetLogger(c.Logger)
	}
	installAPI(s, c)
	return s, nil
}

// installAPI installs additional APIs used for debuging, instrumentation ...
func installAPI(s service.Server, c *Config) {
	if c.EnableProfiling {
		profiling.Profiling{}.Install(s)
		if c.EnableContentionProfiling {
			runtime.SetBlockProfileRate(1)
		}
	}
	if c.EnableMetrics {
		metrics.Metrics{}.Install(s, c.MetricsPath)
	}
}
