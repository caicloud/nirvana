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

package operators

import (
	"context"
	"reflect"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/api-basic/application"
)

func operater(f interface{}) definition.Operator {
	value := reflect.ValueOf(f)
	return definition.OperatorFunc(func(ctx context.Context, i interface{}) (interface{}, error) {
		params := []reflect.Value{reflect.ValueOf(ctx), reflect.ValueOf(i)}
		results := value.Call(params)
		err := results[1].Interface()
		if err != nil {
			return nil, err.(error)
		}
		return results[0].Interface(), nil
	})
}

func ConvertApplicationV1ToApplicationV2() definition.Operator {
	return operater(func(ctx context.Context, app *application.ApplicationV1) (*application.ApplicationV2, error) {
		return &application.ApplicationV2{
			Metadata: application.Metadata{
				Name:      app.Name,
				Partition: app.Partition,
			},
			Spec: application.ApplicationV2Spec{
				Replica:     app.Replica,
				OtherFields: "Some Default Value",
			},
			Status: application.ApplicationV2Status{
				Phase:   app.Phase,
				Message: app.Message,
			},
		}, nil
	})
}

func ConvertApplicationV2ToApplicationV1() definition.Operator {
	return operater(func(ctx context.Context, app *application.ApplicationV2) (*application.ApplicationV1, error) {
		return &application.ApplicationV1{
			Name:      app.Metadata.Name,
			Partition: app.Metadata.Partition,
			Replica:   app.Spec.Replica,
			// Ignore app.Spec.OtherFields
			Phase:   app.Status.Phase,
			Message: app.Status.Message,
		}, nil
	})
}
