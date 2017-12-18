package people

import (
	"github.com/caicloud/nirvana/examples/swapi/pkg/loader"
	"github.com/caicloud/nirvana/definition"
	"context"
	"github.com/caicloud/nirvana/examples/swapi/pkg/model"
	"reflect"
	"github.com/caicloud/nirvana/validator"
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
