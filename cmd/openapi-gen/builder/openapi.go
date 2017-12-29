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

package builder

import (
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/openapi/common"
	"github.com/go-openapi/spec"
)

const (
	openAPIVersion = "2.0"
)

type openAPI struct {
	config  *common.Config
	swagger *spec.Swagger

	definitions map[string]common.OpenAPIDefinition
}

// BuildOpenAPISpec builds open api spec and return swagger struct
func BuildOpenAPISpec(descriptor *definition.Descriptor, c *common.Config) (*spec.Swagger, error) {
	o, err := newOpenAPI(c)
	if err != nil {
		return nil, err
	}
	if err := o.init(descriptor); err != nil {
		return nil, err
	}

	return o.swagger, nil
}

func newOpenAPI(c *common.Config) (*openAPI, error) {
	o := openAPI{
		config: c,
		swagger: &spec.Swagger{
			SwaggerProps: spec.SwaggerProps{
				Swagger:     openAPIVersion,
				Definitions: spec.Definitions{},
				Paths:       &spec.Paths{Paths: map[string]spec.PathItem{}},
				Info:        c.Info,
			},
		},
	}

	if o.config.GetDefinitionName == nil {
		o.config.GetDefinitionName = func(name string) (string, spec.Extensions) {
			return name[strings.LastIndex(name, "/")+1:], nil
		}
	}
	o.definitions = o.config.GetDefinitions(func(name string) spec.Ref {
		defName, _ := o.config.GetDefinitionName(name)
		return spec.MustCreateRef("#/definitions/" + common.EscapeJSONPointer(defName))
	})

	return &o, nil
}

func (o *openAPI) init(descriptor *definition.Descriptor) error {
	return o.buildDescriptor("", nil, nil, descriptor)
}

func (o *openAPI) buildDescriptor(path string, consumes, produces []string, descriptor *definition.Descriptor) error {
	path = strings.TrimSuffix(path, "/") + descriptor.Path
	specPaths, err := o.buildPathItems(path, consumes, produces, descriptor.Definitions)
	if err != nil {
		return err
	}
	o.swagger.SwaggerProps.Paths.Paths = specPaths
	c, p := descriptor.Consumes, descriptor.Produces

	if c == nil {
		c = consumes
	}
	if p == nil {
		p = produces
	}

	for _, d := range descriptor.Children {
		if err := o.buildDescriptor(descriptor.Path, c, p, &d); err != nil {
			return err
		}
	}
	return nil
}

func (o *openAPI) buildPathItems(path string, consumes, produces []string, defs []definition.Definition) (map[string]spec.PathItem, error) {
	specPaths := map[string]spec.PathItem{}
	for _, def := range defs {
		pathItem := spec.PathItem{}
		op, err := o.buildOperation(&def, consumes, produces, service.HTTPCodeFor(def.Method))
		if err != nil {
			return nil, err
		}
		switch def.Method {
		case definition.List:
			pathItem.Get = op
		case definition.Get:
			pathItem.Get = op
		case definition.Create:
			pathItem.Post = op
		case definition.Update:
			pathItem.Put = op
		case definition.Patch:
			pathItem.Patch = op
		case definition.Delete:
			pathItem.Delete = op
		case definition.AsyncCreate:
			pathItem.Post = op
		case definition.AsyncUpdate:
			pathItem.Put = op
		case definition.AsyncPatch:
			pathItem.Patch = op
		case definition.AsyncDelete:
			pathItem.Delete = op
		}
		specPaths[path] = pathItem
	}
	return specPaths, nil
}

func (o *openAPI) buildOperation(def *definition.Definition, consumes, produces []string, defaultStatusCode int) (*spec.Operation, error) {
	op := spec.Operation{
		OperationProps: spec.OperationProps{
			Description: def.Description,
			Consumes:    consumes,
			Produces:    produces,
			// TODO(liubog2008): should support 'ws' scheme
			// NOTE(liubog2008): wss and https will be supported
			// in external doc
			Schemes: nil,
			// TODO(liubog2008): support tags
			Tags: nil,
			// NOTE(liubog2008): specify difference between
			// description and summary?
			// e.g. summary will be description which has
			// at most 50 char?
			Summary:      "",
			ExternalDocs: nil,
			// NOTE(liubog2008): ID should be defined
			// TODO(liubog2008): use function name of handler
			ID: "",
			// TODO(liubog2008): deprecated should be supported
			Deprecated: false,
			// NOTE(liubog2008): define global security
			Security: nil,
		},
	}
	ps, err := o.buildParameters(def.Function, def.Parameters)
	if err != nil {
		return nil, err
	}
	op.Parameters = ps
	resp, err := o.buildResponses(def.Function, def.Results, defaultStatusCode)
	if err != nil {
		return nil, err
	}
	op.Responses = resp
	return &op, nil
}

