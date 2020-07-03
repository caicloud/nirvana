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

package project

import (
	"bytes"
	"fmt"
	"go/build"
	"go/format"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/utils/project"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func newInitCommand() *cobra.Command {
	options := &initOptions{}
	cmd := &cobra.Command{
		Use:   "init /path/to/project",
		Short: "Create a basic project structure",
		Long:  options.Manuals(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Validate(cmd, args); err != nil {
				log.Fatalln(err)
			}
			if err := options.Run(cmd, args); err != nil {
				log.Fatalln(err)
			}
		},
	}
	options.Install(cmd.PersistentFlags())
	return cmd
}

type templateData struct {
	Boilerplate    string
	ProjectAbsDir  string
	ProjectPackage string
	ProjectName    string
	ImagePrefix    string
	ImageSuffix    string
	Registry       string
	BaseRegistry   string
}

// GoBoilerplate returns boilerplate in go style.
func (t *templateData) GoBoilerplate() string {
	return "/*\n" + t.Boilerplate + "\n*/\n"
}

// SharpBoilerplate returns boilerplate in sharp style.
func (t *templateData) SharpBoilerplate() string {
	return "# " + strings.Replace(t.Boilerplate, "\n", "\n# ", -1)
}

// SlashBoilerplate returns boilerplate in slash style.
func (t *templateData) SlashBoilerplate() string {
	return "// " + strings.Replace(t.Boilerplate, "\n", "\n// ", -1)
}

type initOptions struct {
	Boilerplate  string
	ImagePrefix  string
	ImageSuffix  string
	Registry     string
	BaseRegistry string
}

func (o *initOptions) Install(flags *pflag.FlagSet) {
	flags.StringVar(&o.Boilerplate, "boilerplate", "", "Path to boilerplate")
	flags.StringVar(&o.ImagePrefix, "image-prefix", "", "Docker image prefix")
	flags.StringVar(&o.ImageSuffix, "image-suffix", "", "Docker image suffix")
	flags.StringVar(&o.Registry, "registry", "", "Container registry")
	flags.StringVar(&o.BaseRegistry, "base-registry", "", "Container registry for base images")
}

func (o *initOptions) Validate(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("must specify a project path")
	}
	if len(args) > 1 {
		return fmt.Errorf("must not specify multiple project paths")
	}
	return nil
}

func (o *initOptions) Run(cmd *cobra.Command, args []string) error {
	pathToProject, err := filepath.Abs(args[0])
	if err != nil {
		return fmt.Errorf("can't get absolute path for %s: %v", args[0], err)
	}
	projectName := filepath.Base(pathToProject)

	td := &templateData{
		ProjectAbsDir: filepath.Dir(pathToProject),
		ProjectName:   projectName,
		ImagePrefix:   o.ImagePrefix,
		ImageSuffix:   o.ImageSuffix,
		Registry:      o.Registry,
		BaseRegistry:  o.BaseRegistry,
	}
	td.ProjectPackage, err = project.PackageForPath(pathToProject)
	if err != nil {
		return err
	}

	if td.ProjectPackage == "" {
		return fmt.Errorf("project %s is not in GOPATH %s", pathToProject, build.Default.GOPATH)
	}

	if o.Boilerplate != "" {
		data, err := ioutil.ReadFile(o.Boilerplate)
		if err != nil {
			return fmt.Errorf("can't read boilerplate file %s: %v", o.Boilerplate, err)
		}
		data = bytes.Replace(data, []byte("YEAR"), []byte(strconv.Itoa(time.Now().Year())), -1)
		data = bytes.TrimSpace(data)
		td.Boilerplate = string(data)
	}

	directories := o.directories(projectName)
	for i, dir := range directories {
		dir = filepath.Join(pathToProject, dir)
		if _, err = os.Stat(dir); !os.IsNotExist(err) {
			if err != nil {
				return fmt.Errorf("can't get stat for %s: %v", dir, err)
			}
			return fmt.Errorf("%s already exists", dir)
		}
		directories[i] = dir
	}
	for _, dir := range directories {
		if err = os.MkdirAll(dir, 0775); err != nil {
			return fmt.Errorf("can't create directory %s: %v", dir, err)
		}
	}

	files := map[string][]byte{}
	for file, tpl := range o.templates(projectName) {
		tmpl, err := template.New(file).Parse(tpl)
		if err != nil {
			return fmt.Errorf("can't create template %s: %v", file, err)
		}
		buf := bytes.NewBuffer(nil)
		if err = tmpl.Execute(buf, td); err != nil {
			return fmt.Errorf("can't execute template %s: %v", file, err)
		}
		if strings.HasSuffix(file, ".go") {
			files[file], err = format.Source(buf.Bytes())
			if err != nil {
				return fmt.Errorf("can't format go source file %s: %v", file, err)
			}
		} else {
			files[file] = buf.Bytes()
		}
	}
	for file, data := range files {
		file = filepath.Join(pathToProject, file)
		if err = ioutil.WriteFile(file, data, 0664); err != nil {
			return fmt.Errorf("can't write file %s: %v", file, err)
		}
	}
	log.Infof("Created project at %s", pathToProject)
	return nil
}

