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
	"io/ioutil"
	"path"
	"strings"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/api"
)

func init() {
	// Register apidocs config installer into nirvana.
	nirvana.RegisterConfigInstaller(&apiDocsInstaller{})
}

// ExternalConfigName is the external config name of apidocs.
const ExternalConfigName = "apidocs"

type apiDocsInstaller struct{}

// Name is the external config name.
func (i *apiDocsInstaller) Name() string {
	return ExternalConfigName
}

type config struct {
	path  string
	files map[string]string
}

// Install installs stuffs before server starting.
func (i *apiDocsInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	return wrapper(cfg, func(c *config) error {
		files := make(map[string][]byte)
		for filename, path := range c.files {
			data, err := ioutil.ReadFile(path)
			if err != nil {
				return err
			}
			files[filename] = data
		}

		apiDescriptor := make([]interface{}, 0, len(files)+1)
		versions := make([]string, 0, len(files))
		for v, data := range files {
			versions = append(versions, v)
			apiDescriptor = append(apiDescriptor, api.DescriptorForData(api.PathForVersion(c.path, v), data, definition.MIMEJSON))
		}
		data, err := api.GenSwaggerPageData(c.path, versions)
		if err != nil {
			return err
		}
		apiDescriptor = append(
			apiDescriptor,
			api.DescriptorForData(c.path, data, definition.MIMEHTML),
		)
		return builder.AddDescriptor(apiDescriptor...)
	})
}

// Uninstall uninstalls stuffs after server terminating.
func (i *apiDocsInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
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
		return wrapper(c, func(c *config) error {
			return nil
		})
	}
}

// Path returns a configurer to set apidocs path.
func Path(path string) nirvana.Configurer {
	if path == "" {
		path = "/docs"
	}
	return func(c *nirvana.Config) error {
		return wrapper(c, func(c *config) error {
			c.path = path
			return nil
		})
	}
}

// Files Configurer sets apidocs files config.
func Files(files map[string]string) nirvana.Configurer {
	return func(c *nirvana.Config) error {
		return wrapper(c, func(c *config) error {
			c.files = files
			return nil
		})
	}
}

func wrapper(c *nirvana.Config, f func(c *config) error) error {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	if err := f(cfg); err != nil {
		return err
	}
	c.Set(ExternalConfigName, cfg)
	return nil
}

// Option contains basic configurations of apidocs.
type Option struct {
	Path      string `desc:"Path to the API documentation page, default: /docs"`
	FilesPath string `desc:"The folder path that contains all swagger JSON files, default: apis"`
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		Path:      "/docs",
		FilesPath: "apis",
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	// ignore errors, do not serve API documentation if files read failed (e.g. the folder does not exist)
	filesInfo, _ := ioutil.ReadDir(p.FilesPath)
	files := make(map[string]string, len(filesInfo))
	for _, f := range filesInfo {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		// api.xxx.json --> xxx
		name := strings.TrimPrefix(f.Name(), "api.")
		name = strings.TrimSuffix(name, ".json")
		files[name] = path.Join(p.FilesPath, f.Name())
	}

	cfg.Configure(
		Path(p.Path),
		Files(files),
	)
	return nil
}
