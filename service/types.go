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

package service

import (
	"net/http"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
)

// APIStyle defines the API style type of the builder.
type APIStyle string

const (
	// APIStyleREST represents the RESTful API style.
	APIStyleREST APIStyle = "rest"
	// APIStyleRPC represents the RPC API style.
	APIStyleRPC APIStyle = "rpc"
)

// Builder builds service.
type Builder interface {
	// Logger returns logger of builder.
	Logger() log.Logger
	// SetLogger sets logger to server.
	SetLogger(logger log.Logger)
	// Modifier returns modifier of builder.
	Modifier() DefinitionModifier
	// SetModifier sets definition modifier.
	SetModifier(m DefinitionModifier)
	// Filters returns all request filters.
	Filters() []Filter
	// AddFilter add filters to filter requests.
	AddFilter(filters ...Filter)
	// AddDescriptor adds descriptors to router.
	AddDescriptor(descriptors ...interface{}) error
	// Definitions returns all definitions. If a modifier exists, it will be executed.
	Definitions() map[string][]definition.Definition
	// APIStyle returns the API style of this builder.
	APIStyle() APIStyle
	// Build builds a service to handle request.
	Build() (Service, error)
}

// Service handles HTTP requests.
//
// Workflow:
//            Service.ServeHTTP()
//          ----------------------
//          ↓                    ↑
// |-----Filters------|          ↑
//          ↓                    ↑
// |-Router/RPC Action Match-|   ↑
//          ↓                    ↑
// |-------------Middlewares------------|
//          ↓                    ↑
// |-------------Executor---------------|
//          ↓                    ↑
// |-ParameterGenerators-|-DestinationHandlers-|
//          ↓                    ↑
// |------------User Function-----------|
type Service interface {
	http.Handler
}
