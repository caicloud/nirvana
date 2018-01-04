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
	"context"
	"fmt"
	"net/http"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

// Server is a complete API server.
// The server contains a router to handle all requests form clients.
type Server interface {
	// Serve starts to listen and serve requests.
	// The method won't return except an error occurs.
	Serve() error
}

// Config describes configuration of server.
type Config struct {
	// IP is the ip to listen. Empty means `0.0.0.0`.
	IP string
	// Port is the port to listen.
	Port uint16
	// Logger is used to output info inside framework.
	Logger log.Logger
	// Descriptors contains all APIs.
	Descriptors []definition.Descriptor
	// Filters is http filters.
	Filters []service.Filter
	// Modifiers is definition modifiers
	Modifiers service.DefinitionModifiers
	// configSet contains all configurations of plugins.
	configSet map[string]interface{}
}

// Configurer is used to configure server config.
type Configurer func(c *Config) error

// Configure configs by configurers. It panics if an error occurs.
func (c *Config) Configure(configurers ...Configurer) *Config {
	for _, configurer := range configurers {
		if err := configurer(c); err != nil {
			panic(err)
		}
	}
	return c
}

// Config gets external config by name.
func (c *Config) Config(name string) interface{} {
	return c.configSet[name]
}

// Set sets external config by name.
// Set a nil config will delete it.
func (c *Config) Set(name string, config interface{}) {
	if config == nil {
		delete(c.configSet, name)
	} else {
		c.configSet[name] = config
	}
}

// forEach traverse all plugin configs.
func (c *Config) forEach(f func(name string, config interface{}) error) error {
	for name, cfg := range c.configSet {
		if err := f(name, cfg); err != nil {
			return err
		}
	}
	return nil
}

// NewDefaultConfig creates default config.
// Default config contains:
//  Filters: RedirectTrailingSlash, FillLeadingSlash, ParseRequestForm.
//  Modifiers: FirstContextParameter, EmptyConsumeForHTTPGet,
//             ConsumeAllIfComsumesIsEmpty, ProduceAllIfProducesIsEmpty,
//             ConsumeNoneForHTTPGet, ConsumeNoneForHTTPDelete,
//             ProduceNoneForHTTPDelete.
func NewDefaultConfig(ip string, port uint16) *Config {
	return NewConfig().Configure(
		IP(ip),
		Port(port),
		Logger(log.DefaultLogger()),
		Filter(
			service.RedirectTrailingSlash(),
			service.FillLeadingSlash(),
			service.ParseRequestForm(),
		),
		Modifier(
			service.FirstContextParameter(),
			service.ConsumeAllIfComsumesIsEmpty(),
			service.ProduceAllIfProducesIsEmpty(),
			service.ConsumeNoneForHTTPGet(),
			service.ConsumeNoneForHTTPDelete(),
			service.ProduceNoneForHTTPDelete(),
		),
	)
}

// NewConfig creates a pure config.
func NewConfig() *Config {
	return &Config{
		IP:          "",
		Port:        80,
		Logger:      &log.SilentLogger{},
		Filters:     []service.Filter{},
		Descriptors: []definition.Descriptor{},
		Modifiers:   []service.DefinitionModifier{},
		configSet:   make(map[string]interface{}),
	}
}

// server is nirvana server.
type server struct {
	config *Config
	server *http.Server
}

// NewServer creates a nirvana server.
func NewServer(c *Config) Server {
	return &server{
		config: c,
	}
}

var noConfigInstaller = errors.InternalServerError.Build("Nirvana:NoConfigInstaller", "no config installer for external config name ${name}")

func (s *server) builder() (service.Builder, error) {
	builder := service.NewBuilder()
	builder.SetLogger(s.config.Logger)
	builder.AddFilter(s.config.Filters...)
	builder.SetModifier(s.config.Modifiers.Combine())
	if err := builder.AddDescriptor(s.config.Descriptors...); err != nil {
		return nil, err
	}
	err := s.config.forEach(func(name string, config interface{}) error {
		installer := ConfigInstallerFor(name)
		if installer == nil {
			return noConfigInstaller.Error(name)
		}
		err := installer.Install(builder, s.config)
		return err
	})
	if err != nil {
		return nil, err
	}
	return builder, nil
}

// Serve starts to listen and serve requests.
// The method won't return except an error occurs.
func (s *server) Serve() error {
	builder, err := s.builder()
	if err != nil {
		return err
	}
	service, err := builder.Build()
	if err != nil {
		return err
	}
	s.server = &http.Server{
		Addr:    fmt.Sprintf("%s:%d", s.config.IP, s.config.Port),
		Handler: service,
	}
	err = s.server.ListenAndServe()
	e := s.config.forEach(func(name string, config interface{}) error {
		installer := ConfigInstallerFor(name)
		if installer == nil {
			s.config.Logger.Error(noConfigInstaller.Error(name))
		}
		err := installer.Uninstall(builder, s.config)
		s.config.Logger.Error(err)
		return nil
	})
	if e != nil {
		s.config.Logger.Error(e)
	}
	return err
}

// Shutdown gracefully shuts down the server without interrupting any
// active connections.
func (s *server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ConfigInstaller is used to install config to service builder.
type ConfigInstaller interface {
	// Name is the external config name.
	Name() string
	// Install installs stuffs before server starting.
	Install(builder service.Builder, config *Config) error
	// Uninstall uninstalls stuffs after server terminating.
	Uninstall(builder service.Builder, config *Config) error
}

var installers = map[string]ConfigInstaller{}

// ConfigInstallerFor gets installer by name.
func ConfigInstallerFor(name string) ConfigInstaller {
	return installers[name]
}

// RegisterConfigInstaller registers a config installer.
func RegisterConfigInstaller(ci ConfigInstaller) {
	if ConfigInstallerFor(ci.Name()) != nil {
		panic(fmt.Sprintf("Config installer %s has been installed.", ci.Name()))
	}
	installers[ci.Name()] = ci
}

// IP returns a configurer to set ip into config.
func IP(ip string) Configurer {
	return func(c *Config) error {
		c.IP = ip
		return nil
	}
}

// Port returns a configurer to set port into config.
func Port(port uint16) Configurer {
	return func(c *Config) error {
		c.Port = port
		return nil
	}
}

// Logger returns a configurer to set logger into config.
func Logger(logger log.Logger) Configurer {
	return func(c *Config) error {
		if logger == nil {
			c.Logger = &log.SilentLogger{}
		} else {
			c.Logger = logger
		}
		return nil
	}
}

// Descriptor returns a configurer to add descriptors into config.
func Descriptor(descriptors ...definition.Descriptor) Configurer {
	return func(c *Config) error {
		c.Descriptors = append(c.Descriptors, descriptors...)
		return nil
	}
}

// Filter returns a configurer to add filters into config.
func Filter(filters ...service.Filter) Configurer {
	return func(c *Config) error {
		c.Filters = append(c.Filters, filters...)
		return nil
	}
}

// Modifier returns a configurer to add definition modifiers into config.
func Modifier(modifiers ...service.DefinitionModifier) Configurer {
	return func(c *Config) error {
		c.Modifiers = append(c.Modifiers, modifiers...)
		return nil
	}
}
