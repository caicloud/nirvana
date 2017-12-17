package people

import (
	"context"
	"github.com/caicloud/nirvana/examples/swapi/pkg/model"
	"sort"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/validator"
	"reflect"
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
		Consumes: []string{"application/json"},
		Produces: []string{"application/json"},
		Parameters: []definition.Parameter{
			{
				Source: definition.Query,
				Name:   "after",
				Type:   reflect.TypeOf(0),
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
	}
}
