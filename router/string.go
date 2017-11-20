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

import (
	"context"
	"fmt"
	"reflect"
	"strings"
)

// StringNode describes a string router node.
type StringNode struct {
	Handler
	Progeny
	// Prefix is the fixed string to match path.
	Prefix string
}

// Target returns the matching target of the node.
func (n *StringNode) Target() string {
	return n.Prefix
}

// Kind returns the kind of the router node.
func (n *StringNode) Kind() RouterKind {
	return String
}

// Match find an executor matched by path.
// The context contains information to inspect executor.
// The container can save key-value pair from the path.
// If the router is the leaf node to match the path, it will return
// the first executor which Inspect() returns true.
func (n *StringNode) Match(ctx context.Context, c Container, path string) Executor {
	if n.Prefix != "" && !strings.HasPrefix(path, n.Prefix) {
		// No match
		return nil
	}
	if len(n.Prefix) < len(path) {
		// Match prefix
		return n.Handler.Pack(n.Progeny.Match(ctx, c, path[len(n.Prefix):]))
	}
	// Match self
	return n.Handler.UnionExecutor(ctx)
}

// Merge merges r to the current router. The type of r should be same
// as the current one or it panics.
func (n *StringNode) Merge(r Router) (Router, error) {
	node, ok := r.(*StringNode)
	if !ok {
		return nil, fmt.Errorf("unrecognized string router: %s", reflect.TypeOf(r).String())
	}
	commonPrefix := 0
	for commonPrefix < len(n.Prefix) && commonPrefix < len(node.Prefix) {
		if n.Prefix[commonPrefix] != node.Prefix[commonPrefix] {
			break
		}
		commonPrefix++
	}
	if commonPrefix <= 0 {
		return nil, fmt.Errorf("there is no common prefix for the two routers")
	}
	switch {
	case commonPrefix == len(n.Prefix) && commonPrefix == len(node.Prefix):
		n.Handler.Merge(&node.Handler)
		n.Progeny.Merge(&node.Progeny)
	case commonPrefix == len(n.Prefix):
		node.Prefix = node.Prefix[commonPrefix:]
		n.AddRouter(node)
	case commonPrefix == len(node.Prefix):
		copy := *n
		copy.Prefix = copy.Prefix[commonPrefix:]
		*n = *node
		n.AddRouter(&copy)
	default:
		copy := *n
		copy.Prefix = copy.Prefix[commonPrefix:]
		node.Prefix = node.Prefix[commonPrefix:]
		n.Handler = Handler{}
		n.Progeny = Progeny{}
		n.Prefix = n.Prefix[:commonPrefix]
		n.AddRouter(&copy)
		n.AddRouter(node)
	}
	return n, nil
}
