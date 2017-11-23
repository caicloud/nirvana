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

package web

import (
	"net/http"
	"strings"
	"sync"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/router"
)

// Server handles HTTP requests.
//
// Workflow:
//            Server.ServeHTTP()
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
// |-ParameterGenerators-|-TypeHandlers-|
//          ↓                    ↑
// |------------User Function-----------|
type Server interface {
	// AddDescriptors adds descriptors to router.
	AddDescriptors(descriptors ...definition.Descriptor) error
	// AddFilters add filters to filter requests.
	AddFilters(filters ...Filter)
	// SetLogger sets logger to server.
	SetLogger(logger log.Logger)
	// SetModifier sets definition modifier.
	SetModifier(m DefinitionModifier)
	// ServeHTTP is used to handle request.
	ServeHTTP(resp http.ResponseWriter, req *http.Request)
}

// NewDefaultServer creates a server with default config.
// It uses three filters: RedirectTrailingSlash, ParseRequestForm, FillLeadingSlash.
// It uses one modifier: FirstContextParameter.
func NewDefaultServer() Server {
	s := NewServer()
	s.AddFilters(RedirectTrailingSlash(), ParseRequestForm(), FillLeadingSlash())
	s.SetModifier(FirstContextParameter())
	return s

}

// NewServer creates a basic server.
func NewServer() Server {
	root, _, err := router.Parse("/")
	if err != nil {
		// It should not come here or router has a bug.
		panic(err)
	}
	c := &server{
		router:      root,
		executors:   map[string]*executor{},
		descriptors: []definition.Descriptor{},
		filters:     []Filter{},
		logger:      &log.SilentLogger{},
	}
	c.pool.New = newHttpContext
	return c
}

type server struct {
	router      router.Router
	executors   map[string]*executor
	descriptors []definition.Descriptor
	modifier    DefinitionModifier
	filters     []Filter
	logger      log.Logger
	pool        sync.Pool
}

// AddFilters add filters to filter requests.
func (c *server) AddFilters(filters ...Filter) {
	c.filters = append(c.filters, filters...)
}

// SetLogger sets logger to server.
func (c *server) SetLogger(logger log.Logger) {
	if logger != nil {
		c.logger = logger
	} else {
		c.logger = &log.SilentLogger{}
	}
}

// SetModifier sets definition modifier.
func (c *server) SetModifier(m DefinitionModifier) {
	c.modifier = m
}

// AddDescriptors adds descriptors to router.
func (c *server) AddDescriptors(descriptors ...definition.Descriptor) error {
	for _, descriptor := range descriptors {
		err := c.addDescriptors("", nil, nil, descriptor)
		if err != nil {
			return err
		}
	}
	c.descriptors = append(c.descriptors, descriptors...)
	return nil
}

func (c *server) addDescriptors(prefix string, consumes []string, produces []string, descriptor definition.Descriptor) error {
	path := strings.Join([]string{prefix, strings.Trim(descriptor.Path, "/")}, "/")
	if len(descriptor.Middlewares) > 0 || len(descriptor.Definitions) > 0 {
		root, leaf, err := router.Parse(path)
		if err != nil {
			c.logger.Errorf("%s: %s", path, err.Error())
			return err
		}
		if len(descriptor.Middlewares) > 0 {
			leaf.AddMiddleware(descriptor.Middlewares...)
		}
		executor := c.executors[path]
		if executor == nil {
			executor = newExecutor()
			c.executors[path] = executor
			leaf.AddExecutor(executor)
		}
		for _, d := range descriptor.Definitions {
			if err = executor.addDefinition(*c.modifyDefinition(&d, consumes, produces)); err != nil {
				c.logger.Errorf("%s: %s", path, err.Error())
				return err
			}
		}
		_, err = c.router.Merge(root)
		if err != nil {
			return err
		}
	}
	if descriptor.Consumes != nil {
		consumes = descriptor.Consumes
	}
	if descriptor.Produces != nil {
		produces = descriptor.Produces
	}
	for _, child := range descriptor.Children {
		err := c.addDescriptors(path, consumes, produces, child)
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *server) modifyDefinition(d *definition.Definition, consumes []string, produces []string) *definition.Definition {
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
		p.Examples = nil
		newOne.Parameters[i] = p
	}
	newOne.Results = make([]definition.Result, len(d.Results))
	for i, r := range d.Results {
		r.Description = ""
		r.Examples = nil
		newOne.Results[i] = r
	}
	if c.modifier != nil {
		c.modifier(newOne)
	}
	return newOne
}

func (c *server) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	for _, f := range c.filters {
		if !f(resp, req) {
			return
		}
	}
	ctx := c.pool.Get().(*httpCtx)
	defer c.pool.Put(ctx)
	ctx.reset(resp, req)
	defer ctx.clear()
	e := c.router.Match(ctx, &ctx.container, req.URL.Path)
	if e == nil {
		http.NotFound(resp, req)
		return
	}
	err := e.Execute(ctx)
	if err != nil {
		c.logger.Error(err.Error())
	}
	if ctx.response.HeaderWritable() {
		http.Error(&ctx.response, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}
	return
}
