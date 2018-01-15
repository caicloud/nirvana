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
	"context"
	"sort"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/swapi/pkg/model"
	"github.com/caicloud/nirvana/operators/validator"
)

func listDefinition(people []model.Person) definition.Definition {
	return definition.Definition{
		Method: definition.Get,
		Function: func(ctx context.Context, after int, first int) ([]model.Person, error) {
			start := sort.Search(len(people), func(i int) bool {
				return people[i].Id > model.Identity(after)
			})
			end := start + first
			if end >= len(people) {
				end = len(people)
			}
			return people[start:end], nil
		},
		Produces: []string{"application/json"},
		Parameters: []definition.Parameter{
			{
				Source: definition.Query,
				Name:   "after",
			},
			{
				Source:  definition.Query,
				Name:    "first",
				Default: 10,
				Operators: []definition.Operator{
					validator.Int("gte=0"),
				},
			},
		},
		Results: []definition.Result{
			{
				Destination: definition.Data,
			},
			{
				Destination: definition.Error,
			},
		},
	}
}
