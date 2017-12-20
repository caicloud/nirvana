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
	"context"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/router"
)

type executor struct {
	cells map[string][]*cell
}

func newExecutor() *executor {
	return &executor{map[string][]*cell{}}
}

func (e *executor) addDefinition(d definition.Definition) error {
	method := HTTPMethodFor(d.Method)
	if method == "" {
		return fmt.Errorf("no http method for %s", d.Method)
	}

	if !methodIsHeadOrGet(method) && len(d.Consumes) <= 0 {
		return fmt.Errorf("no content type to consume for %s", d.Method)
	}
	if len(d.Produces) <= 0 {
		return fmt.Errorf("no content type to produce for %s", d.Method)
	}
	if d.Function == nil {
		return fmt.Errorf("no function for %s", d.Method)
	}
	value := reflect.ValueOf(d.Function)
	if value.Kind() != reflect.Func {
		return fmt.Errorf("function type %s is invalid", value.Type().String())
	}
	c := &cell{
		method:   method,
		code:     HTTPCodeFor(d.Method),
		function: value,
	}
	for _, ct := range d.Consumes {
		if consumer := ConsumerFor(ct); consumer != nil {
			c.consumers = append(c.consumers, consumer)
		} else {
			return fmt.Errorf("no consumer for content type %s", ct)
		}
	}
	for _, ct := range d.Produces {
		if producer := ProducerFor(ct); producer != nil {
			c.producers = append(c.producers, producer)
		} else {
			return fmt.Errorf("no producer for content type %s", ct)
		}
	}
	ps, err := e.generateParameters(value.Type(), d.Parameters)
	if err != nil {
		return err
	}
	c.parameters = ps
	rs, err := e.generateResults(value.Type(), d.Results)
	if err != nil {
		return err
	}
	c.results = rs
	if err := e.conflictCheck(c); err != nil {
		return err
	}
	e.cells[method] = append(e.cells[method], c)
	return nil
}

func (e *executor) conflictCheck(c *cell) error {
	cs := e.cells[c.method]
	if len(cs) <= 0 {
		return nil
	}
	ctMap := map[string]bool{}
	for _, extant := range cs {
		result := extant.ctMap()
		for k, vs := range result {
			for _, v := range vs {
				ctMap[k+":"+v] = true
			}
		}
	}
	cMap := c.ctMap()
	for k, vs := range cMap {
		for _, v := range vs {
			if !ctMap[k+":"+v] {
				return fmt.Errorf("consumer-producer pair %s:%s conflicts", k, v)
			}
		}
	}
	return nil
}

func (e *executor) generateParameters(typ reflect.Type, ps []definition.Parameter) ([]parameter, error) {
	if typ.NumIn() != len(ps) {
		return nil, fmt.Errorf("function parameters number does not adapt to definition")
	}
	parameters := make([]parameter, 0, len(ps))
	for i, p := range ps {
		generator := ParameterGeneratorFor(p.Source)
		if generator == nil {
			return nil, fmt.Errorf("no parameter generator for source %s", p.Source)
		}

		param := parameter{
			name:         p.Name,
			defaultValue: p.Default,
			generator:    generator,
			operators:    p.Operators,
		}
		if p.Type == nil {
			param.targetType = typ.In(i)
		} else {
			param.targetType = p.Type
		}
		if err := generator.Validate(param.name, param.defaultValue, param.targetType); err != nil {
			return nil, err
		}
		parameters = append(parameters, param)
	}
	return parameters, nil
}

