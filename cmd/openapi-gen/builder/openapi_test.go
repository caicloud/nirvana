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
	"context"
	"reflect"
	"testing"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/utils/openapi/common"
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

var def = definition.Definition{
	Method:   definition.Create,
	Function: Handler,
	Parameters: []definition.Parameter{
		{
			Source: definition.Header,
			Name:   "User-Agent",
		},
		{
			Source: definition.Query,
			Name:   "start",
		},
		{
			Source: definition.Body,
			Name:   "app",
		},
	},
	Results: []definition.Result{
		{
			Destination: definition.Meta,
			// Headers: map[string]string{
			//     "xxx": "X-Xxx",
			// },
		},
		{Destination: definition.Data},
		{Destination: definition.Error},
	},
}

var descriptor = definition.Descriptor{
	Path:        "/api/v1",
	Definitions: []definition.Definition{},
	Consumes:    []string{"application/json"},
	Produces:    []string{"application/json"},
	Children: []definition.Descriptor{
		{
			Path: "/input",
			Definitions: []definition.Definition{
				def,
			},
		},
	},
}

var (
	parameters = []spec.Parameter{
		{
			ParamProps: spec.ParamProps{
				Name:     "User-Agent",
				Required: true,
				In:       "header",
			},
			SimpleSchema: spec.SimpleSchema{
				Type:   "string",
				Format: "",
			},
		},
		{
			ParamProps: spec.ParamProps{
				Name:     "start",
				Required: true,
				In:       "query",
			},
			SimpleSchema: spec.SimpleSchema{
				Type:   "integer",
				Format: "int",
			},
		},
		{
			ParamProps: spec.ParamProps{
				Name:     "app",
				Required: true,
				In:       "body",
				Schema: &spec.Schema{
					SchemaProps: spec.SchemaProps{
						Ref: spec.MustCreateRef("#/definitions/builder.TestInput"),
					},
				},
			},
		},
	}
	responses = &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			StatusCodeResponses: map[int]spec.Response{
				201: {
					ResponseProps: spec.ResponseProps{
						// Headers: map[string]spec.Header{
						//     "X-Xxx": {
						//         SimpleSchema: spec.SimpleSchema{
						//             Type:   "string",
						//             Format: "",
						//         },
						//     },
						// },
						Schema: &spec.Schema{
							SchemaProps: spec.SchemaProps{
								Ref: spec.MustCreateRef("#/definitions/builder.TestOutput"),
							},
						},
					},
				},
			},
		},
	}
	operation = &spec.Operation{
		OperationProps: spec.OperationProps{
			Description: def.Description,
			Consumes:    []string{"application/json"},
			Produces:    []string{"application/json"},
			Parameters:  parameters,
			Responses:   responses,
		},
	}
)

func Handler(ctx context.Context, agent string, start int, input *TestInput) (map[string]string, *TestOutput, error) {
	return nil, nil, nil
}

// TestContent ...
type TestContent struct {
	// Name of the content
	Name     string   `json:"name"`
	Comments []string `json:"comments"`
}

// TestInput ...
type TestInput struct {
	// Name of the input
	Name string `json:"name,omitempty"`
	// ID of the input
	ID   int      `json:"id,omitempty"`
	Tags []string `json:"tags,omitempty"`

	TestContent TestContent `json:"content"`
}

// TestOutput ...
type TestOutput struct {
	// Name of the output
	Name string `json:"name,omitempty"`
	// Number of outputs
	Count int `json:"count,omitempty"`
}

var (
	contentSchema = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "TestContent ...",
			Properties: map[string]spec.Schema{
				"name": {
					SchemaProps: spec.SchemaProps{
						Description: "Name of the content",
						Type:        []string{"string"},
						Format:      "",
					},
				},
				"commnets": {
					SchemaProps: spec.SchemaProps{
						Description: "",
						Type:        []string{"array"},
						Items: &spec.SchemaOrArray{
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Type:   []string{"string"},
									Format: "",
								},
							},
						},
					},
				},
			},
		},
	}
	inputSchema = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "Test input",
			Properties: map[string]spec.Schema{
				"name": {
					SchemaProps: spec.SchemaProps{
						Description: "Name of the input",
						Type:        []string{"string"},
						Format:      "",
					},
				},
				"id": {
					SchemaProps: spec.SchemaProps{
						Description: "ID of the input",
						Type:        []string{"integer"},
						Format:      "int32",
					},
				},
				"tags": {
					SchemaProps: spec.SchemaProps{
						Description: "",
						Type:        []string{"array"},
						Items: &spec.SchemaOrArray{
							Schema: &spec.Schema{
								SchemaProps: spec.SchemaProps{
									Type:   []string{"string"},
									Format: "",
								},
							},
						},
					},
				},
			},
		},
		VendorExtensible: spec.VendorExtensible{
			Extensions: spec.Extensions{"x-test": "test"},
		},
	}

	outputSchema = spec.Schema{
		SchemaProps: spec.SchemaProps{
			Description: "Test output",
			Properties: map[string]spec.Schema{
				"name": {
					SchemaProps: spec.SchemaProps{
						Description: "Name of the output",
						Type:        []string{"string"},
						Format:      "",
					},
				},
				"count": {
					SchemaProps: spec.SchemaProps{
						Description: "Number of outputs",
						Type:        []string{"integer"},
						Format:      "int32",
					},
				},
			},
		},
	}
)

