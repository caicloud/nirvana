/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package router

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// RoutingChain contains the call chain of middlewares and executor.
type RoutingChain interface {
	// Continue continues to execute the next middleware or executor.
	Continue(context.Context) error
}

// Middleware describes the form of middlewares. If you want to
// carry on, call RoutingChain.Continue() and pass the context.
type Middleware func(context.Context, RoutingChain) error

// Executor is an executor for a request.
type Executor interface {
	// Inspect finds a valid executor to execute target context.
	Inspect(context.Context) (Executor, bool)
	// Execute executes with context.
	Execute(context.Context) error
}

// RouterKind is kind of routers.
type RouterKind string

const (
	// String means the router has a fixed string.
	String RouterKind = "String"
	// Regexp means the router has a regular expression.
	Regexp RouterKind = "Regexp"
	// Path means the router matches the rest. Path router only can
	// be placed at the leaf node.
	Path RouterKind = "Path"
)

// Container is a key-value container. It saves key-values from path.
type Container interface {
	// Set sets key-value into the container.
	Set(key, value string)
	// Get gets a value by key from the container.
	Get(key string) (string, bool)
}

// Router describes the interface of a router node.
type Router interface {
	// Target returns the matching target of the node.
	// It can be a fixed string or a regular expression.
	Target() string
	// Kind returns the kind of the router node.
	Kind() RouterKind
	// Match find an executor matched by path.
	// The context contains information to inspect executor.
	// The container can save key-value pair from the path.
	// If the router is the leaf node to match the path, it will return
	// the first executor which Inspect() returns true.
	Match(ctx context.Context, c Container, path string) Executor
	// AddMiddleware adds middleware to the router node.
	// If the router matches a path, all middlewares in the router
	// will be executed by the returned executor.
	AddMiddleware(ms ...Middleware)
	// AddExecutor adds executor to the router node.
	// A router can hold many executors, but there is only one executor
	// is selected for a match.
	AddExecutor(es ...Executor)
	// Merge merges r to the current router. The type of r should be same
	// as the current one or it panics.
	//
	// For instance:
	//  Router A: /namespaces/ -> {namespace}
	//  Router B: /nameless/ -> {other}
	// Result:
	//  /name -> spaces/ -> {namespace}
	//       |-> less/ -> {other}
	Merge(r Router) (Router, error)
}

const (
	// Default match regular expression. All regexp router without expression
	// will use the expression.
	FullMatchTarget = ".*"
	// Tail match expression.
	TailMatchTarget = "*"
)

// Parse parses a path to a router tree. It returns the root router and
// the leaf router. you can add middlewares and executor to the routers.
// A valid path should like:
//  /segments/{segment}/resources/{resource}
//  /segments/{segment:[a-z]{1,2}}.log/paths/{path:*}
func Parse(path string) (Router, Router, error) {
	paths, err := Split(path)
	if err != nil {
		return nil, nil, err
	}
	if len(paths) <= 0 {
		return nil, nil, fmt.Errorf("invalid path")
	}
	segments, err := Reorganize(paths)
	if err != nil {
		return nil, nil, err
	}
	var root Router
	var leaf Router
	var parent Router
	for i, seg := range segments {
		router, err := SegmentToRouter(seg)
		if err != nil {
			return nil, nil, err
		}
		if i == 0 {
			root = router
		}
		if i == len(segments)-1 {
			leaf = router
		}
		if parent != nil {
			if c, ok := parent.(interface {
				AddRouter(router Router)
			}); ok {
				c.AddRouter(router)
			} else {
				return nil, nil, fmt.Errorf("router does not implement RouterContainer: %s", reflect.TypeOf(parent).String())
			}
		}
		parent = router
	}
	return root, leaf, nil
}

// SegmentToRouter converts segment to a router.
func SegmentToRouter(seg *Segment) (Router, error) {
	switch seg.Kind {
	case String:
		return &StringNode{
			Prefix: seg.Match,
		}, nil
	case Regexp:
		if len(seg.Keys) == 1 && seg.Match == (&ExpSegment{FullMatchTarget, seg.Keys[0]}).Target() {
			return &FullMatchRegexpNode{
				Key: seg.Keys[0],
			}, nil
		} else {
			node := &RegexpNode{
				Exp: seg.Match,
			}
			r, err := regexp.Compile("^" + seg.Match + "$")
			if err != nil {
				return nil, err
			}
			node.Regexp = r
			names := r.SubexpNames()
			j := 0
			for i := 0; i < len(names) && j < len(seg.Keys); i++ {
				if names[i] == seg.Keys[j] {
					node.Indices = append(node.Indices, Index{names[i], i})
					j++
				}
			}
			if j != len(seg.Keys) {
				return nil, fmt.Errorf("unmatched keys: %+v", seg)
			}
			return node, nil
		}
	case Path:
		return &PathNode{
			Key: seg.Keys[0],
		}, nil
	}
	return nil, fmt.Errorf("unknown segment: %+v", seg)
}