func (o *initOptions) directories(project string) []string {
	return []string{
		"apis",
		"bin",
		fmt.Sprintf("cmd/%s", project),
		fmt.Sprintf("build/%s", project),
		"docs",
		"hack",
		"pkg/apis/v1/converters",
		"pkg/apis/v1/descriptors",
		"pkg/filters",
		"pkg/handlers",
		"pkg/middlewares",
		"pkg/modifiers",
		"pkg/version",
		"test",
		"vendor",
	}
}

func (o *initOptions) templates(proj string) map[string]string {
	return map[string]string{
		fmt.Sprintf("cmd/%s/main.go", proj):      o.templateMain(),
		fmt.Sprintf("build/%s/Dockerfile", proj): o.templateDockerfile(),
		"docs/README.md":                         o.templateDocsREADME(),
		"hack/README.md":                         o.templateHackREADME(),
		"hack/read_cpus_available.sh":            o.templateHackCPUScript(),
		"hack/script.sh":                         o.templateHackScript(),
		"hack/tools.go":                          o.templateHackTools(),
		"pkg/apis/descriptors.go":                o.templateDescriptors(),
		"pkg/apis/v1/types.go":                   o.templateAPITypes(),
		"pkg/apis/v1/converters/converters.go":   o.templateConverters(),
		"pkg/apis/v1/descriptors/descriptors.go": o.templateDescriptorsV1(),
		"pkg/apis/v1/descriptors/message.go":     o.templateMessageDescriptors(),
		"pkg/filters/filters.go":                 o.templateFilters(),
		"pkg/handlers/message.go":                o.templateMessageHandler(),
		"pkg/middlewares/middlewares.go":         o.templateMiddlewares(),
		"pkg/modifiers/modifiers.go":             o.templateModifiers(),
		"pkg/version/version.go":                 o.templateVersion(),
		"test/test_make.sh":                      o.templateTestScript(),
		".golangci.yml":                          o.templateGolangCI(),
		"go.mod":                                 o.templateGomod(),
		"Makefile":                               o.templateMakefile(),
		project.DefaultProjectFileName:           o.templateProject(),
		"README.md":                              o.templateReadme(),
	}
}

func (o *initOptions) templateMain() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package main

import (
	"fmt"

	"{{ .ProjectPackage }}/pkg/apis"
	"{{ .ProjectPackage }}/pkg/filters"
	"{{ .ProjectPackage }}/pkg/modifiers"
	"{{ .ProjectPackage }}/pkg/version"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/reqlog"
	pversion "github.com/caicloud/nirvana/plugins/version"
)

func main() {
	// Print nirvana banner.
	fmt.Println(nirvana.Logo, nirvana.Banner)

	// Create nirvana command.
	cmd := config.NewNamedNirvanaCommand("server", config.NewDefaultOption())

	// Create plugin options.
	metricsOption := metrics.NewDefaultOption() // Metrics plugin.
	reqlogOption := reqlog.NewDefaultOption()   // Request log plugin.
	versionOption := pversion.NewOption(        // Version plugin.
		"{{ .ProjectName }}",
		version.Version,
		version.Commit,
		version.Package,
	)

	// Enable plugins.
	cmd.EnablePlugin(metricsOption, reqlogOption, versionOption)

	// Create server config.
	serverConfig := nirvana.NewConfig()

	// Configure APIs. These configurations may be changed by plugins.
	serverConfig.Configure(
		nirvana.Logger(log.DefaultLogger()),
		nirvana.Filter(filters.Filters()...),
		nirvana.Modifier(modifiers.Modifiers()...),
		nirvana.Descriptor(apis.Descriptor()),
	)

	// Set nirvana command hooks.
	cmd.SetHook(&config.NirvanaCommandHookFunc{
		PreServeFunc: func(config *nirvana.Config, server nirvana.Server) error {
			// Output project information.
			config.Logger().Infof("Package:%s Version:%s Commit:%s", version.Package, version.Version, version.Commit)
			return nil
		},
	})

	// Start with server config.
	if err := cmd.ExecuteWithConfig(serverConfig); err != nil {
		serverConfig.Logger().Fatal(err)
	}
}
`
}

func (o *initOptions) templateDockerfile() string {
	return `
{{- if .Boilerplate -}}
{{ .SharpBoilerplate }}

