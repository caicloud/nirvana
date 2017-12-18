package builder

import (
	"reflect"
	"testing"

	"github.com/caicloud/nirvana/cmd/openapi-gen/common"
	"github.com/go-openapi/spec"
	"github.com/stretchr/testify/assert"
)

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
		contentName: common.OpenAPIDefinition{
			Schema:       contentSchema,
			Dependencies: []string{},
		},
		inputName: common.OpenAPIDefinition{
			Schema:       inputSchema,
			Dependencies: []string{contentName},
		},
		outputName: common.OpenAPIDefinition{
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

func TestToSchema(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.NoError(t, err, "can new openAPI successfully")

	var (
		simple   string
		inputPtr *TestInput

		definitions = spec.Definitions{
			"builder.TestInput":   inputSchema,
			"builder.TestContent": contentSchema,
		}
	)

	simpleSchema, err := o.toSchema(reflect.TypeOf(simple))
	assert.NoError(t, err, "can convert simple type to schema")
	assert.Equal(t, &spec.Schema{
		SchemaProps: spec.SchemaProps{
			Type:   []string{"string"},
			Format: "",
		},
	}, simpleSchema, "can convert string type to schema")

	complexSchema, err := o.toSchema(reflect.TypeOf(inputPtr))
	assert.NoError(t, err, "can convert complex type to schema")

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
	assert.NoError(t, err, "can new openAPI successfully")

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
	assert.NoError(t, err, "can build Definition for type")
	assert.Equal(t, spec.MustCreateRef("#/definitions/builder.TestInput"), *inputRef, "input ref for building definition should be equal")

	outputRef, err := o.buildDefinitionForType(reflect.TypeOf(outputPtr))
	assert.NoError(t, err, "can build Definition for type")
	assert.Equal(t, spec.MustCreateRef("#/definitions/builder.TestOutput"), *outputRef, "output ref for building definition should be equal")

	assert.Equal(t, definitions, o.swagger.Definitions, "definitions should be equal")
}

func TestBuildDefinitionRecursively(t *testing.T) {
	c := newConfig()
	o, err := newOpenAPI(c)
	assert.NoError(t, err, "can new openAPI successfully")

	definitions := spec.Definitions{
		"builder.TestInput":   inputSchema,
		"builder.TestContent": contentSchema,
	}

	assert.NoError(t, o.buildDefinitionRecursively(inputName), "can build definition recursively for TestInput type")
	assert.Equal(t, definitions, o.swagger.Definitions, "can add input and content schema when build for TestInput type")

	assert.NoError(t, o.buildDefinitionRecursively(contentName), "can build definition recursively for TetContent type")
	assert.Equal(t, definitions, o.swagger.Definitions, "can ignore content schema when build for TestContent type")

	definitions["builder.TestOutput"] = outputSchema

	assert.NoError(t, o.buildDefinitionRecursively(outputName), "can build definition recursively for TestOutput type")
	assert.Equal(t, definitions, o.swagger.Definitions, "can add output schema when build for TestOutput type")
}
