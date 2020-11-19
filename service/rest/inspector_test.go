/*
Copyright 2018 Caicloud Authors

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

package rest

import (
	"context"
	"testing"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/service/executor"
)

type definitionMap struct {
	def definition.Definition
	err errors.Factory
}

func TestAddDefinition(t *testing.T) {
	inspector := newInspector("/test", &log.SilentLogger{})
	units := []definitionMap{
		{
			definition.Definition{
				Method: definition.Method(""),
			},
			executor.DefinitionNoMethod,
		},
		{
			definition.Definition{
				Method: definition.Get,
			},
			executor.DefinitionNoConsumes,
		},
		{
			definition.Definition{
				Method:   definition.Get,
				Consumes: []string{definition.MIMENone},
			},
			executor.DefinitionNoProduces,
		},
		{
			definition.Definition{
				Method:   definition.Get,
				Consumes: []string{definition.MIMENone},
				Produces: []string{definition.MIMEJSON},
			},
			executor.DefinitionNoErrorProduces,
		},

		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
			},
			executor.DefinitionNoFunction,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function:      1,
			},
			executor.DefinitionInvalidFunctionType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{"invalid-content-type"},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			executor.DefinitionNoConsumer,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{"invalid-content-type"},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			executor.DefinitionNoProducer,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{"invalid-content-type"},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			executor.DefinitionNoProducer,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func(ctx context.Context) {
				},
			},
			executor.DefinitionUnmatchedParameters,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source: "InvalidSource",
					},
				},
				Function: func(a int) {
				},
			},
			service.NoParameterGenerator,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source:  definition.Path,
						Name:    "a",
						Default: "InvalidDefaultValue",
					},
				},
				Function: func(a int) {
				},
			},
			service.UnassignableType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source: definition.Path,
						Name:   "a",
					},
				},
				Function: func(a []*int) {
				},
			},
			service.NoConverter,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source: definition.Query,
						Name:   "a",
						Operators: []definition.Operator{
							definition.OperatorFunc("test", func(ctx context.Context, key string, value string) (int, error) {
								return 1, nil
							}),
						},
					},
				},
				Function: func(a []int) {
				},
			},
			executor.InvalidOperatorOutType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Parameters: []definition.Parameter{
					{
						Source: definition.Query,
						Name:   "a",
						Operators: []definition.Operator{
							definition.OperatorFunc("test", func(ctx context.Context, key string, value string) (int, error) {
								return 1, nil
							}),
							definition.OperatorFunc("test", func(ctx context.Context, key string, value string) ([]int, error) {
								return []int{1}, nil
							}),
						},
					},
				},
				Function: func(a []int) {
				},
			},
			executor.InvalidOperatorInType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() int {
					return 0
				},
			},
			executor.DefinitionUnmatchedResults,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Results: []definition.Result{
					{
						Destination: definition.Destination("InvalidDestination"),
					},
				},
				Function: func() int {
					return 0
				},
			},
			executor.NoDestinationHandler,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Results: []definition.Result{
					{
						Destination: definition.Data,
						Operators: []definition.Operator{
							definition.OperatorFunc("test", func(ctx context.Context, key string, value string) (int, error) {
								return 1, nil
							}),
						},
					},
				},
				Function: func() int {
					return 0
				},
			},
			executor.InvalidOperatorInType,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			nil,
		},
		{
			definition.Definition{
				Method:        definition.Get,
				Consumes:      []string{definition.MIMENone},
				Produces:      []string{definition.MIMEJSON},
				ErrorProduces: []string{definition.MIMEJSON},
				Function: func() {
				},
			},
			executor.DefinitionConflict,
		},
	}

	for _, unit := range units {
		err := inspector.addDefinition(unit.def)
		if unit.err != nil {
			if err == nil {
				t.Errorf("Expected error but got nil for %+v", unit.def)
			} else if !unit.err.Derived(err) {
				t.Fatalf("Unexpected err: %v for %+v", err, unit.def)
			}
		} else if err != nil {
			t.Fatalf("Unexpected err: %v for %+v", err, unit.def)
		}
	}
}
