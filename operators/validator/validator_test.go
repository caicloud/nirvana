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

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
)

func TestVar(t *testing.T) {
	op := Int("gt=0,lt=10")
	v, err := op.Operate(context.Background(), "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, 5) {
		t.Fatalf("get %v want %v", v, 5)
	}
	validator := op.(Validator)
	if validator.Category() != CategoryVar ||
		validator.Tag() != "gt=0,lt=10" {
		t.Fatalf("%+v", validator)
	}
}

func TestStruct(t *testing.T) {
	var me = struct {
		Name string `json:"name" validate:"required,printascii"`
	}{"233"}
	op := Struct(me)
	v, err := op.Operate(context.Background(), "", me)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, me) {
		t.Fatalf("got %v want %v", v, me)
	}
	validator := op.(Validator)
	if validator.Category() != CategoryStruct {
		t.Fatalf("%+v", validator)
	}
}

func TestNewCustom(t *testing.T) {
	var anje = struct {
		Name string
	}{"anje"}
	op := NewCustom(definition.OperatorFunc(OperatorKind, func(ctx context.Context, field string, object *struct{ Name string }) (*struct{ Name string }, error) {
		if object.Name != "anje" {
			return nil, errors.BadRequest.Build("badRequest:name", "${name} wrong").Error("anje")
		}
		return object, nil
	}), "check name")
	v, err := op.Operate(context.Background(), "", &anje)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v, &anje) {
		t.Fatalf("got %v want %v", v, anje)
	}
	validator := op.(Validator)
	if validator.Category() != CategoryCustom ||
		validator.Description() != "check name" {
		t.Fatalf("%+v", validator)
	}
}
