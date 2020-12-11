/*
Copyright 2020 Caicloud Authors

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

package api

import (
	"strings"
	"testing"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

func TestNewDefinitionShouldNotPanic(t *testing.T) {
	testCases := []struct {
		name      string
		d         *definition.Definition
		expectErr string
	}{
		{
			name: "len(params) < len(args)",
			d: &definition.Definition{
				Method:     definition.Create,
				Function:   func(s string) {},
				Parameters: []definition.Parameter{},
				Results:    []definition.Result{},
			},
			expectErr: "the number of parameters and function args are not equal",
		},
		{
			name: "len(params) > len(args)",
			d: &definition.Definition{
				Method:   definition.Create,
				Function: func() {},
				Parameters: []definition.Parameter{
					definition.QueryParameterFor("Action", "action"),
				},
				Results: []definition.Result{},
			},
			expectErr: "the number of parameters and function args are not equal",
		},
		{
			name: "len(params) = len(args)",
			d: &definition.Definition{
				Method:   definition.Create,
				Function: func(action string) {},
				Parameters: []definition.Parameter{
					definition.QueryParameterFor("Action", "action"),
				},
				Results: []definition.Result{},
			},
		},
		{
			name: "len(results) < len(returns)",
			d: &definition.Definition{
				Method:     definition.Create,
				Function:   func() error { return nil },
				Parameters: []definition.Parameter{},
				Results:    []definition.Result{},
			},
			expectErr: "the number of results and function return values are not equal",
		},
		{
			name: "len(results) > len(returns)",
			d: &definition.Definition{
				Method:     definition.Create,
				Function:   func() {},
				Parameters: []definition.Parameter{},
				Results:    []definition.Result{definition.ErrorResult()},
			},
			expectErr: "the number of results and function return values are not equal",
		},
		{
			name: "len(results) = len(returns)",
			d: &definition.Definition{
				Method:     definition.Create,
				Function:   func() error { return nil },
				Parameters: []definition.Parameter{},
				Results:    []definition.Result{definition.ErrorResult()},
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := NewDefinition(NewTypeContainer(), testCase.d, service.APIStyleRPC)
			if len(testCase.expectErr) > 0 {
				if err == nil {
					t.Fatalf("Unexpected success")
				}
				if !strings.Contains(err.Error(), testCase.expectErr) {
					t.Fatalf("Unexpcted error: expected=%s\ngot=%s", testCase.expectErr, err.Error())
				}
				return
			}
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}
		})
	}
}
