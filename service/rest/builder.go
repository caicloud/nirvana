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

package rest

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/service/rest/router"
)

type binding struct {
	middlewares []definition.Middleware
	definitions []definition.Definition
}

type builder struct {
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

// AddDescriptor adds descriptors to router.
func (b *builder) AddDescriptor(descriptors ...interface{}) error {
	for _, obj := range descriptors {
		descriptor, ok := obj.(definition.Descriptor)
		if !ok {
			return fmt.Errorf("%s is not a definition.Descriptor", reflect.TypeOf(obj).String())
		}
		b.addDescriptor("", nil, nil, nil, descriptor)
	}
	return nil
}

func (b *builder) addDescriptor(prefix string, consumes []string, produces []string, tags []string, descriptor definition.Descriptor) {
	path := strings.Join([]string{prefix, strings.Trim(descriptor.Path, "/")}, "/")
	if descriptor.Consumes != nil {
		consumes = descriptor.Consumes
	}
	if descriptor.Produces != nil {
		produces = descriptor.Produces
	}
	if descriptor.Tags != nil {
		tags = descriptor.Tags
	}
	if len(descriptor.Middlewares) > 0 || len(descriptor.Definitions) > 0 {
		bd, ok := b.bindings[path]
		if !ok {
			bd = &binding{}
			b.bindings[path] = bd
		}
		if len(descriptor.Middlewares) > 0 {
			bd.middlewares = append(bd.middlewares, descriptor.Middlewares...)
		}
		if len(descriptor.Definitions) > 0 {
			for _, d := range descriptor.Definitions {
				bd.definitions = append(bd.definitions, *b.copyDefinition(&d, consumes, produces, tags))
			}
		}
	}
	for _, child := range descriptor.Children {
		b.addDescriptor(strings.TrimRight(path, "/"), consumes, produces, tags, child)
	}
}

// copyDefinition creates a copy from original definition. Those fields with type interface{} only have shallow copies.
func (b *builder) copyDefinition(d *definition.Definition, consumes []string, produces []string, tags []string) *definition.Definition {
	newOne := &definition.Definition{
		Method:      d.Method,
		Summary:     d.Summary,
		Function:    d.Function,
		Description: d.Description,
		Example:     d.Example,
	}
	if len(d.Consumes) > 0 {
		consumes = d.Consumes
	}
	newOne.Consumes = make([]string, len(consumes))
	copy(newOne.Consumes, consumes)

	if len(d.Produces) > 0 {
		produces = d.Produces
	}
	newOne.Produces = make([]string, len(produces))
	copy(newOne.Produces, produces)

	if len(d.Tags) > 0 {
		tags = d.Tags
	}
	newOne.Tags = make([]string, len(tags))
	copy(newOne.Tags, tags)

	if len(d.ErrorProduces) > 0 {
		produces = d.ErrorProduces
	}
	newOne.ErrorProduces = make([]string, len(produces))
	copy(newOne.ErrorProduces, produces)

	newOne.Parameters = make([]definition.Parameter, len(d.Parameters))
	for i, p := range d.Parameters {
		newParameter := p
		newParameter.Operators = make([]definition.Operator, len(p.Operators))
		copy(newParameter.Operators, p.Operators)
		newOne.Parameters[i] = newParameter
	}
	newOne.Results = make([]definition.Result, len(d.Results))
	for i, r := range d.Results {
		newResult := r
		newResult.Operators = make([]definition.Operator, len(r.Operators))
		copy(newResult.Operators, r.Operators)
		newOne.Results[i] = newResult
	}
	return newOne
}

// Definitions returns all definitions. If a modifier exists, it will be executed.
// All results are copied from original definitions. Modifications can not affect
// original data.
func (b *builder) Definitions() map[string][]definition.Definition {
	result := make(map[string][]definition.Definition)
	for path, bd := range b.bindings {
		if len(bd.definitions) > 0 {
			definitions := make([]definition.Definition, len(bd.definitions))
			for i, d := range bd.definitions {
				newCopy := b.copyDefinition(&d, nil, nil, nil)
				if b.modifier != nil {
					b.modifier(newCopy)
				}
				definitions[i] = *newCopy
			}
			result[path] = definitions
		}
	}
	return result
}

// APIStyle returns the API style of this builder.
func (b *builder) APIStyle() service.APIStyle {
	return service.APIStyleREST
}

// Build builds a service to handle request.
func (b *builder) Build() (service.Service, error) {
	if len(b.bindings) <= 0 {
		return nil, noRouter.Error()
	}
	var root router.Router
	for path, bd := range b.bindings {
		b.logger.V(log.LevelDebug).Infof("Definitions: %d Middlewares: %d Path: %s",
			len(bd.definitions), len(bd.middlewares), path)
		top, leaf, err := router.Parse(path)
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
			inspector := newInspector(path)
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
		for _, m := range bd.middlewares {
			leaf.AddMiddleware(m)
		}
		if root == nil {
			root = top
		} else if root, err = root.Merge(top); err != nil {
			return nil, err
		}
	}
	s := &server{
		root:      root,
		filters:   b.filters,
		logger:    b.logger,
		producers: service.AllProducers(),
	}
	return s, nil
}

type server struct {
	root      router.Router
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

	executor, err := s.root.Match(ctx, ctx.ValueContainer(), req.URL.EscapedPath())
	if err != nil {
		if err := service.WriteError(ctx, s.producers, err); err != nil {
			s.logger.Error(err)
		}
		return
	}
	err = executor.Execute(ctx)
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
