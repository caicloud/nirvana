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

// Package main is definition of api
package main

import (
	"context"
	"encoding/json"
	"os"

	"github.com/caicloud/nirvana/cmd/openapi-gen/builder"
	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/examples/getting-started/openapi/pkg/api"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/operators/validator"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/utils/openapi/common"
	"github.com/go-openapi/spec"
)

var echo = definition.Descriptor{
	Path:        "/echo",
	Description: "Echo API",
	Definitions: []definition.Definition{
		{
			Method:   definition.Get,
			Function: Echo,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "msg",
					Description: "Corresponding to the second parameter",
					Operators:   []definition.Operator{validator.String("gt=10")},
				},
			},
			Results: []definition.Result{
				{
					Destination: definition.Data,
					Description: "Corresponding to the first result",
				},
				{
					Destination: definition.Error,
					Description: "Corresponding to the second result",
				},
			},
		},
		{
			Method:   definition.Create,
			Function: EchoV2,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Body,
					Name:        "msg",
					Description: "Corresponding to the second parameter",
					Operators:   []definition.Operator{validator.Struct(&api.Message{})},
				},
			},
			Results: []definition.Result{
				{
					Destination: definition.Data,
					Description: "Corresponding to the first result",
				},
				{
					Destination: definition.Error,
					Description: "Corresponding to the second result",
				},
			},
		},
	},
}

// API function.
func Echo(ctx context.Context, msg string) (string, error) {
	return msg, nil
}

var message map[string]string

// API function.
func EchoV2(ctx context.Context, msg *api.Message) (string, error) {
	log.Infof("receive message from %v\n", msg.Name)
	message[msg.Name] = msg.Message
	return msg.Message, nil
}

func buildOpenAPI() {
	swagger, err := builder.BuildOpenAPISpec(&echo, &common.Config{
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       "echo server openapi",
				Description: "This is open API documentation of echo server",
				Contact: &spec.ContactInfo{
					Name: "nirvana",
					URL:  "https://gonirvana.io",
				},
				License: &spec.License{
					Name: "Apache License, Version 2.0",
					URL:  "http://www.apache.org/licenses/LICENSE-2.0",
				},
				Version: "v0.1.0",
			},
		},
		GetDefinitions: api.GetOpenAPIDefinitions,
	})
	if err != nil {
		panic(err)
	}
	encoder := json.NewEncoder(os.Stdout)
	if err := encoder.Encode(swagger); err != nil {
		panic(err)
	}
}

func main() {
	buildOpenAPI()
	message = make(map[string]string)

	cmd := config.NewDefaultNirvanaCommand()
	cmd.EnablePlugin(&metrics.Option{Path: "/metrics"})
	if err := cmd.Execute(echo); err != nil {
		log.Fatal(err)
	}
}