{{ end -}}
FROM {{ if .BaseRegistry }}{{ .BaseRegistry }}/debian:stretch{{ else }}debian:stretch{{ end }}

ADD bin/{{ .ProjectName }} /usr/local/bin

ENTRYPOINT ["{{ .ProjectName }}"]
`
}

func (o *initOptions) templateDocsREADME() string {
	return `<!-- DOCTOC SKIP -->

# docs
`
}

func (o *initOptions) templateHackREADME() string {
	return `<!-- DOCTOC SKIP -->

Scripts used to manage this repository.
`
}

func (o *initOptions) templateHackCPUScript() string {
	return `#!/bin/bash

set -e

if [ -z "$MAX_CPUS" ]; then
    MAX_CPUS=1

    case "$(uname -s)" in
    Darwin)
        MAX_CPUS=$(sysctl -n machdep.cpu.core_count)
        ;;
    Linux)
        CFS_QUOTA=$(cat /sys/fs/cgroup/cpu/cpu.cfs_quota_us)
        if [ "$CFS_QUOTA" -ge 100000 ]; then
            MAX_CPUS=$(("$CFS_QUOTA" / 100 / 1000))
        fi
        ;;
    *)
        # Unsupported host OS. Must be Linux or Mac OS X.
        ;;
    esac
fi

echo "$MAX_CPUS"
`
}

func (o *initOptions) templateHackScript() string {
	return `#!/bin/bash

set -e

echo "Hello world"
`
}

func (o *initOptions) templateHackTools() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

// +build tools

// This package imports things required by build scripts, to force` + " `go mod` " + `to see them as dependencies
package tools

import _ "github.com/caicloud/nirvana/utils/api"
`
}

func (o *initOptions) templateAPITypes() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package v1

// Message describes a message entry.
type Message struct {
	ID      int    ` + "`json:\"id\"`" + `
	Title   string ` + "`json:\"title\"`" + `
	Content string ` + "`json:\"content\"`" + `
}
`
}

func (o *initOptions) templateConverters() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package converters
`
}

func (o *initOptions) templateDescriptors() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

// +nirvana:api=descriptors:"Descriptor"

package apis

import (
	descriptorsv1 "{{ .ProjectPackage }}/pkg/apis/v1/descriptors"
	"{{ .ProjectPackage }}/pkg/middlewares"

	def "github.com/caicloud/nirvana/definition"
)

// Descriptor returns a combined descriptor for APIs of all versions.
func Descriptor() def.Descriptor {
	return def.Descriptor{
		Description: "APIs",
		Path:        "/apis",
		Middlewares: middlewares.Middlewares(),
		Consumes:    []string{def.MIMEJSON},
		Produces:    []string{def.MIMEJSON},
		Children: []def.Descriptor{
			descriptorsv1.Descriptor(),
		},
	}
}
`
}

func (o *initOptions) templateDescriptorsV1() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package descriptors

import (
	def "github.com/caicloud/nirvana/definition"
)

// descriptors describe APIs of current version.
var descriptors []def.Descriptor

// register registers descriptors.
func register(ds ...def.Descriptor) {
	descriptors = append(descriptors, ds...)
}

