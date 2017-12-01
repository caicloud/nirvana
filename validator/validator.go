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
