/*
Copyright 2018 Caicloud Authors

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

package apidocs

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

func init() {
	// Register apidocs config installer into nirvana.
	nirvana.RegisterConfigInstaller(&apidocsInstaller{})
}

// ExternalConfigName is the external config name of apidocs.
const ExternalConfigName = "apidocs"

type apidocsInstaller struct{}

// Name is the external config name.
func (i *apidocsInstaller) Name() string {
	return ExternalConfigName
}

type config struct {
	path  string
	files map[string]string
}

// Install installs stuffs before server starting.
func (i *apidocsInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		var descriptors []definition.Descriptor
		var routerPaths []string

		for path, filePath := range c.files {
			b, err := ioutil.ReadFile(filePath)
			if err != nil {
				panic(err)
			}

			router := strings.Join([]string{strings.TrimRight(c.path, "/"), strings.Trim(path, "/")}, "/")
			routerPaths = append(routerPaths, router)

			ds := definition.Descriptor{
				Path:     path,
				Consumes: []string{definition.MIMEAll},
				Produces: []string{definition.MIMEJSON},
				Definitions: []definition.Definition{{
					Method:  definition.Get,
					Results: definition.DataErrorResults(""),
					Function: func(ctx context.Context) ([]byte, error) {
						return b, nil
					}},
				}}

			descriptors = append(descriptors, []definition.Descriptor{ds}...)
		}

		err = builder.AddDescriptor(definition.Descriptor{
			Path:     c.path,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Children: descriptors,
			Definitions: []definition.Definition{{
				Method:  definition.List,
				Results: definition.DataErrorResults(""),
				Function: func(ctx context.Context) ([]string, error) {
					return routerPaths, nil
				},
			}},
		})
	})
	return err
}

// Uninstall uninstalls stuffs after server terminating.
func (i *apidocsInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	return nil
}

// Disable returns a configurer to disable current plugin for a certain nirvana server.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		// Set to nil will delete plugin config from nirvana config.
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

// Path returns a configurer to set apidocs path.
func Path(path string) nirvana.Configurer {
	if path == "" {
		path = "/apidocs"
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.path = path
		})
		return nil
	}
}

// Files Configurer sets apidocs files config.
func Files(files map[string]string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.files = files
		})
		return nil
	}
}

func wrapper(c *nirvana.Config, f func(c *config)) {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// Option contains basic configurations of apidocs.
type Option struct {
	Path  string `desc:"Path to list information of all API docs"`
	Files string `desc:"Comma separated of apidocsVersion:apidocsPath, it can be v1:./apis/api.v1.json,v2:./apis/api.v2.json"`
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		Path:  "/apidocs",
		Files: "",
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	filesSet := strings.Split(p.Files, ",")
	filesMap := make(map[string]string, len(filesSet))

	for _, s := range filesSet {
		if s == "" {
			continue
		}

		if !strings.Contains(s, ":") {
			panic(fmt.Errorf("Please specify the apidocs config with apidocsVersion:apidocsPath, it can be v1:./apis/api.v1.json,v2:./apis/api.v2.json"))
		}

		set := strings.Split(s, ":")
		if len(set) != 2 {
			panic(fmt.Errorf("Please specify the apidocs config with apidocsVersion:apidocsPath, it can be v1:./apis/api.v1.json,v2:./apis/api.v2.json"))
		}

		filePath := strings.TrimSpace(set[1])
		if _, err := os.Stat(filePath); err != nil {
			if os.IsNotExist(err) {
				panic(err)
			}
		}

		filesMap[strings.TrimSpace(set[0])] = filePath
	}

	cfg.Configure(
		Path(p.Path),
		Files(filesMap),
	)
	return nil
}
