package router

import (
	"context"
	"fmt"
	"reflect"
)

// PathNode matches all rest path.
type PathNode struct {
	Handler
	// Key is the key for the rest path.
	Key string
}

// Target returns the matching target of the node.
func (n *PathNode) Target() string {
	return ""
}

// Kind returns the kind of the router node.
func (n *PathNode) Kind() RouterKind {
	return Path
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (n *PathNode) Match(ctx context.Context, c Container, path string) Executor {
	c.Set(n.Key, path)
	return n.Handler.UnionExecutor(ctx)
}

// Merge merges r to the current router. The type of r should be same
// as the current one.
func (n *PathNode) Merge(r Router) (Router, error) {
	node, ok := r.(*PathNode)
	if !ok {
		return nil, fmt.Errorf("unrecognized path router: %s", reflect.TypeOf(r).String())
	}
	if n.Key != node.Key {
		return nil, fmt.Errorf("unmatched path key: %s %s", n.Key, node.Key)
	}
	n.Handler.Merge(&node.Handler)
	return n, nil
}
