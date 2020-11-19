/*
Copyright 2020 Caicloud Authors

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

package rpc

import (
	"fmt"
	"net/http"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/service/executor"
)

type binding struct {
	middlewares []definition.Middleware
	definition  definition.Definition
	executor    executor.MiddlewareExecutor
}

type builder struct {
	// bindings contains all RPC action definitions, the key is a unique id (path + version + name),
	// it is currently formatted as an API URL path, eg: /?Version=2020-10-10&Action=Echo, which is useful for both
	// printing logs and generating API documents/client
	bindings map[string]*binding
	modifier service.DefinitionModifier
	filters  []service.Filter
	logger   log.Logger
}

// NewBuilder creates a service builder.
func NewBuilder() service.Builder {
	return &builder{
		bindings: make(map[string]*binding),
		logger:   &log.SilentLogger{},
	}
}

// Filters returns all request filters.
func (b *builder) Filters() []service.Filter {
	result := make([]service.Filter, len(b.filters))
	copy(result, b.filters)
	return result
}

// AddFilter add filters to filter requests.
func (b *builder) AddFilter(filters ...service.Filter) {
	b.filters = append(b.filters, filters...)
}

// Logger returns logger of builder.
func (b *builder) Logger() log.Logger {
	return b.logger
}

// SetLogger sets logger to builder.
func (b *builder) SetLogger(logger log.Logger) {
	if logger != nil {
		b.logger = logger
	} else {
		b.logger = &log.SilentLogger{}
	}
}

// Modifier returns modifier of builder.
func (b *builder) Modifier() service.DefinitionModifier {
	return b.modifier
}

// SetModifier sets definition modifier.
func (b *builder) SetModifier(m service.DefinitionModifier) {
	b.modifier = m
}

func genRPCPath(prefix, version, action string) string {
	return fmt.Sprintf("%s?Version=%s&Action=%s", prefix, version, action)
}

// AddDescriptor adds descriptors to router.
func (b *builder) AddDescriptor(descriptors ...interface{}) error {
	for _, obj := range descriptors {
		descriptor, ok := obj.(definition.RPCDescriptor)
		if !ok {
			return fmt.Errorf("not a Descriptor")
		}
		p := descriptor.Path
		if p == "" {
			p = "/"
		}
		for _, action := range descriptor.Actions {
			b.bindings[genRPCPath(p, action.Version, action.Name)] = &binding{
				middlewares: descriptor.Middlewares,
				definition:  b.genDefinition(action, descriptor.Consumes, descriptor.Produces, descriptor.Tags),
			}
		}
	}
	return nil
}

func (b *builder) genDefinition(action definition.RPCAction, consumes []string, produces []string, tags []string) definition.Definition {
	if len(action.Consumes) > 0 {
		consumes = action.Consumes
	}
	if len(action.Produces) > 0 {
		produces = action.Produces
	}
	if len(action.Tags) > 0 {
		tags = action.Tags
	}
	errorProduces := produces
	if len(action.ErrorProduces) > 0 {
		errorProduces = action.ErrorProduces
	}

	return definition.Definition{
		Method:        definition.Create,
		Consumes:      consumes,
		Produces:      produces,
		Tags:          tags,
		ErrorProduces: errorProduces,
		Function:      action.Function,
		Parameters:    action.Parameters,
		Results:       action.Results,
		Summary:       action.Summary,
		Description:   action.Description,
		Examples:      action.Examples,
	}
}

// Definitions returns all definitions. If a modifier exists, it will be executed.
// All results are copied from original definitions. Modifications can not affect
// original data.
func (b *builder) Definitions() map[string][]definition.Definition {
	result := make(map[string][]definition.Definition)
	for path, bd := range b.bindings {
		d := bd.definition
		if b.modifier != nil {
			b.modifier(&d)
		}
		result[path] = []definition.Definition{d}
	}
	return result
}

// Build builds a service to handle request.
func (b *builder) Build() (service.Service, error) {
	if len(b.bindings) <= 0 {
		return nil, fmt.Errorf("no router")
	}

	var err error
	for path, bd := range b.bindings {
		b.logger.V(log.LevelDebug).Infof("Path: %s, Consumes: %v, Produces: %v", path, bd.definition.Consumes, bd.definition.Produces)
		if b.modifier != nil {
			b.modifier(&bd.definition)
		}
		bd.executor, err = executor.DefinitionToExecutor(path, bd.definition, http.StatusOK)
		if err != nil {
			return nil, err
		}
	}

	s := &server{
		executors: b.bindings,
		filters:   b.filters,
		logger:    b.logger,
		producers: service.AllProducers(),
	}
	return s, nil
}

type server struct {
	executors map[string]*binding
	filters   []service.Filter
	logger    log.Logger
	producers []service.Producer
}

func (s *server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	for _, f := range s.filters {
		if !f(resp, req) {
			return
		}
	}
	ctx := service.NewHTTPContext(resp, req)

	action := req.URL.Query().Get("Action")
	version := req.URL.Query().Get("Version")
	path := genRPCPath(req.URL.Path, version, action)
	e, ok := s.executors[path]
	if !ok {
		if err := service.WriteError(ctx, s.producers, noExecutorForAction.Error(path)); err != nil {
			s.logger.Error(err)
		}
		return
	}

	err := executor.NewMiddlewareExecutor(e.middlewares, e.executor).Execute(ctx)
	if err == nil && ctx.ResponseWriter().HeaderWritable() {
		err = service.InvalidService.Error()
	}
	if err != nil {
		if ctx.ResponseWriter().HeaderWritable() {
			if err := service.WriteError(ctx, s.producers, err); err != nil {
				s.logger.Error(err)
			}
		} else {
			s.logger.Error(err)
		}
	}
}