// Split splits string segments and regexp segments.
//
// For instance:
//  /segments/{segment:[a-z]{1,2}}.log/paths/{path:*}
// TO:
//  /segments/ {segment:[a-z]{1,2}} .log/paths/ {path:*}
func Split(path string) ([]string, error) {
	result := make([]string, 0, 5)
	lastElementPos := 0
	braceCounter := 0
	for i, c := range path {
		switch c {
		case '{':
			braceCounter++
			if braceCounter == 1 {
				if i > lastElementPos {
					result = append(result, path[lastElementPos:i])
				}
				lastElementPos = i
			}
		case '}':
			braceCounter--
			if braceCounter == 0 {
				result = append(result, path[lastElementPos:i+1])
				lastElementPos = i + 1
			}
		}
	}
	if braceCounter > 0 {
		return nil, fmt.Errorf("unmatched braces")
	}
	if lastElementPos < len(path) {
		result = append(result, path[lastElementPos:])
	}
	return result, nil
}

// Segment contains information to construct a router.
type Segment struct {
	// Match is the target string.
	Match string
	// Keys contains keys from segments.
	Keys []string
	// Kind is the router kind which the segment can be converted to.
	Kind RouterKind
}

// Reorganize reorganizes the form of paths.
//
// For instance:
//  /segments/ {segment:[a-z]{1,2}} .log/paths/ {path:*}
// To:
//  {/segments/ {} String} {(?P<segment>[a-z]{1,2})\.log {segment} Regexp} {/paths/ {} String} { {path} Path}
func Reorganize(paths []string) ([]*Segment, error) {
	segments := make([]*Segment, 0, len(paths))
	var segment *Segment
	for i := 0; i < len(paths); i++ {
		p := paths[i]
		if !strings.HasPrefix(p, "{") {
			if segment == nil {
				// String segment
				segments = append(segments, &Segment{p, nil, String})
			} else {
				// Regexp segment
				slashPos := strings.Index(p, "/")
				if slashPos < 0 {
					// No slash
					segment.Match += regexp.QuoteMeta(p)
				} else {
					segment.Match += regexp.QuoteMeta(p[:slashPos])
					segments = append(segments, segment, &Segment{p[slashPos:], nil, String})
					segment = nil
				}
			}
		} else {
			// Regexp segment
			seg, err := ParseExpSegment(p)
			if err != nil {
				return nil, err
			}
			if seg.Tail() {
				if i != len(paths)-1 {
					return nil, fmt.Errorf("key %s should be last element in the path", seg.Key)
				}
				if segment != nil {
					segments = append(segments, segment)
					segment = nil
				}
				segments = append(segments, &Segment{"", []string{seg.Key}, Path})
				break
			}
			if segment == nil {
				segment = &Segment{"", []string{}, Regexp}
			}
			segment.Match += seg.Target()
			segment.Keys = append(segment.Keys, seg.Key)
		}
	}
	if segment != nil {
		segments = append(segments, segment)
	}
	return segments, nil
}

// ExpSegment describes a regexp segment.
type ExpSegment struct {
	// Exp is the regular expression.
	Exp string
	// Key is the key for the expression.
	Key string
}

// ParseExpSegment parses a regexp segment to ExpSegment.
func ParseExpSegment(exp string) (*ExpSegment, error) {
	if !strings.HasPrefix(exp, "{") || !strings.HasSuffix(exp, "}") {
		return nil, fmt.Errorf("exp does not have normative format: %s", exp)
	}
	exp = exp[1 : len(exp)-1]
	pos := strings.Index(exp, ":")
	seg := &ExpSegment{}
	if pos < 0 {
		seg.Exp = FullMatchTarget
		seg.Key = exp
	} else {
		seg.Exp = exp[pos+1:]
		seg.Key = exp[:pos]
	}
	return seg, nil
}

// Tail returns whether the segment contains a tail match target.
func (es *ExpSegment) Tail() bool {
	return es.Exp == TailMatchTarget
}

// Target returns the whole regular expression for the segment.
func (es *ExpSegment) Target() string {
	return fmt.Sprintf("(?P<%s>%s)", es.Key, es.Exp)
}
