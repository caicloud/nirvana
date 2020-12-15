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

package golang

import (
	"reflect"
	"testing"

	"github.com/caicloud/nirvana/utils/api"
)

func TestDeterministicClientPackageAlias(t *testing.T) {
	const (
		rounds  = 20
		rootPkg = "github.com/caicloud/nirvana"
	)
	types := map[api.TypeName]*api.Type{
		api.TypeName("github.com/caicloud/repo/v1.Foo"): {
			Name:    "Foo",
			PkgPath: "github.com/caicloud/repo/v1",
			Kind:    reflect.Struct,
		},
		api.TypeName("github.com/caicloud/repo/pkg/v1.Bar"): {
			Name:    "Bar",
			PkgPath: "github.com/caicloud/repo/pkg/v1",
			Kind:    reflect.Struct,
		},
	}
	namer, err := newTypeNamer(rootPkg, types)
	if err != nil {
		t.Fatal(err)
	}
	expected := namer.packages

	for i := 0; i < rounds; i++ {
		namer, err := newTypeNamer(rootPkg, types)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(expected, namer.packages) {
			t.Fatalf("Package alias changed:\nExpected: %#v\nGot: %#v\n", expected, namer.packages)
		}
	}
}
