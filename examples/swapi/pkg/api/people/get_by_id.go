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
	"fmt"
	"reflect"
	"sort"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/swapi/pkg/model"
)

func getByIdDefinition(people []model.Person) definition.Definition {
	return definition.Definition{
		Method: definition.Get,
		Function: func(ctx context.Context, id int) (model.Person, error) {
			index := sort.Search(len(people), func(i int) bool {
				return people[i].Id > model.Identity(id)
			})
			if index < len(people) && people[index].Id == model.Identity(id) {
				return people[index], nil
			}
			return model.Person{}, fmt.Errorf("not found")
		},
		Consumes: []string{"application/json"},
		Produces: []string{"application/json"},
		Parameters: []definition.Parameter{
			{
				Source: definition.Path,
				Name:   "id",
				Type:   reflect.TypeOf(0),
			},
		},
		Results: []definition.Result{
			{
				Type: definition.Data,
			},
			{
				Type: definition.Error,
			},
		},
	}
}