func (o *openAPI) buildParameters(handler interface{}, params []definition.Parameter) ([]spec.Parameter, error) {
	specParams := []spec.Parameter{}

	ht := reflect.TypeOf(handler)
	if ht.Kind() != reflect.Func {
		return nil, fmt.Errorf("Handler is not a function")
	}

	// In default config of nirvana, the first parameter of handler
	// should be context. But the context is no useless for openapi.
	// So ignore it.
	const ignoreFirstContext = true

	for i, param := range params {
		parameterIndex := i
		if ignoreFirstContext {
			parameterIndex++
		}
		specParam := spec.Parameter{
			ParamProps: spec.ParamProps{
				Name:        param.Name,
				Description: param.Description,
				Required:    false,
			},
		}
		specParam.Default = param.Default
		if specParam.Default == nil {
			specParam.Required = true
		}
		switch param.Source {
		case definition.Path:
			specParam.In = "path"
		case definition.Query:
			specParam.In = "query"
		case definition.Header:
			specParam.In = "header"
		case definition.Form:
			specParam.In = "formData"
		case definition.File:
			specParam.In = "formData"
		case definition.Body:
			specParam.In = "body"
			if !specParam.Required {
				return nil, fmt.Errorf("body param %v MUST be required", param.Name)
			}
			dataType := ht.In(parameterIndex)
			s, err := o.toSchema(dataType)
			if err != nil {
				return nil, err
			}
			specParam.Schema = s
			specParams = append(specParams, specParam)
			continue
		case definition.Prefab:
			// NOTE(liubog2008): skip prefab
		case definition.Auto:
			// NOTE(liubog2008): handle by ref
		}
		// TODO(liubog2008): support array type of parameter, e.g. []string
		dataType := ht.In(parameterIndex)
		if openAPIType, openAPIFormat := common.GetOpenAPITypeFormat(getCanonicalizeTypeName(dataType)); openAPIType != "" {
			specParam.SimpleSchema = spec.SimpleSchema{
				Type:   openAPIType,
				Format: openAPIFormat,
			}
		}
		specParams = append(specParams, specParam)
	}
	return specParams, nil
}

func (o *openAPI) buildResponses(handler interface{}, results []definition.Result, defaultStatusCode int) (*spec.Responses, error) {
	respsProps := spec.ResponsesProps{
		StatusCodeResponses: map[int]spec.Response{},
	}

	rightResponse := spec.Response{
		ResponseProps: spec.ResponseProps{
		// Headers: map[string]spec.Header{},
		},
	}

	ht := reflect.TypeOf(handler)
	if ht.Kind() != reflect.Func {
		return nil, fmt.Errorf("Handler is not a function")
	}

	haveData, haveError := false, false
	for i, res := range results {
		switch res.Destination {
		case definition.Data:
			if haveData {
				return nil, fmt.Errorf("Only one data type result is allowed")
			}
			rightResponse.Description = res.Description
			if defaultStatusCode != http.StatusNoContent {
				dataType := ht.Out(i)
				s, err := o.toSchema(dataType)
				if err != nil {
					return nil, err
				}
				rightResponse.Schema = s
			}
			haveData = true
		case definition.Error:
			if haveError {
				return nil, fmt.Errorf("Only one error type result is allowed")
			}
			// How to generate specified error response
			// which is defined in handler
			// TODO(liubog2008):
			haveError = true
		case definition.Meta:
			// for _, k := range res.Headers {
			//     rightResponse.Headers[k] = spec.Header{
			//         // TODO(liubog2008): now only string is supported
			//         SimpleSchema: spec.SimpleSchema{
			//             Type:   "string",
			//             Format: "",
			//         },
			//         // TODO(liubog2008): add header description for each one
			//         HeaderProps: spec.HeaderProps{
			//             Description: res.Description,
			//         },
			//     }
			// }
		}
	}
	if !haveData && defaultStatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("MUST specify a data result if status code is not 204")
	}
	respsProps.StatusCodeResponses[defaultStatusCode] = rightResponse
	return &spec.Responses{
		ResponsesProps: respsProps,
	}, nil
}

func (o *openAPI) toSchema(t reflect.Type) (*spec.Schema, error) {
	if openAPIType, openAPIFormat := common.GetOpenAPITypeFormat(getCanonicalizeTypeName(t)); openAPIType != "" {
		return &spec.Schema{
			SchemaProps: spec.SchemaProps{
				Type:   []string{openAPIType},
				Format: openAPIFormat,
			},
		}, nil
	}
	ref, err := o.buildDefinitionForType(t)
	if err != nil {
		return nil, err
	}
	return &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: *ref,
		},
	}, nil
}

func getCanonicalizeTypeName(t reflect.Type) string {
	if t.PkgPath() == "" {
		return t.Name()
	}
	path := t.PkgPath()
	if strings.Contains(path, "/vendor/") {
		path = path[strings.Index(path, "/vendor/")+len("/vendor/"):]
	}
	return path + "." + t.Name()
}

func (o *openAPI) buildDefinitionForType(t reflect.Type) (*spec.Ref, error) {
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	name := getCanonicalizeTypeName(t)
	if err := o.buildDefinitionRecursively(name); err != nil {
		return nil, err
	}
	defName, _ := o.config.GetDefinitionName(name)
	ref, err := spec.NewRef("#/definitions/" + common.EscapeJSONPointer(defName))
	if err != nil {
		return nil, err
	}
	return &ref, nil
}

func (o *openAPI) buildDefinitionRecursively(name string) error {
	uniqueName, extensions := o.config.GetDefinitionName(name)
	if _, ok := o.swagger.Definitions[uniqueName]; ok {
		return nil
	}
	item, ok := o.definitions[name]
	if ok {
		schema := spec.Schema{
			VendorExtensible:   item.Schema.VendorExtensible,
			SchemaProps:        item.Schema.SchemaProps,
			SwaggerSchemaProps: item.Schema.SwaggerSchemaProps,
		}
		if extensions != nil {
			if schema.Extensions == nil {
				schema.Extensions = spec.Extensions{}
			}
			for k, v := range extensions {
				schema.Extensions[k] = v
			}
		}
		o.swagger.Definitions[uniqueName] = schema
		for _, v := range item.Dependencies {
			if err := o.buildDefinitionRecursively(v); err != nil {
				return err
			}
		}
		return nil
	}
	return fmt.Errorf("cannot find model definition for %v", name)
}