// Descriptor returns a combined descriptor for current version.
func Descriptor() def.Descriptor {
	return def.Descriptor{
		Description: "v1 APIs",
		Path:        "/v1",
		Children:    descriptors,
	}
}
`
}

func (o *initOptions) templateMessageDescriptors() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package descriptors

import (
	"{{ .ProjectPackage }}/pkg/handlers"

	def "github.com/caicloud/nirvana/definition"
)

func init() {
	register([]def.Descriptor{{ print "{{" }}
		Path:        "/messages",
		Definitions: []def.Definition{listMessages},
	}, {
		Path:        "/messages/{message}",
		Definitions: []def.Definition{getMessage},
	},
	}...)
}

var listMessages = def.Definition{
	Method:   def.List,
	Summary: "List Messages",
	Description: "Query a specified number of messages and returns an array",
	Function: handlers.ListMessages,
	Parameters: []def.Parameter{
		{
			Source:      def.Query,
			Name:        "count",
			Default:     10,
			Description: "Number of messages",
		},
	},
	Results: def.DataErrorResults("A list of messages"),
}

var getMessage = def.Definition{
	Method:   def.Get,
	Summary: "Get Message",
	Description: "Get a message by id",
	Function: handlers.GetMessage,
	Parameters: []def.Parameter{
		def.PathParameterFor("message", "Message id"),
	},
	Results: def.DataErrorResults("A message"),
}
`
}

func (o *initOptions) templateFilters() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package filters

import "github.com/caicloud/nirvana/service"

// Filters returns a list of filters.
func Filters() []service.Filter {
	return []service.Filter{
		service.RedirectTrailingSlash(),
		service.FillLeadingSlash(),
		service.ParseRequestForm(),
	}
}
`
}

func (o *initOptions) templateMessageHandler() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package handlers

import (
	"context"
	"fmt"

	v1 "{{ .ProjectPackage }}/pkg/apis/v1"
)

// ListMessages returns all messages.
func ListMessages(ctx context.Context, count int) ([]v1.Message, error) {
	messages := make([]v1.Message, count)
	for i := 0; i < count; i++ {
		messages[i].ID = i
		messages[i].Title = fmt.Sprintf("Example %d", i)
		messages[i].Content = fmt.Sprintf("Content of example %d", i)
	}
	return messages, nil
}

// GetMessage return a message by id.
func GetMessage(ctx context.Context, id int) (*v1.Message, error) {
	return &v1.Message{
		ID:      id,
		Title:   "This is an example",
		Content: "Example content",
	}, nil
}
`
}

func (o *initOptions) templateMiddlewares() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package middlewares

import def "github.com/caicloud/nirvana/definition"

// Middlewares returns a list of middlewares.
func Middlewares() []def.Middleware {
	return []def.Middleware{}
}
`
}

func (o *initOptions) templateModifiers() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

// +nirvana:api=modifiers:"Modifiers"

package modifiers

import "github.com/caicloud/nirvana/service"

// Modifiers returns a list of modifiers.
func Modifiers() []service.DefinitionModifier {
	return []service.DefinitionModifier{
		service.FirstContextParameter(),
		service.ConsumeAllIfConsumesIsEmpty(),
		service.ProduceAllIfProducesIsEmpty(),
		service.ConsumeNoneForHTTPGet(),
		service.ConsumeNoneForHTTPDelete(),
		service.ProduceNoneForHTTPDelete(),
	}
}
`
}

func (o *initOptions) templateVersion() string {
	return `
{{- if .Boilerplate -}}
{{ .GoBoilerplate }}
{{- end }}

package version

// Following values should be substituted with a real value during build.
var (
	Version = "Unknown"
	Commit = "Unknown"
	Package = "{{ .ProjectPackage }}"
)
`
}

func (o *initOptions) templateTestScript() string {
	return `#!/bin/bash

set -e

ROOT=$(dirname "${BASH_SOURCE}")/..

function test_make() {
  cd $ROOT
  make lint
  make test
  make build
  make build-linux
  make container
  make clean
  cd ..
}

test_make
`
}

func (o *initOptions) templateGolangCI() string {
	return `run:
  # concurrency: 2
  deadline: 5m

linter-settings:
  goconst:
    min-len: 2
    min-occurrences: 2

linters:
  enable:
    - golint
    - goconst
    - gofmt
    - goimports
    - misspell
    - unparam

issues:
  exclude-use-default: false
  exclude-rules:
    - path: _test.go
      linters:
        - errcheck
  exclude:
    - (comment on exported (method|function|type|const)|should have( a package)? comment|comment should be of the form)
`
}

func (o *initOptions) templateGomod() string {
	return `
{{- if .Boilerplate -}}
{{ .SlashBoilerplate }}
//
{{ end -}}
// go.mod example
//
// Refer to https://github.com/golang/go/wiki/Modules#gomod
// for detailed go.mod and go mod command documentation.
//
// module github.com/my/module/v3
//
// require (
//     github.com/some/dependency v1.2.3
//     github.com/another/dependency v0.1.0
//     github.com/additional/dependency/v4 v4.0.0
// )

module {{ .ProjectPackage }}

require (
	github.com/caicloud/nirvana master
)
`
}

