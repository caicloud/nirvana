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
