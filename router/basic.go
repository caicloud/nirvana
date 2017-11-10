package router

import (
	"context"
	"fmt"
	"reflect"
	"sort"
)

// Handler contains middlewares and executor.
type Handler struct {
	Middlewares []Middleware
	Executor    Executor
}

// AddMiddleware adds middleware to the router node.
// If the router matches a path, all middlewares in the router
// will be executed by the returned executor.
func (h *Handler) AddMiddleware(ms ...Middleware) {
	h.Middlewares = append(h.Middlewares, ms...)
}

// AddExecutor adds executor to the router node.
// A router can hold many executors, but there is only one executor
// is selected for a match.
func (h *Handler) AddExecutor(es ...Executor) {
	var executors Executors
	if h.Executor != nil {
		if array, ok := h.Executor.(Executors); ok {
			executors = array
		} else {
			executors = Executors{h.Executor}
		}
	}
	for _, e := range es {
		if e != nil {
			if array, ok := e.(Executors); ok {
				executors = append(executors, array...)
			} else {
				executors = append(executors, e)
			}
		}
	}
	if executors != nil {
		h.Executor = executors
	}
}

// Merge merges middlewares and executors.
func (h *Handler) Merge(o *Handler) {
	h.AddMiddleware(o.Middlewares...)
	h.AddExecutor(o.Executor)
}

// Pack packs middlewares with the executor.
func (h *Handler) Pack(e Executor) Executor {
	if e == nil {
		return nil
	}
	if len(h.Middlewares) <= 0 {
		return e
	}
	return NewMiddlewareExecutor(h.Middlewares, e)
}

// UnionExecutor packs middlewares and own executor.
func (h *Handler) UnionExecutor(ctx context.Context) Executor {
	if h.Executor == nil {
		return nil
	}
	e, ok := h.Executor.Inspect(ctx)
	if !ok {
		return nil
	}
	return h.Pack(e)
}

// CharRouter
type CharRouter struct {
	Char   byte
	Router *StringNode
}

// Progeny contains all children routers.
type Progeny struct {
	StringRouters []CharRouter
	RegexpRouters []Router
	PathRouter    Router
}

// FindStringRouter find a router by first char.
func (p *Progeny) FindStringRouter(char byte) Router {
	length := len(p.StringRouters)
	if length <= 3 {
		// If the length is less than 3, use linear search.
		for _, cr := range p.StringRouters {
			if cr.Char == char {
				return cr.Router
			}
		}
		return nil
	}
	// Binary search.
	index := sort.Search(len(p.StringRouters), func(i int) bool {
		return char <= p.StringRouters[i].Char
	})
	if index >= length {
		return nil
	}
	target := p.StringRouters[index]
	if char != target.Char {
		return nil
	}
	return target.Router
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (p *Progeny) Match(ctx context.Context, c Container, path string) Executor {
	if len(path) <= 0 {
		return nil
	}

	// Match string routers
	if len(p.StringRouters) > 0 {
		if router := p.FindStringRouter(path[0]); router != nil {
			if executor := router.Match(ctx, c, path); executor != nil {
				return executor
			}
		}
	}

	// Match regexp routers
	for _, regexp := range p.RegexpRouters {
		if executor := regexp.Match(ctx, c, path); executor != nil {
			return executor
		}
	}

	// Match path router
	if p.PathRouter != nil {
		return p.PathRouter.Match(ctx, c, path)
	}
	return nil
}

//  AddRouter adds a router to current progeny.
func (p *Progeny) AddRouter(router Router) {
	switch router.Kind() {
	case String:
		target := router.Target()
		if len(target) <= 0 {
			panic("invalid router target")
		}
		r, ok := router.(*StringNode)
		if !ok {
			panic(fmt.Sprintf("unknown string node: %s", reflect.TypeOf(router).String()))
		}
		c := target[0]
		sr := p.FindStringRouter(c)
		if sr != nil {
			_, err := sr.Merge(router)
			if err != nil {
				panic(err.Error())
			}
			return
		}
		length := len(p.StringRouters)
		index := 0
		if length > 0 {
			index = sort.Search(length, func(i int) bool {
				return c < p.StringRouters[i].Char
			})
		}
		cr := CharRouter{c, r}
		if index >= length {
			p.StringRouters = append(p.StringRouters, cr)
		} else {
			p.StringRouters = append(p.StringRouters[:index+1], p.StringRouters[index:]...)
			p.StringRouters[index] = cr
		}
	case Regexp:
		found := false
		for _, r := range p.RegexpRouters {
			if r.Target() == router.Target() {
				_, err := r.Merge(router)
				if err != nil {
					panic(err.Error())
				}
				found = true
				break
			}
		}
		if !found {
			p.RegexpRouters = append(p.RegexpRouters, router)
		}
	case Path:
		if p.PathRouter != nil {
			r, err := p.PathRouter.Merge(router)
			if err != nil {
				panic(fmt.Sprintf("failed to merge path router : %s", err.Error()))
			}
			p.PathRouter = r
		} else {
			p.PathRouter = router
		}
	default:
		panic(fmt.Sprintf("unknwon router kind: %s", router.Kind()))
	}
}

// Merge merges children routers.
func (p *Progeny) Merge(o *Progeny) {
	for _, r := range o.StringRouters {
		p.AddRouter(r.Router)
	}
	for _, r := range o.RegexpRouters {
		p.AddRouter(r)
	}
	if o.PathRouter != nil {
		p.AddRouter(o.PathRouter)
	}
}
