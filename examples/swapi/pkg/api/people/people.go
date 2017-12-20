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

package people

import (
	"sort"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/swapi/pkg/loader"
)

func API(l loader.ModelLoader) definition.Descriptor {
	people := l.LoadPeople()
	sort.Stable(ById(people))

	return definition.Descriptor{
		Path:        "/people",
		Description: "It contains all APIs in v1",
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Definitions: []definition.Definition{
			listDefinition(people),
			getByIdDefinition(people),
		},
	}
}
