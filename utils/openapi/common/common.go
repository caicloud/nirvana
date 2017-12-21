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

package common

import (
	"strings"

	"github.com/go-openapi/spec"
)

// NOTE(liubog2008): it is copied from k8s.io/kube-openapi/common
// TODO(liubog2008): Config should be changed to fit our cases in the future

// OpenAPIDefinition describes single type. Normally these definitions are auto-generated using gen-openapi.
type OpenAPIDefinition struct {
	Schema       spec.Schema
	Dependencies []string
}

// ReferenceCallback is defined to get ref from path
type ReferenceCallback func(path string) spec.Ref

// GetOpenAPIDefinitions is collection of all definitions.
type GetOpenAPIDefinitions func(ReferenceCallback) map[string]OpenAPIDefinition

// OpenAPIDefinitionGetter gets openAPI definitions for a given type. If a type implements this interface,
// the definition returned by it will be used, otherwise the auto-generated definitions will be used. See
// GetOpenAPITypeFormat for more information about trade-offs of using this interface or GetOpenAPITypeFormat method when
// possible.
type OpenAPIDefinitionGetter interface {
	OpenAPIDefinition() *OpenAPIDefinition
}

// Config is set of configuration for openAPI spec generation.
type Config struct {
	// List of supported protocols such as https, http, etc.
	// NOTE(liubog2008): it is not supported now
	// NOTE(liubog2008): ProtocolList should be defined for each operation
	// but not in a global config
	// ProtocolList []string

	// Info is general information about the API.
	// TODO(liubog2008): it should be generated from comments
	Info *spec.Info

	// DefaultResponse will be used if an operation does not have any responses listed. It
	// will show up as ... "responses" : {"default" : $DefaultResponse} in the spec.
	// NOTE(liubog2008): it is not supported now
	// DefaultResponse *spec.Response

	// CommonResponses will be added as a response to all operation specs. This is a good place to add common
	// responses such as authorization failed.
	// NOTE(liubog2008): it is not supported now
	// CommonResponses map[int]spec.Response

	// OpenAPIDefinitions should provide definition for all models used by routes. Failure to provide this map
	// or any of the models will result in spec generation failure.
	GetDefinitions GetOpenAPIDefinitions

	// GetDefinitionName returns a friendly name for a definition base on the serving path. parameter `name` is the full name of the definition.
	// It is an optional function to customize model names.
	GetDefinitionName func(name string) (string, spec.Extensions)

	// PostProcessSpec runs after the spec is ready to serve. It allows a final modification to the spec before serving.
	// NOTE(liubog2008): use to convert external and internal docs
	// NOTE(liubog2008): it is not supported now
	// PostProcessSpec func(*spec.Swagger) (*spec.Swagger, error)

	// SecurityDefinitions is list of all security definitions for OpenAPI service. If this is not nil, the user of config
	// is responsible to provide DefaultSecurity and (maybe) add unauthorized response to CommonResponses.
	// NOTE(liubog2008): it is not supported now
	// SecurityDefinitions *spec.SecurityDefinitions

	// DefaultSecurity for all operations. This will pass as spec.SwaggerProps.Security to OpenAPI.
	// For most cases, this will be list of acceptable definitions in SecurityDefinitions.
	// NOTE(liubog2008): it is not supported now
	// DefaultSecurity []map[string][]string
}

var schemaTypeFormatMap = map[string][]string{
	"int":    {"integer", "int"},
	"uint":   {"integer", "uint"},
	"int8":   {"integer", "int8"},
	"uint8":  {"integer", "uint8"},
	"int16":  {"integer", "int16"},
	"uint16": {"integer", "uint16"},
	"int32":  {"integer", "int32"},
	"uint32": {"integer", "uint32"},
	// NOTE(liubog2008): js and JSON only support int up to 2^53
	// type of uint64 need to be changed to string
	"int64":  {"integer", "int64"},
	"uint64": {"integer", "uint64"},

	"byte": {"integer", "uint8"},
	// base64 encoded characters
	"[]byte": {"string", "byte"},

	"float64":   {"number", "double"},
	"float32":   {"number", "float"},
	"bool":      {"boolean", ""},
	"time.Time": {"string", "date-time"},
	"string":    {"string", ""},
	"integer":   {"integer", ""},
	"number":    {"number", ""},
	"boolean":   {"boolean", ""},

	"interface{}": {"object", ""},
}

// GetOpenAPITypeFormat is a reference for converting go (or any custom type) to a simple open API type,format pair. There are
// two ways to customize spec for a type. If you add it here, a type will be converted to a simple type and the type
// comment (the comment that is added before type definition) will be lost. The spec will still have the property
// comment. The second way is to implement OpenAPIDefinitionGetter interface. That function can customize the spec (so
// the spec does not need to be simple type,format) or can even return a simple type,format (e.g. IntOrString). For simple
// type formats, the benefit of adding OpenAPIDefinitionGetter interface is to keep both type and property documentation.
// Example:
// type Sample struct {
//      ...
//      // port of the server
//      port IntOrString
//      ...
// }
// // IntOrString documentation...
// type IntOrString { ... }
//
// Adding IntOrString to this function:
// "port" : {
//           format:      "string",
//           type:        "int-or-string",
//           Description: "port of the server"
// }
//
// Implement OpenAPIDefinitionGetter for IntOrString:
//
// "port" : {
//           $Ref:    "#/definitions/IntOrString"
//           Description: "port of the server"
// }
// ...
// definitions:
// {
//           "IntOrString": {
//                     format:      "string",
//                     type:        "int-or-string",
//                     Description: "IntOrString documentation..."    // new
//           }
// }
//
func GetOpenAPITypeFormat(typeName string) (string, string) {
	mapped, ok := schemaTypeFormatMap[typeName]
	if !ok {
		return "", ""
	}
	return mapped[0], mapped[1]
}

// EscapeJSONPointer encode json pointer by rfc6901
func EscapeJSONPointer(p string) string {
	// Escaping reference name using rfc6901
	p = strings.Replace(p, "~", "~0", -1)
	p = strings.Replace(p, "/", "~1", -1)
	return p
}