func getOpenAPIDefinitions(ref common.ReferenceCallback) map[string]common.OpenAPIDefinition {
	return map[string]common.OpenAPIDefinition{
		contentName: {
			Schema:       contentSchema,
			Dependencies: []string{},
		},
		inputName: {
			Schema:       inputSchema,
			Dependencies: []string{contentName},
		},
		outputName: {
			Schema:       outputSchema,
			Dependencies: []string{},
		},
	}
}

const (
	contentName = "github.com/caicloud/nirvana/cmd/openapi-gen/builder.TestContent"
	inputName   = "github.com/caicloud/nirvana/cmd/openapi-gen/builder.TestInput"
	outputName  = "github.com/caicloud/nirvana/cmd/openapi-gen/builder.TestOutput"
)

func newConfig() *common.Config {
	return &common.Config{
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       "TestAPI",
				Description: "Test API",
				Version:     "unversioned",
			},
		},
		GetDefinitions: getOpenAPIDefinitions,
	}
}

func TestBuildPathItem(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.Nil(t, err, "can new openAPI successfully")

	pathItem, err := o.buildPathItem(descriptor.Consumes, descriptor.Produces, descriptor.Children[0].Definitions)
	assert.Nil(t, err)

	assert.Equal(t, &spec.PathItem{
		PathItemProps: spec.PathItemProps{
			Post: operation,
		},
	}, pathItem)
}

func TestBuildOperation(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.Nil(t, err, "can new openAPI successfully")

	op, err := o.buildOperation(&def, descriptor.Consumes, descriptor.Produces, 201)
	assert.Nil(t, err)

	assert.Equal(t, operation, op)
}

func TestBuildParameters(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.Nil(t, err, "can new openAPI successfully")

	ps, err := o.buildParameters(def.Function, def.Parameters)
	assert.Nil(t, err)

	assert.Equal(t, parameters, ps)
}

func TestBuildResponses(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.Nil(t, err, "can new openAPI successfully")

	resp, err := o.buildResponses(def.Function, def.Results, 201)
	assert.Nil(t, err)

	assert.Equal(t, responses, resp, "can build response")
}

func TestToSchema(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.Nil(t, err, "can new openAPI successfully")

	var (
		simple   string
		inputPtr *TestInput

		definitions = spec.Definitions{
			"builder.TestInput":   inputSchema,
			"builder.TestContent": contentSchema,
		}
	)

	simpleSchema, err := o.toSchema(reflect.TypeOf(simple))
	assert.Nil(t, err, "can convert simple type to schema")
	assert.Equal(t, &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:   []string{"string"},
			Format: "",
		},
	}, simpleSchema, "can convert string type to schema")

	complexSchema, err := o.toSchema(reflect.TypeOf(inputPtr))
	assert.Nil(t, err, "can convert complex type to schema")

	assert.Equal(t, &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Ref: spec.MustCreateRef("#/definitions/builder.TestInput"),
		},
	}, complexSchema, "can convert *TestInput type to schema")

	assert.Equal(t, definitions, o.swagger.Definitions, "definitions should be equal")
}

func TestGetCanonializeTypeName(t *testing.T) {
	var s spec.Schema

	typ := reflect.TypeOf(s)
	path := typ.PkgPath()
	assert.Equal(t, "github.com/caicloud/nirvana/vendor/github.com/go-openapi/spec", path, "pkg path of type contains vendor")
	assert.Equal(t, "github.com/go-openapi/spec.Schema", getCanonicalizeTypeName(typ), "can get canonialize type name")
}

func TestBuildDefinitionForType(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.Nil(t, err, "can new openAPI successfully")

	var (
		inputStruct TestInput
		outputPtr   *TestOutput

		definitions = spec.Definitions{
			"builder.TestInput":   inputSchema,
			"builder.TestContent": contentSchema,
			"builder.TestOutput":  outputSchema,
		}
	)

	inputRef, err := o.buildDefinitionForType(reflect.TypeOf(inputStruct))
	assert.Nil(t, err, "can build Definition for type")
	assert.Equal(t, spec.MustCreateRef("#/definitions/builder.TestInput"), *inputRef, "input ref for building definition should be equal")

	outputRef, err := o.buildDefinitionForType(reflect.TypeOf(outputPtr))
	assert.Nil(t, err, "can build Definition for type")
	assert.Equal(t, spec.MustCreateRef("#/definitions/builder.TestOutput"), *outputRef, "output ref for building definition should be equal")

	assert.Equal(t, definitions, o.swagger.Definitions, "definitions should be equal")
}

func TestBuildDefinitionRecursively(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.Nil(t, err, "can new openAPI successfully")

	definitions := spec.Definitions{
		"builder.TestInput":   inputSchema,
		"builder.TestContent": contentSchema,
	}

	assert.Nil(t, o.buildDefinitionRecursively(inputName), "can build definition recursively for TestInput type")
	assert.Equal(t, definitions, o.swagger.Definitions, "can add input and content schema when build for TestInput type")

	assert.Nil(t, o.buildDefinitionRecursively(contentName), "can build definition recursively for TetContent type")
	assert.Equal(t, definitions, o.swagger.Definitions, "can ignore content schema when build for TestContent type")

	definitions["builder.TestOutput"] = outputSchema

	assert.Nil(t, o.buildDefinitionRecursively(outputName), "can build definition recursively for TestOutput type")
	assert.Equal(t, definitions, o.swagger.Definitions, "can add output schema when build for TestOutput type")

	assert.Error(t, o.buildDefinitionRecursively("builder.XXX"), "cannot find model definition for builder.XXX")
}
