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

package definition

// RPCDescriptor describes a descriptor for API definition in RPC style.
type RPCDescriptor struct {
	// Path describes url path prefix for all RPCActions, default: "/".
	Path string
	// Description describes the usage of the path.
	Description string
	// Middlewares contains path middlewares.
	Middlewares []Middleware
	// Tags indicates tags of current definitions and child definitions.
	// It will override parent descriptor's tags.
	Tags []string
	// Consumes indicates content types that current definitions
	// and child definitions can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates content types that current definitions
	// and child definitions can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Actions contain actions in this descriptor. These actions will inherit the Middlewares, Tags, Consumes, Produces
	// of the descriptor if values in the action are not specified.
	Actions []RPCAction
}

// RPCAction defines an API handler in RPC style.
type RPCAction struct {
	// Version defines the version this API belongs to.
	// Need to use time format, eg: 2020-10-10
	Version string
	// Name defines the Action name.
	Name string
	// Consumes indicates how many content types the handler can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates how many content types the handler can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Tags indicates tags of the API handler.
	// It will override parent descriptor's tags.
	Tags []string
	// ErrorProduces is used to generate data for error. If this field is empty,
	// it means that this field equals to Produces.
	// In some cases, succeessful data and error data should be generated in
	// different ways.
	ErrorProduces []string
	// Function is a function handler. It must be func type.
	Function interface{}
	// Parameters describes function parameters.
	Parameters []Parameter
	// Results describes function retrun values.
	Results []Result
	// Description describes the API handler.
	Description string
	// Example contains the example for the API handler.
	Example interface{}
}
