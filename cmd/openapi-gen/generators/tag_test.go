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

func TestGetOpenAPITagValue(t *testing.T) {
	cases := []struct {
		description string
		comments    []string
		values      []string
	}{
		{
			"should get values successfully",
			[]string{
				`+nirvana:openapi`,
				`+nirvana:openapi=true`,
				`+nirvana:openapi=false`,
				`+nirvana:openapi=xxx`,
			},
			[]string{
				``,
				`true`,
				`false`,
				`xxx`,
			},
		},
		{
			"should ignore wrong case",
			[]string{
				`asdfasdg`,
				`+    nirvana:openapi`,
				`+  nirvana:openapi=true`,
				`+nirvana:openapi=false`,
				`+ nirvana:openapi=  true`,
				`+nirvana:openapi=xxx`,
			},
			[]string{
				`false`,
				`xxx`,
			},
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.values, getOpenAPITagValue(c.comments), c.description)
	}
}

func TestHasOpenAPITagValue(t *testing.T) {
	cases := []struct {
		description string
		comments    []string
		has         []string
		not         []string
	}{
		{
			"should get values successfully",
			[]string{
				`+nirvana:openapi`,
				`+nirvana:openapi=true`,
				`+nirvana:openapi=false`,
				`+nirvana:openapi=xxx`,
			},
			[]string{
				``,
				`true`,
				`false`,
				`xxx`,
			},
			[]string{
				`xxxxxx`,
				`yyy`,
			},
		},
		{
			"should ignore wrong case",
			[]string{
				`asdfasdg`,
				`+    nirvana:openapi`,
				`+  nirvana:openapi=true`,
				`+nirvana:openapi=false`,
				`+ nirvana:openapi=  true`,
				`+nirvana:openapi=xxx`,
			},
			[]string{
				`false`,
				`xxx`,
			},
			[]string{
				`true`,
				`yasdy`,
			},
		},
	}
	for _, c := range cases {
		for _, v := range c.has {

			assert.Equal(t, true, hasOpenAPITagValue(c.comments, v), c.description)
		}
		for _, v := range c.not {

			assert.Equal(t, false, hasOpenAPITagValue(c.comments, v), c.description)
		}
	}
}

func TestHasOptionalTag(t *testing.T) {
	cases := []struct {
		description string
		m           *types.Member
		optional    bool
	}{
		{
			"Tag doesn't have omitempty",
			&types.Member{
				Tags: `json:"xxx"`,
			},
			false,
		},
		{
			"Tag has omitempty",
			&types.Member{
				Tags: `json:"xxx,omitempty"`,
			},
			true,
		},
		{
			"tag which is not json tag",
			&types.Member{
				Tags: `bson:"xxx,omitempty"`,
			},
			false,
		},
		{
			"tag which is not right tag",
			&types.Member{
				Tags: `bson:xxx,omitempty"`,
			},
			false,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.optional, hasOptionalTag(c.m), c.description)
	}
}

func TestGetJSONTag(t *testing.T) {
	cases := []struct {
		description string
		m           *types.Member
		tags        []string
	}{
		{
			"should get single json tag successfully",
			&types.Member{
				Tags: `json:"aaa"`,
			},
			[]string{
				"aaa",
			},
		},
		{
			"should get json tag with omitempty successfully",
			&types.Member{
				Tags: `json:"aaa,omitempty"`,
			},
			[]string{
				"aaa", "omitempty",
			},
		},
		{
			"should get json tag with inline successfully",
			&types.Member{
				Tags: `json:",inline"`,
			},
			[]string{
				"", "inline",
			},
		},
		{
			"shouldn't get bson tag",
			&types.Member{
				Tags: `bson:",inline"`,
			},
			nil,
		},
		{
			"shouldn't get wrong json tag",
			&types.Member{
				Tags: `json:,inline"`,
			},
			nil,
		},
	}

	for _, c := range cases {
		assert.Equal(t, c.tags, getJSONTags(c.m), c.description)
	}
}

func TestIsInline(t *testing.T) {
	cases := []struct {
		description string
		m           *types.Member
		inline      bool
	}{
		{
			"member is inline",
			&types.Member{
				Tags: `json:",inline"`,
			},
			true,
		},
		{
			"member is not inline",
			&types.Member{
				Tags: `json:"mm"`,
			},
			false,
		},
		{
			"member is omitempty but not inline",
			&types.Member{
				Tags: `json:"mm,omitempty"`,
			},
			false,
		},
		{
			"member have no json tag",
			&types.Member{
				Tags: `bson:",inline"`,
			},
			false,
		},
		{
			"member have wrong json tag",
			&types.Member{
				Tags: `json:,inline"`,
			},
			false,
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.inline, isInline(c.m), c.description)
	}
}
