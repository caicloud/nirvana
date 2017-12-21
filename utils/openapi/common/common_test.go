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

package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOpenAPITypeFormat(t *testing.T) {
	cases := map[string][]string{
		"int":         {"integer", "int"},
		"uint":        {"integer", "uint"},
		"int8":        {"integer", "int8"},
		"uint8":       {"integer", "uint8"},
		"int16":       {"integer", "int16"},
		"uint16":      {"integer", "uint16"},
		"int32":       {"integer", "int32"},
		"uint32":      {"integer", "uint32"},
		"int64":       {"integer", "int64"},
		"uint64":      {"integer", "uint64"},
		"byte":        {"integer", "uint8"},
		"[]byte":      {"string", "byte"},
		"float64":     {"number", "double"},
		"float32":     {"number", "float"},
		"bool":        {"boolean", ""},
		"time.Time":   {"string", "date-time"},
		"string":      {"string", ""},
		"integer":     {"integer", ""},
		"number":      {"number", ""},
		"boolean":     {"boolean", ""},
		"interface{}": {"object", ""},
	}
	for k, y := range cases {
		typ, format := GetOpenAPITypeFormat(k)
		assert.Equal(t, y[0], typ, "can get openapi type")
		assert.Equal(t, y[1], format, "can get openapi format")
	}
}

func TestEscapeJSONPointer(t *testing.T) {
	cases := []struct {
		input  string
		output string
	}{
		{
			"xxx",
			"xxx",
		},
		{
			"x~x",
			"x~0x",
		},
		{
			"x/x/x",
			"x~1x~1x",
		},
	}
	for _, c := range cases {
		assert.Equal(t, c.output, EscapeJSONPointer(c.input), "should escape json pointer")
	}
}
