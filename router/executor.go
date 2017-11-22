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

package router

import "context"

// MiddlewareExecutor is a combination of middlewares and executor.
type MiddlewareExecutor struct {
	// Middlewares contains all middlewares for the executor.
	Middlewares []Middleware
	// Index is used to record the count of executed middleware.
	Index int
	// Executor executes after middlewares.
	Executor
}

// NewMiddlewareExecutor creates a new executor with middlewares.
func NewMiddlewareExecutor(ms []Middleware, e Executor) Executor {
	return &MiddlewareExecutor{ms, 0, e}
}

// Inspect checks whether the executor can executes with the context.
func (me *MiddlewareExecutor) Inspect(c context.Context) (Executor, bool) {
	if _, ok := me.Executor.Inspect(c); !ok {
		return nil, false
	}
	return me, true
}

// Execute executes middlewares and executor.
func (me *MiddlewareExecutor) Execute(c context.Context) error {
	me.Index = 0
	return me.Continue(c)
}

// Continue continues to execute the next middleware or executor.
func (me *MiddlewareExecutor) Continue(c context.Context) error {
	if me.Index >= len(me.Middlewares) {
		if me.Executor != nil {
			return me.Executor.Execute(c)
		}
		return nil
	}
	m := me.Middlewares[me.Index]
	me.Index++
	return m(c, me)
}

// Executors is a list of executors.
type Executors []Executor

// Inspect finds an appropriate executor from the list.
func (e Executors) Inspect(c context.Context) (Executor, bool) {
	for _, executor := range e {
		if result, ok := executor.Inspect(c); ok {
			return result, true
		}
	}
	return nil, false
}

// Execute executes the fisrt appropriate executor.
func (e Executors) Execute(c context.Context) error {
	for _, executor := range e {
		if result, ok := executor.Inspect(c); ok {
			return result.Execute(c)
		}
	}
	return nil
}
