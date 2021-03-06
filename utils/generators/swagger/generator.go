/*
Copyright 2018 Caicloud Authors

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

package swagger

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/api"
	"github.com/caicloud/nirvana/utils/project"

	"github.com/go-openapi/spec"
)

var defaultSourceMapping = map[definition.Source]string{
	definition.Path:   "path",
	definition.Query:  "query",
	definition.Header: "header",
	definition.Form:   "formData",
	definition.File:   "formData",
	definition.Body:   "body",
	definition.Prefab: "",
}

var defaultDestinationMapping = map[definition.Destination]string{
	definition.Meta:  "header",
	definition.Data:  "body",
	definition.Error: "",
}

// Generator is for generating swagger specifications.
type Generator struct {
	config             *project.Config
	apis               *api.Definitions
	schemas            map[string]*spec.Schema
	schemaMappings     map[api.TypeName]*spec.Schema
	paths              map[string]*spec.PathItem
	sourceMapping      map[definition.Source]string
	destinationMapping map[definition.Destination]string
}

// NewDefaultGenerator creates a swagger generator with default mappings.
func NewDefaultGenerator(
	config *project.Config,
	apis *api.Definitions,
) *Generator {
	return NewGenerator(config, apis, nil, nil)
}

// NewGenerator creates a swagger generator.
func NewGenerator(
	config *project.Config,
	apis *api.Definitions,
	sourceMapping map[definition.Source]string,
	destinationMapping map[definition.Destination]string,
) *Generator {
	g := &Generator{
		config:         config,
		apis:           apis,
		schemas:        map[string]*spec.Schema{},
		schemaMappings: map[api.TypeName]*spec.Schema{},
		paths:          map[string]*spec.PathItem{},
	}
	if sourceMapping == nil {
		g.sourceMapping = defaultSourceMapping
	}
	if destinationMapping == nil {
		g.destinationMapping = defaultDestinationMapping
	}
	return g
}

// Generate generates swagger specifications.
func (g *Generator) Generate() (map[string]spec.Swagger, error) {
	g.parseSchemas()
	g.parsePaths()

	swaggers := make(map[string]spec.Swagger, len(g.config.Versions))
	for _, version := range g.config.Versions {
		title := fmt.Sprintln(g.config.Project, "APIs")
		description := g.config.Description
		if description != "" && version.Description != "" {
			description += "\n" + version.Description
		}
		schemes := version.Schemes
		if len(schemes) <= 0 {
			schemes = g.config.Schemes
		}
		host := version.Host
		if host == "" {
			host = g.config.Host
		}
		contact := version.Contact
		if contact == nil {
			contact = g.config.Contact
		}
		basePath := version.BasePath
		if basePath == "" {
			basePath = g.config.BasePath
		}

		swagger := g.buildSwaggerInfo(
			title, version.Name, description,
			schemes, host, basePath, contact,
			version.PathRules,
		)
		var filename string
		if version.Module != "" {
			filename = strings.ToLower(version.Module) + "." + strings.ToLower(version.Name)
		} else {
			filename = strings.ToLower(version.Name)
		}
		swaggers[filename] = *swagger
	}

	if len(swaggers) <= 0 {
		swagger := g.buildSwaggerInfo(
			g.config.Project, "unknown", g.config.Description,
			g.config.Schemes, g.config.Host, g.config.BasePath, g.config.Contact,
			nil,
		)
		swaggers["unknown"] = *swagger
	}
	return swaggers, nil
}

func (g *Generator) buildSwaggerInfo(
	title, version, description string,
	schemes []string,
	host string,
	basePath string,
	contact *project.Contact,
	rules []project.PathRule,
) *spec.Swagger {
	swagger := &spec.Swagger{}
	swagger.Swagger = "2.0"
	swagger.Schemes = schemes
	swagger.Host = host
	swagger.BasePath = basePath
	swagger.Info = &spec.Info{}
	swagger.Info.Title = title
	swagger.Info.Version = version
	swagger.Info.Description = g.escapeNewline(description)
	if contact != nil {
		swagger.Info.Contact = &spec.ContactInfo{
			ContactInfoProps: spec.ContactInfoProps{
				Name:  contact.Name,
				Email: contact.Email,
			},
		}
	}
	swagger.Definitions = spec.Definitions{}
	swagger.Paths = &spec.Paths{
		Paths: map[string]spec.PathItem{},
	}
	for path, definition := range g.schemas {
		swagger.Definitions[path] = *definition
	}
	if len(rules) > 0 {
		for path, item := range g.paths {
			for _, rule := range rules {
				if replacedPath := rule.Replace(path); replacedPath != "" {
					swagger.Paths.Paths[replacedPath] = *item
					break
				}
			}
		}
	} else {
		for path, item := range g.paths {
			swagger.Paths.Paths[path] = *item
		}
	}
	return swagger
}

func (g *Generator) parseSchemas() {
	for _, typ := range g.apis.Types {
		g.schemaForType(typ)
	}
}

func (g *Generator) schemaForType(typ *api.Type) *spec.Schema {
	schema, ok := g.schemaMappings[typ.TypeName()]
	if !ok {
		switch typ.Kind {
		case reflect.Array, reflect.Slice:
			elem := g.schemaForTypeName(typ.Elem)
			if elem == nil {
				break
			}
			schema = spec.ArrayProperty(elem)
			schema.Title = "[]" + elem.Title
		case reflect.Ptr:
			schema = g.schemaForTypeName(typ.Elem)
		case reflect.Map:
			keySchema := g.schemaForTypeName(typ.Key)
			if keySchema == nil {
				break
			}
			elemSchema := g.schemaForTypeName(typ.Elem)
			if elemSchema == nil {
				break
			}
			schema = spec.MapProperty(elemSchema)
			schema.Title = fmt.Sprintf("map[%s]%s", keySchema.Title, elemSchema.Title)
			schema.Items = &spec.SchemaOrArray{
				Schema: keySchema,
			}
		case reflect.Struct:
			if typ.TypeName() == "time.Time" {
				schema = spec.DateTimeProperty()
				schema.Title = "Time"
			} else {
				schema = g.schemaForStruct(typ)
			}
		case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16,
			reflect.Int32, reflect.Int64, reflect.Uint, reflect.Uint8,
			reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
			reflect.Float32, reflect.Float64, reflect.String, reflect.Interface:
			schema = g.schemaForBasicType(typ)
		}
		if schema != nil {
			g.schemaMappings[typ.TypeName()] = schema
		}
	}
	return g.copySchema(schema)
}

func (g *Generator) setSchemaForStruct(schema *spec.Schema, typ *api.Type) {
	for _, field := range typ.Fields {
		if !field.Anonymous && field.Name[0] >= 'a' && field.Name[0] <= 'z' {
			// ignore unexported fields
			continue
		}
		jsontag := strings.TrimSpace(field.Tag.Get("json"))
		if jsontag == "-" {
			continue
		}
		// type Test struct {
		//     T0 `json:",inline"`  // inline
		//     T1                   // inline
		//     T2 T2                // not inline
		//     T3 `json:"T3"`       // not inline
		// }
		if jsontag == ",inline" || (jsontag == "" && field.Anonymous) {
			fieldType, ok := g.apis.Types[field.Type]
			if !ok {
				continue
			}
			g.setSchemaForStruct(schema, fieldType)
			continue
		}
		fieldSchema := g.schemaForTypeName(field.Type)
		if fieldSchema == nil {
			// Ignore invalid field.
			continue
		}
		name := jsontag
		if comma := strings.Index(jsontag, ","); comma > 0 {
			name = strings.TrimSpace(jsontag[:comma])
		}
		if name == "" {
			name = field.Name
		}
		fieldSchema.Description = g.escapeNewline(field.Comments)
		schema.SetProperty(name, *fieldSchema)
	}
}

func (g *Generator) schemaForStruct(typ *api.Type) *spec.Schema {
	typeName := typ.TypeName()
	schema, ok := g.schemaMappings[typeName]
	if ok {
		return schema
	}
	// To prevent recursive struct.
	key := strings.Replace(string(typeName), "/", "_", -1)
	ref := spec.RefSchema("#/definitions/" + key)
	ref.Title = typ.Name
	g.schemaMappings[typeName] = ref

	schema = &spec.Schema{}
	schema.Title = ref.Title
	g.setSchemaForStruct(schema, typ)
	g.schemas[key] = schema
	return ref
}

func (g *Generator) schemaForBasicType(typ *api.Type) *spec.Schema {
	var types = map[reflect.Kind][]string{
		reflect.Bool:    {"boolean", "bool"},
		reflect.Int:     {"number", "int"},
		reflect.Int8:    {"number", "int8"},
		reflect.Int16:   {"number", "int16"},
		reflect.Int32:   {"number", "int32"},
		reflect.Int64:   {"number", "int64"},
		reflect.Uint:    {"number", "uint"},
		reflect.Uint8:   {"number", "uint8"},
		reflect.Uint16:  {"number", "uint16"},
		reflect.Uint32:  {"number", "uint32"},
		reflect.Uint64:  {"number", "uint64"},
		reflect.Uintptr: {"number", "uintptr"},
		reflect.Float32: {"number", "float32"},
		reflect.Float64: {"number", "float64"},
		reflect.String:  {"string", "string"},

		// Interface is special. It can be anything.
		reflect.Interface: {"undefined", "interface{}"},
	}
	formats, ok := types[typ.Kind]
	if !ok {
		return nil
	}
	schema := &spec.Schema{SchemaProps: spec.SchemaProps{
		Type: []string{formats[0]}, Format: formats[1]}}
	schema.Title = typ.Name
	return schema
}

func (g *Generator) schemaForTypeName(name api.TypeName) *spec.Schema {
	typ, ok := g.apis.Types[name]
	if !ok {
		return nil
	}
	return g.schemaForType(typ)
}

func (g *Generator) copySchema(source *spec.Schema) *spec.Schema {
	if source == nil {
		return nil
	}
	dest := *source
	return &dest
}

func (g *Generator) parsePaths() {
	for path, defs := range g.apis.Definitions {
		operations := map[string][]*spec.Operation{}
		for _, def := range defs {
			op := g.operationFor(&def)
			ops := operations[def.HTTPMethod]
			ops = append(ops, op)
			operations[def.HTTPMethod] = ops
		}
		for method, ops := range operations {
			for i, op := range ops {
				itemPath := path
				if i > 0 {
					itemPath = fmt.Sprintf("%s [%d]", path, i)
				}
				item := g.paths[itemPath]
				if item == nil {
					item = &spec.PathItem{}
					g.paths[itemPath] = item
				}
				switch method {
				case http.MethodGet:
					item.Get = op
				case http.MethodHead:
					item.Head = op
				case http.MethodPost:
					item.Post = op
				case http.MethodPut:
					item.Put = op
				case http.MethodPatch:
					item.Patch = op
				case http.MethodDelete:
					item.Delete = op
				case http.MethodOptions:
					item.Options = op
				case string(definition.Any):
					item.Get = op
					item.Head = op
					item.Post = op
					item.Put = op
					item.Patch = op
					item.Delete = op
					item.Options = op
				default:
					continue
				}
			}
		}
	}
}

func (g *Generator) operationFor(def *api.Definition) *spec.Operation {
	operation := &spec.Operation{}
	consumes := map[string]bool{}
	for _, c := range def.Consumes {
		if !consumes[c] {
			consumes[c] = true
			operation.Consumes = append(operation.Consumes, c)
		}
	}
	produces := map[string]bool{}
	for _, p := range def.Produces {
		if !produces[p] {
			produces[p] = true
			operation.Produces = append(operation.Produces, p)
		}
	}
	tags := map[string]bool{}
	for _, t := range def.Tags {
		if !tags[t] {
			tags[t] = true
			operation.Tags = append(operation.Tags, t)
		}
	}
	for _, p := range def.ErrorProduces {
		if !produces[p] {
			produces[p] = true
			operation.Produces = append(operation.Produces, p)
		}
	}
	operation.Summary = def.Summary
	if operation.Summary == "" {
		// Use function name as API summary.
		typ, ok := g.apis.Types[def.Function]
		if ok {
			operation.Summary = typ.Name
		}
		if operation.Summary == "" {
			operation.Summary = "Unknown API"
		}
	}
	operation.Description = def.Description
	if operation.Description == "" {
		// Use function comments as API description.
		typ, ok := g.apis.Types[def.Function]
		if ok {
			operation.Description = typ.Comments
		}
	}
	operation.Description = g.escapeNewline(operation.Description)
	for _, param := range def.Parameters {
		parameters := g.generateParameter(&param)
		if len(parameters) > 0 {
			operation.Parameters = append(operation.Parameters, parameters...)
		}
	}
	operation.Responses = &spec.Responses{
		ResponsesProps: spec.ResponsesProps{
			StatusCodeResponses: map[int]spec.Response{
				def.HTTPCode: *g.generateResponse(def.Results, def.Example),
			},
		},
	}
	return operation
}

func (g *Generator) generateParameter(param *api.Parameter) []spec.Parameter {
	if param.Source == definition.Auto {
		return g.generateAutoParameter(param.Type)
	}
	source := g.sourceMapping[param.Source]
	if source == "" {
		return nil
	}
	schema := g.schemaForTypeName(param.Type)
	parameter := spec.Parameter{
		ParamProps: spec.ParamProps{
			Name:        param.Name,
			Description: g.escapeNewline(param.Description),
			Schema:      schema,
			In:          source,
			Required:    !param.Optional,
		},
	}
	if param.Default != nil {
		parameter.WithDefault(param.Default)
	}
	body := "body"
	if parameter.In != body {
		// Only body parameter can hold a schema. Other parameters uses type
		// and format.
		parameter.Type = schema.Type[0]
		parameter.Format = schema.Format
		if parameter.Type == "array" {
			// Array is a special type. It needs additional configs.
			// CollectionFormat has two valid values: csv, multi.
			// But we don't known which one should be used. So unknown.
			parameter.CollectionFormat = "unknown"
			parameter.Items = &spec.Items{}
			parameter.Items.Type = schema.Items.Schema.Type[0]
			parameter.Items.Format = schema.Items.Schema.Format
		}
		parameter.Schema = nil
		parameter.SimpleSchema.Example = param.Example
	} else {
		// add parameter name for body, it required by swagger ui,
		// cause api.Parameter.Name is always nil when In is body
		parameter.Name = body
		// handle the common case that the parameter is a struct, set the example on the schema of its ref
		ref := schema.Ref.String()
		if schema.Type == nil && ref != "" {
			// "#/definitions/xxx" --> "xxx"
			k := ref[len("#/definitions/"):]
			if v, ok := g.schemas[k]; ok {
				v.WithExample(param.Example)
			}
		} else {
			// handle cases where the parameters are array or map etc, set the example directly on the current schema
			schema.WithExample(param.Example)
		}
	}
	return []spec.Parameter{parameter}
}

func (g *Generator) generateAutoParameter(typ api.TypeName) []spec.Parameter {
	structType, ok := g.apis.Types[typ]
	if !ok {
		return nil
	}
	if structType.Kind == reflect.Ptr {
		return g.generateAutoParameter(structType.Elem)
	}
	if structType.Kind != reflect.Struct {
		return nil
	}
	return g.enum(structType)
}

var converters = map[string]service.Converter{
	"bool":       service.ConvertToBool,
	"int":        service.ConvertToInt,
	"int8":       service.ConvertToInt8,
	"int16":      service.ConvertToInt16,
	"int32":      service.ConvertToInt32,
	"int64":      service.ConvertToInt64,
	"uint":       service.ConvertToUint,
	"uint8":      service.ConvertToUint8,
	"uint16":     service.ConvertToUint16,
	"uint32":     service.ConvertToUint32,
	"uint64":     service.ConvertToUint64,
	"float32":    service.ConvertToFloat32,
	"float64":    service.ConvertToFloat64,
	"string":     service.ConvertToString,
	"time.Time":  service.ConvertToTime,
	"*bool":      service.ConvertToBoolP,
	"*int":       service.ConvertToIntP,
	"*int8":      service.ConvertToInt8P,
	"*int16":     service.ConvertToInt16P,
	"*int32":     service.ConvertToInt32P,
	"*int64":     service.ConvertToInt64P,
	"*uint":      service.ConvertToUintP,
	"*uint8":     service.ConvertToUint8P,
	"*uint16":    service.ConvertToUint16P,
	"*uint32":    service.ConvertToUint32P,
	"*uint64":    service.ConvertToUint64P,
	"*float32":   service.ConvertToFloat32P,
	"*float64":   service.ConvertToFloat64P,
	"*string":    service.ConvertToStringP,
	"*time.Time": service.ConvertToTimeP,
	"[]bool":     service.ConvertToBoolSlice,
	"[]int":      service.ConvertToIntSlice,
	"[]float64":  service.ConvertToFloat64Slice,
	"[]string":   service.ConvertToStringSlice,
}

func (g *Generator) enum(typ *api.Type) []spec.Parameter {
	results := make([]spec.Parameter, 0, len(typ.Fields))
	for _, field := range typ.Fields {
		tag := field.Tag.Get("source")
		parameters := []spec.Parameter(nil)
		if tag != "" {
			source, name, apc, err := service.ParseAutoParameterTag(tag)
			rawDefaultValue, defaultExist := apc.Get(service.AutoParameterConfigKeyDefaultValue)
			var defaultValue []byte
			if c := converters[string(field.Type)]; defaultExist && c != nil {
				// we don't find a good way to handle the default value of non-basic types,
				// so for now the default value of those types are always empty
				v, _ := c(context.TODO(), []string{rawDefaultValue})
				defaultValue, _ = json.Marshal(v)
			}
			_, optional := apc.Get(service.AutoParameterConfigKeyOptional)

			if err == nil {
				parameters = g.generateParameter(&api.Parameter{
					Source:      source,
					Name:        name,
					Description: g.escapeNewline(field.Comments),
					Type:        field.Type,
					Default:     defaultValue,
					Optional:    optional || defaultExist,
				})
			}
		} else {
			fieldType, ok := g.apis.Types[field.Type]
			if ok && fieldType.Kind == reflect.Struct {
				parameters = g.enum(fieldType)
			}
		}
		if len(parameters) > 0 {
			results = append(results, parameters...)
		}
	}
	return results
}

func parseDestination(d definition.Destination) definition.Destination {
	switch {
	// for the custom Destination
	case strings.Contains(string(d), string(definition.Meta)):
		return definition.Meta
	case strings.Contains(string(d), string(definition.Data)):
		return definition.Data
	case strings.Contains(string(d), string(definition.Error)):
		return definition.Error
	default:
		return d
	}
}

func (g *Generator) generateResponse(results []api.Result, example interface{}) *spec.Response {
	response := &spec.Response{}
	for _, result := range results {
		switch g.destinationMapping[parseDestination(result.Destination)] {
		case "body":
			response.Description = g.escapeNewline(result.Description)
			schema := g.schemaForTypeName(result.Type)
			// responses.xx.schema should NOT have additional properties
			// additionalProperty: title
			schema.Title = ""
			response.Schema = schema
		}
	}
	response.AddExample("application/json", example)
	if response.Schema == nil && response.Description == "" {
		response.Description = "No Content"
	}
	return response
}

func (g *Generator) escapeNewline(content string) string {
	return strings.Replace(strings.TrimSpace(content), "\n", "<br/>", -1)
}
