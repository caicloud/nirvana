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

package generators

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/gengo/types"
)

func TestResolveAliasAndPtrType(t *testing.T) {
	cases := []struct {
		description string
		input       *types.Type
	}{
		{
			`can resolve: type A *string`,
			&types.Type{
				Kind: types.Alias,
				Underlying: &types.Type{
					Kind: types.Pointer,
					Elem: &types.Type{
						Kind: types.Builtin,
					},
				},
			},
		},
		{
			"can resolve: type A B, type B string",
			&types.Type{
				Kind: types.Alias,
				Underlying: &types.Type{
					Kind: types.Alias,
					Underlying: &types.Type{
						Kind: types.Builtin,
					},
				},
			},
		},
	}
	for _, c := range cases {
		assert.Equal(t, types.Builtin, resolveAliasAndPtrType(c.input).Kind, c.description)
	}
}

func TestGetReferableName(t *testing.T) {
	cases := []struct {
		description string
		m           *types.Member
		name        string
	}{
		{
			"use json tag if it exists",
			&types.Member{
				Name: "A",
				Tags: `json:"a"`,
			},
			"a",
		},
		{
			"use struct name if json tag doesn't exist",
			&types.Member{
				Name: "A",
				Tags: `bson:"a"`,
			},
			"A",
		},
		{
			"skip if json tag is '-'",
			&types.Member{
				Name: "A",
				Tags: `json:"-"`,
			},
			"",
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.name, getReferableName(c.m), c.description)
	}
}
