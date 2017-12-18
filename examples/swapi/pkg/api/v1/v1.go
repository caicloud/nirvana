package v1

import (
	"github.com/caicloud/nirvana/examples/swapi/pkg/loader"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/swapi/pkg/api/people"
	"context"
)

func API(model loader.ModelLoader) definition.Descriptor {
	return definition.Descriptor{
		Path:        "/api/v1",
		Description: "It contains all APIs in v1",
		Consumes:    []string{"application/json"},
		Produces:    []string{"application/json"},
		Definitions: []definition.Definition{
			{
				Method:   definition.Get,
				Consumes: []string{"application/json"},
				Produces: []string{"application/json"},
				Function: func(ctx context.Context) string {
					return "hello world"
				},
				Parameters: []definition.Parameter{},
				Results: []definition.Result{
					{
						Type: definition.Data,
					},
				},
			},
		},
		Children: []definition.Descriptor{
			people.API(model),
		},
	}
}
