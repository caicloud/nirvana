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
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/swapi/pkg/loader"
	"github.com/caicloud/nirvana/examples/swapi/pkg/model"
	"github.com/caicloud/nirvana/validator"
	"reflect"
)

func API(l loader.ModelLoader) definition.Descriptor {
	people := l.LoadPeople()
	return definition.Descriptor{
		Path:        "/people",
		Description: "It contains all APIs in v1",
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Definitions: []definition.Definition{
			{
				Method: definition.Get,
				Function: func(ctx context.Context, after string, first int) ([]model.Person, error) {
					return people[:10], nil
				},
				Consumes: []string{"application/json"},
				Produces: []string{"application/json"},
				Parameters: []definition.Parameter{
					{
						Source: definition.Query,
						Name:   "after",
						Type:   reflect.TypeOf(""),
					},
					{
						Source:  definition.Query,
						Name:    "first",
						Type:    reflect.TypeOf(0),
						Default: 10,
						Operators: []definition.Operator{
							validator.Var("gte=0"),
						},
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
			},
		},
	}
}