func (o *initOptions) templateMakefile() string {
	return `
{{- if .Boilerplate -}}
{{ .SharpBoilerplate }}
#
{{ end -}}
# The old school Makefile, following are required targets. The Makefile is written
# to allow building multiple binaries. You are free to add more targets or change
# existing implementations, as long as the semantics are preserved.
#
#   make              - default to 'build' target
#   make lint         - code analysis
#   make test         - run unit test (or plus integration test)
#   make build        - alias to build-local target
#   make build-local  - build local binary targets
#   make build-linux  - build linux binary targets
#   make container    - build containers
#   $ docker login registry -u username -p xxxxx
#   make push         - push containers
#   make clean        - clean up targets
#
# Not included but recommended targets:
#   make e2e-test
#
# The makefile is also responsible to populate project version information.
#

#
# Tweak the variables based on your project.
#

# This repo's root import path (under GOPATH).
ROOT := {{ .ProjectPackage }}

# Target binaries. You can build multiple binaries for a single project.
TARGETS := {{ .ProjectName }}

# Container image prefix and suffix added to targets.
# The final built images are:
#   $[REGISTRY]/$[IMAGE_PREFIX]$[TARGET]$[IMAGE_SUFFIX]:$[VERSION]
# $[REGISTRY] is an item from $[REGISTRIES], $[TARGET] is an item from $[TARGETS].
IMAGE_PREFIX ?= $(strip {{ .ImagePrefix }})
IMAGE_SUFFIX ?= $(strip {{ .ImageSuffix }})

# Container registries.
REGISTRY ?= {{ .Registry }}

# Container registry for base images.
BASE_REGISTRY ?= {{ .BaseRegistry }}

#
# These variables should not need tweaking.
#

# It's necessary to set this because some environments don't link sh -> bash.
export SHELL := /bin/bash

# It's necessary to set the errexit flags for the bash shell.
export SHELLOPTS := errexit

# Project main package location (can be multiple ones).
CMD_DIR := ./cmd

# Project output directory.
OUTPUT_DIR := ./bin

# Build direcotory.
BUILD_DIR := ./build

# Current version of the project.
VERSION ?= $(shell git describe --tags --always --dirty)

# Available cpus for compiling, please refer to https://github.com/caicloud/engineering/issues/8186#issuecomment-518656946 for more information.
CPUS ?= $(shell /bin/bash hack/read_cpus_available.sh)

# Track code version with Docker Label.
DOCKER_LABELS ?= git-describe="$(shell date -u +v%Y%m%d)-$(shell git describe --tags --always --dirty)"

# Golang standard bin directory.
GOPATH ?= $(shell go env GOPATH)
BIN_DIR := $(GOPATH)/bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

# Default golang flags used in build and test
# -mod=vendor: force go to use the vendor files instead of using the` + " `$GOPATH/pkg/mod` " + `
# -p: the number of programs that can be run in parallel
# -count: run each test and benchmark 1 times. Set this flag to disable test cache
export GOFLAGS ?= -mod=vendor -p=$(CPUS) -count=1

#
# Define all targets. At least the following commands are required:
#

# All targets.
.PHONY: lint test build container push

build: build-local

# more info about` + " `GOGC` " + `env: https://github.com/golangci/golangci-lint#memory-usage-of-golangci-lint
lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run

$(GOLANGCI_LINT):
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(BIN_DIR) v1.23.6

test:
	@go test -race -coverprofile=coverage.out ./...
	@go tool cover -func coverage.out | tail -n 1 | awk '{ print "Total coverage: " $$3 }'

build-local:
	@for target in $(TARGETS); do                                                      \
	  go build -v -o $(OUTPUT_DIR)/$${target}                                          \
	    -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)                      \
	      -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                                     \
	    $(CMD_DIR)/$${target};                                                         \
	done

build-linux:
	@docker run --rm -it                                                               \
	  -v $(PWD):/go/src/$(ROOT)                                                        \
	  -w /go/src/$(ROOT)                                                               \
	  -e GOOS=linux                                                                    \
	  -e GOARCH=amd64                                                                  \
	  -e GOPATH=/go                                                                    \
	  -e GOFLAGS="$(GOFLAGS)"                                                          \
	  -e SHELLOPTS="$(SHELLOPTS)"                                                      \
{{- if .BaseRegistry }}
	  $(BASE_REGISTRY)/golang:1.13.9-stretch                                           \
{{- else }}
	  golang:1.13.9-stretch                                                            \
{{- end }}
	    /bin/bash -c 'for target in $(TARGETS); do                                     \
	      go build -v -o $(OUTPUT_DIR)/$${target}                                      \
	        -ldflags "-s -w -X $(ROOT)/pkg/version.VERSION=$(VERSION)                  \
	          -X $(ROOT)/pkg/version.REPOROOT=$(ROOT)"                                 \
	        $(CMD_DIR)/$${target};                                                     \
	    done'

container: build-linux
	@for target in $(TARGETS); do                                                      \
	  image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                  \
{{- if .Registry }}
	  docker build -t $(REGISTRY)/$${image}:$(VERSION)                                 \
{{- else }}
	  docker build -t $${image}:$(VERSION)                                             \
{{- end }}
	    --label $(DOCKER_LABELS)                                                       \
	    -f $(BUILD_DIR)/$${target}/Dockerfile .;                                       \
	done

push: container
	@for target in $(TARGETS); do                                                      \
	  image=$(IMAGE_PREFIX)$${target}$(IMAGE_SUFFIX);                                  \
{{- if .Registry }}
	  docker push $(REGISTRY)/$${image}:$(VERSION);                                    \
{{- else }}
	  docker push $${image}:$(VERSION);                                                \
{{- end }}
	done

.PHONY: clean
clean:
	@-rm -vrf ${OUTPUT_DIR}
`
}

