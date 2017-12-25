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

	"github.com/caicloud/nirvana/definition"
)

var std = validator.New()

// Kind distinguishs validation type based on different Validator implementation.
// This type is intended to be used by doc gen.
type Kind string

// Kinds for doc gen. Users should not care about these.
const (
	KindCustom Kind = "Custom"
	KindVar    Kind = "Var"
	KindStruct Kind = "Struct"
)

// Info shows Validator information which will be used for doc gen.
type Info struct {
	Kind        Kind
	Tag         string
	Description string
}

// Infoer can retrieve the validator info used for doc gen.
type Infoer interface {
	Info() Info
}

type varValidator struct {
	tag string
}

func (v varValidator) Operate(ctx context.Context, object interface{}) (interface{}, error) {
	return object, std.VarCtx(ctx, object, v.tag)
}

func (v varValidator) Info() Info {
	return Info{
		Kind: KindVar,
		Tag:  v.tag,
	}
}

type structValidator struct{}

func (v structValidator) Operate(ctx context.Context, object interface{}) (interface{}, error) {
	return object, std.StructCtx(ctx, object)
}

func (v structValidator) Info() Info {
	return Info{
		Kind: KindStruct,
	}
}

type customValidator struct {
	operator    definition.Operator
	description string
}

func (v customValidator) Operate(ctx context.Context, object interface{}) (interface{}, error) {
	return v.operator.Operate(ctx, object)
}

func (v customValidator) Info() Info {
	return Info{
		Kind:        KindCustom,
		Description: v.description,
	}
}

// NewCustom calls f for validation, using description for doc gen.
// User should only do custom validation in f.
// Validations which can be done by Var and Struct should be done in another Operator.
// Exp:
// []definition.Operator{validator.Struct(), NewCustom(f,"custom validation description")}
func NewCustom(operator definition.Operator, description string) definition.Operator {
	return customValidator{
		operator:    operator,
		description: description,
	}
}

// Var returns an operator to validate a single variable using tag style validation and allows passing of contextual
// validation information via context.Context.
func Var(tag string) definition.Operator {
	return varValidator{tag: tag}
}

// Struct returns an operator to validate a structs exposed fields, and automatically validates nested structs, unless otherwise specified
// and also allows passing of context.Context for contextual validation information.
func Struct() definition.Operator {
	return structValidator{}
}
