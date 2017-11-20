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

package namer

import (
	"strings"

	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

// NewPublicNamer is a helper function that returns a namer that makes
// CamelCase names. See the NameStrategy struct for an explanation of the
// arguments to this constructor.
func NewPublicNamer(prependPackageNames int, ignoreWords ...string) *NameStrategy {
	n := &NameStrategy{
		NameStrategy: &namer.NameStrategy{
			Join:                namer.Joiner(namer.IC, namer.IC),
			IgnoreWords:         map[string]bool{},
			PrependPackageNames: prependPackageNames,
		},
	}
	for _, w := range ignoreWords {
		n.IgnoreWords[w] = true
	}
	return n
}

// NewPrivateNamer is a helper function that returns a namer that makes
// camelCase names. See the NameStrategy struct for an explanation of the
// arguments to this constructor.
func NewPrivateNamer(prependPackageNames int, ignoreWords ...string) *NameStrategy {
	n := &NameStrategy{
		NameStrategy: &namer.NameStrategy{
			Join:                namer.Joiner(namer.IL, namer.IC),
			IgnoreWords:         map[string]bool{},
			PrependPackageNames: prependPackageNames,
		},
	}
	for _, w := range ignoreWords {
		n.IgnoreWords[w] = true
	}
	return n
}

// NameStrategy is a general Namer. The easiest way to use it is to copy the
// Public/PrivateNamer variables, and modify the members you wish to change.
//
// The Name method produces a name for the given type, of the forms:
// Anonymous types: <Prefix><Type description><Suffix>
// Named types: <Prefix><Optional Prepended Package name(s)><Original name><Suffix>
//
// In all cases, every part of the name is run through the capitalization
// functions.
//
// The IgnoreWords map can be set if you have directory names that are
// semantically meaningless for naming purposes, e.g. "proto".
//
// Prefix and Suffix can be used to disambiguate parallel systems of type
// names. For example, if you want to generate an interface and an
// implementation, you might want to suffix one with "Interface" and the other
// with "Implementation". Another common use-- if you want to generate private
// types, and one of your source types could be "string", you can't use the
// default lowercase private namer. You'll have to add a suffix or prefix.
type NameStrategy struct {
	*namer.NameStrategy
}

func (ns *NameStrategy) removePrefixAndSuffix(s string) string {
	// The join function may have changed capitalization.
	lowerIn := strings.ToLower(s)
	lowerP := strings.ToLower(ns.Prefix)
	lowerS := strings.ToLower(ns.Suffix)
	b, e := 0, len(s)
	if strings.HasPrefix(lowerIn, lowerP) {
		b = len(ns.Prefix)
	}
	if strings.HasSuffix(lowerIn, lowerS) {
		e -= len(ns.Suffix)
	}
	return s[b:e]
}

// See the comment on NameStrategy.
func (ns *NameStrategy) Name(t *types.Type) string {
	// pre get name
	name := ns.NameStrategy.Name(t)
	// the fllowing type should be changed
	switch t.Kind {
	case types.Slice:
		// the default NameStrategy name []bool to SliceBool
		// but we want BoolSlice
		name = ns.Join(ns.Prefix, []string{
			ns.removePrefixAndSuffix(ns.Name(t.Elem)),
			"Slice",
		}, ns.Suffix)
	}
	ns.Names[t] = name
	return name
}
