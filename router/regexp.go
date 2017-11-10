package router

import (
	"context"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// Index contains the key and it's index of the submatches.
type Index struct {
	// Key is the name for the value.
	Key string
	// Pos is the index of value in submatches.
	Pos int
}

// RegexpNode contains infomation for matching a regexp segment.
type RegexpNode struct {
	Handler
	Progeny
	// Indices contains all positions to get values from submatches.
	Indices []Index
	// Exp is the regular expression.
	Exp string
	// Regexp is a regexp instance to match.
	Regexp *regexp.Regexp
}

// Target returns the matching target of the node.
func (n *RegexpNode) Target() string {
	return n.Exp
}

// Kind returns the kind of the router node.
func (n *RegexpNode) Kind() RouterKind {
	return Regexp
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (n *RegexpNode) Match(ctx context.Context, c Container, path string) Executor {
	// Match self
	index := strings.IndexByte(path, '/')
	if index < 0 {
		index = len(path)
	}
	segment := path[:index]
	result := n.Regexp.FindStringSubmatch(segment)
	if result == nil {
		return nil
	}
	// Match progeny
	var executor Executor
	if index < len(path) {
		executor = n.Handler.Pack(n.Progeny.Match(ctx, c, path[index:]))
	} else {
		executor = n.UnionExecutor(ctx)
	}

	if executor == nil {
		// Unmatched
		return nil
	}

	// Set values
	for _, i := range n.Indices {
		c.Set(i.Key, result[i.Pos])
	}
	return executor
}

// Merge merges r to the current router. The type of r should be same
// as the current one or it panics.
func (n *RegexpNode) Merge(r Router) (Router, error) {
	node, ok := r.(*RegexpNode)
	if !ok {
		return nil, fmt.Errorf("unrecognized regexp router: %s", reflect.TypeOf(r).String())
	}
	if n.Exp != node.Exp {
		return nil, fmt.Errorf("unmatched regexp: %s %s", n.Exp, node.Exp)
	}
	n.Handler.Merge(&node.Handler)
	n.Progeny.Merge(&node.Progeny)
	return n, nil
}

// FullMatchRegexpNode is an optimizing of RegexpNode.
type FullMatchRegexpNode struct {
	Handler
	Progeny
	// Key is the name for the only value.
	Key string
}

// Target returns the matching target of the node.
func (n *FullMatchRegexpNode) Target() string {
	return (&ExpSegment{FullMatchTarget, n.Key}).Target()
}

// Kind returns the kind of the router node.
func (n *FullMatchRegexpNode) Kind() RouterKind {
	return Regexp
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (n *FullMatchRegexpNode) Match(ctx context.Context, c Container, path string) Executor {
	index := strings.IndexByte(path, '/')
	var executor Executor
	if index > 0 {
		executor = n.Handler.Pack(n.Progeny.Match(ctx, c, path[index:]))
	} else {
		index = len(path)
		executor = n.UnionExecutor(ctx)
	}
	if executor == nil {
		// Unmatched
		return nil
	}
	c.Set(n.Key, path[:index])
	return executor
}

// Merge merges r to the current router. The type of r should be same
// as the current one or it panics.
func (n *FullMatchRegexpNode) Merge(r Router) (Router, error) {
	node, ok := r.(*FullMatchRegexpNode)
	if !ok {
		return nil, fmt.Errorf("unrecognized full match router: %s", reflect.TypeOf(r).String())
	}
	if n.Key != node.Key {
		return nil, fmt.Errorf("unmatched full match key: %s %s", n.Key, node.Key)
	}
	n.Handler.Merge(&node.Handler)
	n.Progeny.Merge(&node.Progeny)
	return n, nil
}
