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

package validator

import (
	"context"
	"reflect"
	"testing"
)

func TestVar(t *testing.T) {
	v, err := Var("gt=0,lt=10")(context.Background(), 5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, 5) {
		t.Fatalf("get %v want %v", v, 5)
	}
}

func TestVarWithValue(t *testing.T) {
	v, err := VarWithValue("other", "eqcsfield")(context.Background(), "other")
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, "other") {
		t.Fatalf("get %v want %v", v, "other")
	}
}

func TestStruct(t *testing.T) {
	var me = struct {
		Name string `json:"name" validate:"required,printascii"`
	}{"233"}

	v, err := Struct()(context.Background(), me)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, me) {
		t.Fatalf("get %v want %v", v, me)
	}
}
