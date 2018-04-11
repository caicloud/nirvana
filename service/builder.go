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

package service

import (
	"context"
	"net/http"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service/router"
)

// Builder builds service.
type Builder interface {
	// SetLogger sets logger to server.
	SetLogger(logger log.Logger)
	// SetModifier sets definition modifier.
	SetModifier(m DefinitionModifier)
	// AddFilters add filters to filter requests.
	AddFilter(filters ...Filter)
	// AddDescriptors adds descriptors to router.
	AddDescriptor(descriptors ...definition.Descriptor) error
	// Build builds a service to handle request.
	Build() (Service, error)
}

type binding struct {
	middlewares []router.Middleware
	definitions []definition.Definition
}

type builder struct {
	bindings map[string]*binding
	modifier DefinitionModifier
	filters  []Filter
	logger   log.Logger
}

// NewBuilder creates a service builder.
func NewBuilder() Builder {
	return &builder{
		bindings: make(map[string]*binding),
		logger:   &log.SilentLogger{},
	}
}

// AddFilters add filters to filter requests.
func (b *builder) AddFilter(filters ...Filter) {
	b.filters = append(b.filters, filters...)
}

// SetLogger sets logger to server.
func (b *builder) SetLogger(logger log.Logger) {
	if logger != nil {
		b.logger = logger
	} else {
		b.logger = &log.SilentLogger{}
	}
}

// SetModifier sets definition modifier.
func (b *builder) SetModifier(m DefinitionModifier) {
	b.modifier = m
}

// AddDescriptor adds descriptors to router.
func (b *builder) AddDescriptor(descriptors ...definition.Descriptor) error {
	for _, descriptor := range descriptors {
		b.addDescriptor("", nil, nil, descriptor)
	}
	return nil
}

func (b *builder) addDescriptor(prefix string, consumes []string, produces []string, descriptor definition.Descriptor) {
	path := strings.Join([]string{prefix, strings.Trim(descriptor.Path, "/")}, "/")
	if descriptor.Consumes != nil {
		consumes = descriptor.Consumes
	}
	if descriptor.Produces != nil {
		produces = descriptor.Produces
	}
	if len(descriptor.Middlewares) > 0 || len(descriptor.Definitions) > 0 {
		bd, ok := b.bindings[path]
		if !ok {
			bd = &binding{}
			b.bindings[path] = bd
		}
		if len(descriptor.Middlewares) > 0 {
			for _, m := range descriptor.Middlewares {
				func(m definition.Middleware) {
					bd.middlewares = append(bd.middlewares, func(ctx context.Context, chain router.RoutingChain) error {
						return m(ctx, chain)
					})
				}(m)
			}
		}
		if len(descriptor.Definitions) > 0 {
			for _, d := range descriptor.Definitions {
				bd.definitions = append(bd.definitions, *b.copyDefinition(&d, consumes, produces))
			}
		}
	}
	for _, child := range descriptor.Children {
		b.addDescriptor(strings.TrimRight(path, "/"), consumes, produces, child)
	}
}

func (b *builder) copyDefinition(d *definition.Definition, consumes []string, produces []string) *definition.Definition {
	// It copy fields except document.
	newOne := &definition.Definition{}
	*newOne = *d
	if d.Consumes != nil {
		newOne.Consumes = make([]string, len(d.Consumes))
		copy(newOne.Consumes, d.Consumes)
	} else if consumes != nil {
		newOne.Consumes = make([]string, len(consumes))
		copy(newOne.Consumes, consumes)
	}
	if d.Produces != nil {
		newOne.Produces = make([]string, len(d.Produces))
		copy(newOne.Produces, d.Produces)
	} else if produces != nil {
		newOne.Produces = make([]string, len(produces))
		copy(newOne.Produces, produces)
	}
	newOne.Parameters = make([]definition.Parameter, len(d.Parameters))
	for i, p := range d.Parameters {
		p.Description = ""
		newOne.Parameters[i] = p
	}
	newOne.Results = make([]definition.Result, len(d.Results))
	for i, r := range d.Results {
		r.Description = ""
		newOne.Results[i] = r
	}
	if len(newOne.ErrorProduces) <= 0 {
		newOne.ErrorProduces = newOne.Produces
	}
	return newOne
}

// Build builds a service to handle request.
func (b *builder) Build() (Service, error) {
	if len(b.bindings) <= 0 {
		return nil, noRouter.Error()
	}
	var root router.Router
	for path, bd := range b.bindings {
		b.logger.V(log.LevelDebug).Infof("Definitions: %d Middlewares: %d Path: %s",
			len(bd.definitions), len(bd.middlewares), path)
		router, leaf, err := router.Parse(path)
		if err != nil {
			b.logger.Errorf("Can't parse path: %s, %s", path, err.Error())
			return nil, err
		}
		if len(bd.definitions) > 0 {
			// RedirectTrailingSlash would redirect "/somepath/" to "/somepath". Any definition under "/somepath/"
			// will never be executed.
			if len(path) > 1 && strings.HasSuffix(path, "/") {
				b.logger.Warningf("If RedirectTrailingSlash filter is enabled, following %d definition(s) would not be executed", len(bd.definitions))
			}
			inspector := newInspector(path, b.logger)
			for _, d := range bd.definitions {
				b.logger.V(log.LevelDebug).Infof("  Method: %s Consumes: %v Produces: %v",
					d.Method, d.Consumes, d.Produces)
				if b.modifier != nil {
					b.modifier(&d)
				}
				if err := inspector.addDefinition(d); err != nil {
					return nil, err
				}
			}

			leaf.SetInspector(inspector)
		}
		if len(bd.middlewares) > 0 {
			leaf.AddMiddleware(bd.middlewares...)
		}
		if root == nil {
			root = router
		} else {
			root, err = root.Merge(router)
			if err != nil {
				return nil, err
			}
		}
	}
	s := &service{
		root:      root,
		filters:   b.filters,
		logger:    b.logger,
		producers: AllProducers(),
	}
	return s, nil
}

// Service handles HTTP requests.
//
// Workflow:
//            Service.ServeHTTP()
//          ----------------------
//          ↓                    ↑
// |-----Filters------|          ↑
//          ↓                    ↑
// |---Router Match---|          ↑
//          ↓                    ↑
// |-------------Middlewares------------|
//          ↓                    ↑
// |-------------Executor---------------|
//          ↓                    ↑
// |-ParameterGenerators-|-DestinationHandlers-|
//          ↓                    ↑
// |------------User Function-----------|
type Service interface {
	http.Handler
}

type service struct {
	root      router.Router
	filters   []Filter
	logger    log.Logger
	producers []Producer
}

func (s *service) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	for _, f := range s.filters {
		if !f(resp, req) {
			return
		}
	}
	ctx := newHTTPContext(resp, req)

	executor, err := s.root.Match(ctx, &ctx.container, req.URL.Path)
	if err != nil {
		if err := writeError(ctx, s.producers, err); err != nil {
			s.logger.Error(err)
		}
		return
	}
	err = executor.Execute(ctx)
	if err == nil && ctx.response.HeaderWritable() {
		err = invalidService.Error()
	}
	if err != nil {
		if ctx.response.HeaderWritable() {
			if err := writeError(ctx, s.producers, err); err != nil {
				s.logger.Error(err)
			}
		} else {
			s.logger.Error(err)
		}
	}
}
