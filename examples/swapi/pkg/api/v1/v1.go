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

package v1

import (
	"context"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/swapi/pkg/api/people"
	"github.com/caicloud/nirvana/examples/swapi/pkg/loader"
)

func API(model loader.ModelLoader) definition.Descriptor {
	return definition.Descriptor{
		Path:        "/api/v1",
		Description: "It contains all APIs in v1",
		Produces:    []string{"application/json"},
		Definitions: []definition.Definition{
			{
				Method:   definition.Get,
				Produces: []string{"application/json"},
				Function: func(ctx context.Context) string {
					return "hello world"
				},
				Parameters: []definition.Parameter{},
				Results: []definition.Result{
					{
						Destination: definition.Data,
					},
				},
			},
		},
		Children: []definition.Descriptor{
			people.API(model),
		},
	}
}
