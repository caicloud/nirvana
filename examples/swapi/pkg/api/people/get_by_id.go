package people

import (
	"context"
	"github.com/caicloud/nirvana/examples/swapi/pkg/model"
	"sort"
	"github.com/caicloud/nirvana/definition"
	"reflect"
	"fmt"
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
