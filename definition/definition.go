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

package definition

import (
	"context"
	"reflect"
)

// Chain contains all subsequent actions.
type Chain interface {
	// Continue continues to execute the next subsequent actions.
	Continue(context.Context) error
}

// Middleware describes the form of middlewares. If you want to
// carry on, call Chain.Continue() and pass the context.
type Middleware func(context.Context, Chain) error

type Operator interface {
	Operate(ctx context.Context, object interface{}) (interface{}, error)
}

type OperatorFunc func(ctx context.Context, object interface{}) (interface{}, error)

func (f OperatorFunc) Operate(ctx context.Context, object interface{}) (interface{}, error) {
	return f(ctx, object)
}

// Method is an alternative of HTTP method. It's more clearer than HTTP method.
// A definition method binds a certain HTTP method and a success status code.
type Method string

const (
	// List binds to http.MethodGet and code http.StatusOK(200).
	List Method = "List"
	// Get binds to http.MethodGet and code http.StatusOK(200).
	Get Method = "Get"
	// Create binds to http.MethodPost and code http.StatusCreated(201).
	Create Method = "Create"
	// Update binds to http.MethodPut and code http.StatusOK(200).
	Update Method = "Update"
	// Patch binds to http.MethodPatch and code http.StatusOK(200).
	Patch Method = "Patch"
	// Delete binds to http.MethodDelete and code http.StatusNoContent(204).
	Delete Method = "Delete"
	// AsyncCreate binds to http.MethodPost and code http.StatusAccepted(202).
	AsyncCreate Method = "AsyncCreate"
	// AsyncUpdate binds to http.MethodPut and code http.StatusAccepted(202).
	AsyncUpdate Method = "AsyncUpdate"
	// AsyncPatch binds to http.MethodPatch and code http.StatusAccepted(202).
	AsyncPatch Method = "AsyncPatch"
	// AsyncDelete binds to http.MethodDelete and code http.StatusAccepted(202).
	AsyncDelete Method = "AsyncDelete"
)

// Source indicates which place a value is from.
type Source string

const (
	// Path means value is from URL path.
	Path Source = "Path"
	// Query means value is from URL query string.
	Query Source = "Query"
	// Header means value is from request header.
	Header Source = "Header"
	// Form means value is from request body and content type must be
	// "application/x-www-form-urlencoded" and "multipart/form-data".
	Form Source = "Form"
	// File means value is from request body and content type must be
	// "multipart/form-data".
	File Source = "File"
	// Body means value is from request body.
	Body Source = "Body"
	// Prefab means value is from a prefab generator.
	// May a prefab will combine many data to generate value.
	Prefab Source = "Prefab"
	// Auto identifies a struct and generate field values by field tag.
	//
	// Tag name is "source". Its value format is "Source,Name".
	//
	// ex.
	// type Example struct {
	//     Start       int    `source:"Query,start"`
	//     ContentType string `source:"Header,Content-Type"`
	// }
	Auto Source = "Auto"
)

// Type indicates the target type to place function results.
type Type string

const (
	// Meta means result will be set into  header of response.
	Meta Type = "Meta"
	// Data means result will be set into body of response.
	Data Type = "Data"
	// Error means the result is an error and should be treat specially.
	Error Type = "Error"
)

// Example is just an example.
type Example struct {
	Description string
	Instance    interface{}
}

// Parameter describes a function parameter.
type Parameter struct {
	// Source is the parameter value generated from.
	Source Source
	// Name is the name to get value from a request.
	// ex. a query name, a header key, etc.
	Name string
	// Type is used to override function parameter type.
	// If you want to override the type in function parameter, you can specify it here.
	// When the type is same as function parameter type, the field can be ignored.
	// If the type is not compatible with function parameter type, you must add
	// a operator to convert it or it will panic.
	Type reflect.Type
	// Default value is used when a request does not provide a value
	// for the parameter.
	// If parameter type is set, the default value must can be assigned to that type.
	// If parameter type is not set, the default value must can be assigned to
	// function parameter type.
	Default interface{}
	// Operators can modify and validate the target value.
	// Parameter value is passed to the first operator, then
	// previous operator's result is as next operator's parameter.
	// The result of last operator will be passed to target function.
	Operators []Operator
	// Description describes the parameter.
	Description string
	// Examples contains many examples for the parameter.
	Examples []Example
}

// Result describes how to handle a result from function results.
type Result struct {
	// Type is the target for the result. Different types make different behavior.
	Type Type
	// Headers is a map from key used by result to http Header
	// Only used when Type is Meta
	Headers map[string]string
	// Operators can modify the result value.
	// Result value is passed to the first operator, then
	// previous operator's result is as next operator's parameter.
	// The result of last operator will be passed to type handler.
	Operators []Operator
	// Description describes the result.
	Description string
	// Examples contains many examples for the result.
	Examples []Example
}

// Definition defines an API handler.
type Definition struct {
	// Method is definition method.
	Method Method
	// Consumes indicates how many content types the handler can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates how many content types the handler can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Function is a function handler. It must be func type.
	Function interface{}
	// Parameters describes function parameters.
	Parameters []Parameter
	// Results describes function retrun values.
	Results []Result
	// Description describes the API handler.
	Description string
	// Examples contains many examples for the API handler.
	Examples []Example
}

// Descriptor describes a descriptor for API definitions.
type Descriptor struct {
	// Path is the url path. It will inherit parent's path.
	//
	// If parent path is "/api/v1", current is "/some",
	// It means current definitions handles "/api/v1/some".
	Path string
	// Consumes indicates content types children handlers can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates content types children handlers can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Middlewares contains path middlewares.
	Middlewares []Middleware
	// Definitions contains handlers for current path.
	Definitions []Definition
	// Children is used to place sub-descriptors.
	Children []Descriptor
	// Description describes the usage of the path.
	Description string
}
