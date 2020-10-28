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

package project

import (
	"testing"
)

func TestPackageForPath(t *testing.T) {
	goPaths = []string{"/go"}
	goSrcPaths = []string{"/go/src"}

	tests := []struct {
		name      string
		directory string
		wanted    string
	}{
		{
			name:      "in GOPATH test",
			directory: "/go/src/github.com/caicloud/test1",
			wanted:    "github.com/caicloud/test1",
		},
		{
			name:      "not in GOPATH test",
			directory: "/usr/test",
			wanted:    "test",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got, _ := PackageForPath(tt.directory); got != tt.wanted {
				t.Errorf("%s PackageForPath() = %v, want %v", tt.name, got, tt.wanted)
			}
		})
	}
}
