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

	"gopkg.in/go-playground/validator.v9"
)

var std = validator.New()

// Var returns an operator to validate a single variable using tag style validation and allows passing of contextual
// validation information via context.Context.
func Var(tag string) func(ctx context.Context, object interface{}) (interface{}, error) {
	return func(ctx context.Context, object interface{}) (interface{}, error) {
		return object, std.VarCtx(ctx, object, tag)
	}
}

// VarWithValue returns an operator to validate a single variable, against another variable/field's value using tag style validation and
// allows passing of contextual validation information via context.Context.
func VarWithValue(other interface{}, tag string) func(ctx context.Context, object interface{}) (interface{}, error) {
	return func(ctx context.Context, object interface{}) (interface{}, error) {
		return object, std.VarWithValueCtx(ctx, object, other, tag)
	}
}

// Struct returns an operator to validate a structs exposed fields, and automatically validates nested structs, unless otherwise specified
// and also allows passing of context.Context for contextual validation information.
func Struct() func(ctx context.Context, object interface{}) (interface{}, error) {
	return func(ctx context.Context, object interface{}) (interface{}, error) {
		return object, std.StructCtx(ctx, object)
	}
}
