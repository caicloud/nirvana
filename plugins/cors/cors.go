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

package cors

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/cors"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

func init() {
	nirvana.RegisterConfigInstaller(&corsInstaller{})
}

// ExternalConfigName is the external config name of request logger.
const ExternalConfigName = "cors"

// config is cors config.
type config struct {
	cors.Options
	logger log.Logger
}

type corsInstaller struct {
	cors *cors.Cors
}

// Name is the external config name.
func (i *corsInstaller) Name() string {
	return ExternalConfigName
}

// Install installs stuffs before server starting.
func (i *corsInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		resp_begin := i.buildCORSPolicies(c)
		err = builder.AddDescriptor(definition.Descriptor{
			Path: "/",
			Middlewares: []definition.Middleware{
				func(ctx context.Context, next definition.Chain) error {
					httpCtx := service.HTTPContextFrom(ctx)

					resp_begin(httpCtx)
					err := next.Continue(ctx)
					return err
				},
			},
		})
	})
	return err
}

type injector func(ctx service.HTTPContext)

func (i *corsInstaller) buildCORSPolicies(c *config) (end_resp injector) {
	logger := c.logger
	printer := func(msg ...interface{}) {
		if logger != nil {
			logger.Infoln(msg...)
		} else {
			log.Infoln(msg...)
		}
	}
	corsHeaderGenerator := func(ctx service.HTTPContext) {
		printer("corsHeaderGenerator: start")
		i.cors = cors.New(c.Options)
		i.cors.HandlerFunc(ctx.ResponseWriter(), ctx.Request())
		printer("corsHeaderGenerator: end")
	}
	return func(ctx service.HTTPContext) {
		corsHeaderGenerator(ctx)
	}
}

// Uninstall uninstalls stuffs after server terminating.
func (i *corsInstaller) Uninstall(service.Builder, *nirvana.Config) error {
	return nil
}

// Disable returns a configurer to disable cors.
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
			c.logger = nil
		})
		return nil
	}
}

// Logger Configurer sets logger.
func Logger(l log.Logger) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.logger = l
		})
		return nil
	}
}

// AllowedOrigins is a list of origins a cross-domain request can be executed from.
// If the special "*" value is present in the list, all origins will be allowed.
// An origin may contain a wildcard (*) to replace 0 or more characters
// (i.e.: http://*.domain.com). Usage of wildcards implies a small performance penalty.
// Only one wildcard can be used per origin.
// Default value is ["*"]
func SetAllowedOrigins(origins []string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.AllowedOrigins = origins
		})
		return nil
	}
}

// AllowOriginFunc is a custom function to validate the origin. It take the origin
// as argument and returns true if allowed or false otherwise. If this option is
// set, the content of AllowedOrigins is ignored.
func SetAllowOriginFunc(f func(origin string) bool) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.AllowOriginFunc = f
		})
		return nil
	}
}

// AllowOriginRequestFunc is a custom function to validate the origin. It takes the HTTP Request object and the origin as
// argument and returns true if allowed or false otherwise. If this option is set, the content of `AllowedOrigins`
// and `AllowOriginFunc` is ignored.
func SetAllowOriginRequestFunc(f func(r *http.Request, origin string) bool) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.AllowOriginRequestFunc = f
		})
		return nil
	}
}

type converter func(string) string

// convert converts a list of string using the passed converter function
func convert(s []string, c converter) []string {
	out := []string{}
	for _, i := range s {
		out = append(out, c(i))
	}
	return out
}

// AllowedMethods is a list of methods the client is allowed to use with
// cross-domain requests. Default value is simple methods (HEAD, GET and POST).
func SetAllowedMethods(methods []string) nirvana.Configurer {

	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			if len(methods) == 0 {
				// Default is spec's "simple" methods
				c.AllowedMethods = []string{http.MethodGet, http.MethodPost, http.MethodHead}
			} else {
				c.AllowedMethods = convert(methods, strings.ToUpper)
			}
		})
		return nil
	}
}

// AllowedHeaders is list of non simple headers the client is allowed to use with
// cross-domain requests.
// If the special "*" value is present in the list, all headers will be allowed.
// Default value is [] but "Origin" is always appended to the list.
func SetAllowedHeaders(headers []string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			if len(headers) == 0 {
				// Use sensible defaults
				c.AllowedHeaders = []string{"Origin", "Accept", "Content-Type", "X-Requested-With"}
			}
			c.AllowedHeaders = headers
		})
		return nil
	}
}

// ExposedHeaders indicates which headers are safe to expose to the API of a CORS
// API specification
func SetExposedHeaders(headers []string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.ExposedHeaders = headers
		})
		return nil
	}
}

// MaxAge indicates how long (in seconds) the results of a preflight request
// can be cached
func SetMaxAge(age int) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.MaxAge = age
		})
		return nil
	}
}

// AllowCredentials indicates whether the request can include user credentials like
// cookies, HTTP authentication or client side SSL certificates.

func SetAllowCredentials(isAllow bool) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.AllowCredentials = isAllow
		})
		return nil
	}
}

// OptionsPassthrough instructs preflight to let other potential next handlers to
// process the OPTIONS method. Turn this on if your application handles OPTIONS.
func SetOptionsPassthrough(isPass bool) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.OptionsPassthrough = isPass
		})
		return nil
	}
}

// Debugging flag adds additional output to debug server side CORS issues
func SetDebug(isDebug bool) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.Debug = isDebug
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
			Options: cors.Options{
				AllowedOrigins: []string{"*"},
			},
		}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

type Option struct {
	AllowedOrigins         []string
	AllowOriginFunc        func(origin string) bool
	AllowOriginRequestFunc func(r *http.Request, origin string) bool
	AllowedMethods         []string
	AllowedHeaders         []string
	ExposedHeaders         []string
	MaxAge                 int
	AllowCredentials       bool
	OptionsPassthrough     bool
	Debug                  bool
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		AllowedOrigins: []string{"*"},
	}
}

func NewOption(opt Option) *Option {
	return &opt
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	cfg.Configure(
		SetAllowedOrigins(p.AllowedOrigins),
		SetAllowOriginFunc(p.AllowOriginFunc),
		SetAllowOriginRequestFunc(p.AllowOriginRequestFunc),
		SetAllowedMethods(p.AllowedMethods),
		SetAllowedHeaders(p.AllowedHeaders),
		SetExposedHeaders(p.ExposedHeaders),
		SetMaxAge(p.MaxAge),
		SetAllowCredentials(p.AllowCredentials),
		SetOptionsPassthrough(p.OptionsPassthrough),
		SetDebug(p.Debug),
	)
	return nil
}