func (o *initOptions) templateProject() string {
	return `
{{- if .Boilerplate -}}
{{ .SharpBoilerplate }}
#
{{ end -}}
# This file describes your project. It's used to generate api docs and
# clients. All fields in this file won't affect nirvana configurations.

project: {{ .ProjectName }}
description: This project uses nirvana as API framework
schemes:
- http
hosts:
- localhost:8080
contacts:
- name: nobody
  email: nobody@nobody.io
  description: Maintain this project
versions:
- name: v1
  description: The v1 version is the first version of this project
  rules:
  - prefix: /apis/v1/
`
}

func (o *initOptions) templateReadme() string {
	return `# Project {{ .ProjectName }}

<!-- Write one paragraph of this project description here -->

## Getting Started

### Prerequisites

<!-- Describe packages, tools and everything we needed here -->

### Building

<!-- Describe how to build this project -->

### Running

<!-- Describe how to run this project -->

## Versioning

<!-- Place versions of this project and write comments for every version -->

## Contributing

<!-- Tell others how to contribute this project -->

## Authors

<!-- Put authors here -->

## License

<!-- A link to license file -->

`
}

func (o *initOptions) Manuals() string {
	return `
This command generates standard nirvana project structure.
.                                   #
├── .golangci.yml                   #
├── go.mod                          #
├── Makefile                        #
├── OWNERS                          #
├── README.md                       #
├── apis                            # Store apidocs (swagger json)
├── bin                             # Store the compiled binary
├── build                           # Store Dockerfile
│   └── demo-admin                  #
│       └── Dockerfile              #
├── cmd                             # Store startup commands for project
│   └── demo-admin                  #
│       └── main.go                 #
├── docs                            # Store docs
│   └── README.md                   #
├── hack                            # Store scripts
│   ├── README.md                   #
│   ├── read_cpus_available.sh      # Script to read available cpus
│   └── script.sh                   #
├── nirvana.yaml                    # File to describes your project
├── pkg                             # Store structures and converters required by API, distinguish by version
│   ├── apis                        #
│   │   ├── descriptors.go          # Store API descriptions (routing and others), distinguish by version
│   │   └── v1                      #
│   │       ├── converters          #
│   │       │   └── converters.go   #
│   │       ├── descriptors         #
│   │       │   ├── descriptors.go  #
│   │       │   └── message.go      # Store API definition of message
│   │       └── types.go            #
│   ├── filters                     # Store HTTP Request filter
│   │   └── filter.go               #
│   ├── handlers                    # Store the logical processing required by APIs
│   │   └── message.go              #
│   ├── middlewares                 # Store middlewares
│   │   └── middlewares.go          #
│   ├── modifiers                   # Store definition modifiers
│   │   └── modifiers.go            #
│   └── version                     # Store version information of project
│       └── version.go              #
├── test                            # Store all tests (except unit tests), e.g. integration, e2e tests.
│   └── test_make.sh                #
└── vendor                          #
`
}
