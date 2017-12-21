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

package service

import "github.com/caicloud/nirvana/definition"

// DefinitionModifier is used in Server. It's used to modify definition.
// If you want to add some common data into all definitions, you can write
// a customized modifier for it.
type DefinitionModifier func(d *definition.Definition)

// With decorates another modifier. Current modifier is prior to a.
func (m DefinitionModifier) With(a DefinitionModifier) DefinitionModifier {
	return func(d *definition.Definition) {
		m(d)
		a(d)
	}
}

// FirstContextParameter adds a context prefab parameter into all definitions.
// Then you don't need to manually write the parameter to every definitions.
func FirstContextParameter() DefinitionModifier {
	return func(d *definition.Definition) {
		if len(d.Parameters) > 0 {
			p := d.Parameters[0]
			if p.Source == definition.Prefab && p.Name == "context" {
				return
			}
		}
		ps := make([]definition.Parameter, len(d.Parameters)+1)
		ps[0] = definition.Parameter{
			Name:   "context",
			Source: definition.Prefab,
		}
		copy(ps[1:], d.Parameters)
		d.Parameters = ps
	}
}

// LastErrorResult adds a error result into all definitions.
// Then you don't need to manually write the result to every definitions.
func LastErrorResult() DefinitionModifier {
	return func(d *definition.Definition) {
		length := len(d.Results)
		if length > 0 {
			r := d.Results[length-1]
			if r.Type == definition.Error {
				return
			}
		}
		d.Results = append(d.Results, definition.Result{
			Type: definition.Error,
		})
	}
}