func (e *executor) generateResults(typ reflect.Type, rs []definition.Result) ([]result, error) {
	if typ.NumOut() != len(rs) {
		return nil, fmt.Errorf("function results number does not adapt to definition")
	}
	results := make([]result, 0, len(rs))
	for i, r := range rs {
		handler := TypeHandlerFor(r.Type)
		if handler == nil {
			return nil, fmt.Errorf("no type handler for type %s", r.Type)
		}
		result := result{
			index:     i,
			handler:   handler,
			operators: r.Operators,
		}
		if err := handler.Validate(typ.Out(i)); err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	sort.Sort(resultsSorter(results))
	return results, nil

}

type resultsSorter []result

// Len is the number of elements in the collection.
func (s resultsSorter) Len() int {
	return len(s)
}

// Less reports whether the element with
// index i should sort before the element with index j.
func (s resultsSorter) Less(i, j int) bool {
	return s[i].handler.Priority() < s[j].handler.Priority()
}

// Swap swaps the elements with indexes i and j.
func (s resultsSorter) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func methodIsHeadOrGet(method string) bool {
	return http.MethodGet == method || http.MethodHead == method
}

// Inspect finds a valid executor to execute target context.
func (e *executor) Inspect(ctx context.Context) (router.Executor, bool) {
	req := HTTPRequest(ctx)
	if req == nil {
		return nil, false
	}
	cells := []*cell{}
	if cs, ok := e.cells[req.Method]; ok && len(cs) > 0 {
		cells = append(cells, cs...)
	}
	if len(cells) <= 0 {
		return nil, false
	}
	ct, err := ContentType(req)
	if err != nil || (!methodIsHeadOrGet(req.Method) && ct == "") {
		return nil, false
	}
	accepted := 0
	for i, c := range cells {
		if c.acceptable(ct) {
			if accepted != i {
				cells[accepted] = c
			}
			accepted++
		}
	}
	if accepted <= 0 {
		return nil, false
	}
	ats, err := AcceptTypes(req)
	if err != nil || len(ats) <= 0 {
		return nil, false
	}
	cells = cells[:accepted]
	var target *cell
	for _, c := range cells {
		if c.producible(ats) {
			target = c
			break
		}
	}
	if target == nil {
		for _, at := range ats {
			if at == acceptTypeAll {
				target = cells[0]
			}
		}
	}
	if target == nil {
		return nil, false
	}
	return target, true
}

// Execute executes with context.
func (e *executor) Execute(ctx context.Context) error {
	if e, ok := e.Inspect(ctx); ok {
		return e.Execute(ctx)
	}
	return fmt.Errorf("no executor to execute")
}

type cell struct {
	method     string
	code       int
	consumers  []Consumer
	producers  []Producer
	parameters []parameter
	results    []result
	function   reflect.Value
}

type parameter struct {
	name         string
	targetType   reflect.Type
	defaultValue interface{}
	generator    ParameterGenerator
	operators    []definition.Operator
}

type result struct {
	index     int
	handler   TypeHandler
	operators []definition.Operator
}

func (e *cell) ctMap() map[string][]string {
	result := map[string][]string{}
	for _, c := range e.consumers {
		for _, p := range e.producers {
			ct := c.ContentType()
			result[ct] = append(result[ct], p.ContentType())
		}
	}
	return result
}

func (e *cell) acceptable(ct string) bool {
	if methodIsHeadOrGet(e.method) {
		return true
	}
	for _, c := range e.consumers {
		if c.ContentType() == ct {
			return true
		}
	}
	return false
}

func (e *cell) producible(ats []string) bool {
	for _, at := range ats {
		for _, c := range e.producers {
			if c.ContentType() == at {
				return true
			}
		}
	}
	return false
}

// Inspect finds a valid executor to execute target context.
func (e *cell) Inspect(ctx context.Context) (router.Executor, bool) {
	return e, true
}

func (e *cell) formatError(flag string, name string, err error) error {
	return fmt.Errorf("[%s]%s: %s", strings.ToLower(flag), name, err.Error())
}

// Execute executes with context.
func (e *cell) Execute(ctx context.Context) (err error) {
	c := httpContext(ctx)
	if c == nil {
		return fmt.Errorf("can't find http context")
	}
	paramValues := make([]reflect.Value, 0, len(e.parameters))
	for _, p := range e.parameters {
		result, err := p.generator.Generate(ctx, &c.container, p.name, p.targetType)
		if err != nil {
			http.Error(&c.response,
				e.formatError(string(p.generator.Source()), p.name, err).Error(),
				http.StatusBadRequest)
			return nil
		}
		if result == nil && p.defaultValue != nil {
			result = p.defaultValue
		}
		for _, operator := range p.operators {
			result, err = operator.Operate(ctx, result)
			if err != nil {
				http.Error(&c.response,
					e.formatError(string(p.generator.Source()), p.name, err).Error(),
					http.StatusBadRequest)
				return nil
			}
		}
		if result == nil {
			http.Error(&c.response,
				e.formatError(string(p.generator.Source()), p.name, fmt.Errorf("required field but got empty")).Error(),
				http.StatusBadRequest)
			return nil
		} else if closer, ok := result.(io.Closer); ok {
			defer func() {
				if e := closer.Close(); e != nil && err == nil {
					// Need to print error here.
					err = e
				}
			}()
		}

		paramValues = append(paramValues, reflect.ValueOf(result))
	}
	resultValues := e.function.Call(paramValues)
	for _, r := range e.results {
		v := resultValues[r.index]
		data := v.Interface()
		for _, operator := range r.operators {
			newData, err := operator.Operate(ctx, data)
			if err != nil {
				return err
			}
			data = newData
		}
		if data != nil {
			if closer, ok := data.(io.Closer); ok {
				defer func() {
					if e := closer.Close(); e != nil && err == nil {
						// Need to print error here.
						err = e
					}
				}()
			}
		}
		goon, err := r.handler.Handle(ctx, e.producers, e.code, data)
		if err != nil {
			return err
		}
		if !goon {
			break
		}
	}
	if c.response.HeaderWritable() {
		c.response.WriteHeader(e.code)
	}
	return nil
}
