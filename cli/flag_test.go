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

package cli

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type testCase struct {
	flag     string
	defValue interface{}
	want     interface{}
}

func getTestCase(t string) testCase {
	switch t {
	case "Bool":
		return testCase{
			flag:     "true",
			defValue: false,
			want:     true,
		}
	case "BoolSlice":
		return testCase{
			flag:     "true,false",
			defValue: []bool{},
			want:     []bool{true, false},
		}
	case "Duration":
		return testCase{
			flag:     "1s",
			defValue: time.Duration(0),
			want:     time.Second,
		}
	case "Float32", "Float64":
		return testCase{
			flag:     "1.32",
			defValue: 1,
			want:     1.32,
		}
	case "Int", "Int16", "Int32", "Int64", "Uint", "Uint16", "Uint32", "Uint64":
		return testCase{
			flag:     "1",
			defValue: 0,
			want:     1,
		}
	case "IntSlice":
		return testCase{
			flag:     "1,2,3",
			defValue: []int{},
			want:     []int{1, 2, 3},
		}
	case "String":
		return testCase{
			flag:     "dev",
			defValue: "",
			want:     "dev",
		}
	case "StringSlice":
		return testCase{
			flag:     "for,dev",
			defValue: []string{},
			want:     []string{"for", "dev"},
		}
	default:
		return testCase{
			flag:     "",
			defValue: "",
			want:     "",
		}
	}
}

func Test_mergeWithEnvPrefix(t *testing.T) {

	type T struct {
		name string
		key  string
		want string
	}

	tests := []T{
		{"no prefix", "Key", "KEY"},
		{"no prefix", "KEY", "KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeWithEnvPrefix(tt.key); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}

	envKeyReplacer = UnderlineReplacer
	tests = []T{
		{"with replacer", "Key-Key", "KEY_KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeWithEnvPrefix(tt.key); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}

	SetEnvPrefix("test")
	tests = []T{
		{"with prefix and replacer", "Key-key", "TEST_KEY_KEY"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := mergeWithEnvPrefix(tt.key); got != tt.want {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
