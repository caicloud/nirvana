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

package v1

import (
	"context"

	"github.com/caicloud/nirvana/definition"
)

// Desc is global descriptor of api
var Desc = definition.Descriptor{
	Path:        "/api/v1",
	Definitions: []definition.Definition{},
	Consumes:    []string{"application/json"},
	Produces:    []string{"application/json"},
	Children: []definition.Descriptor{
		{
			Path: "/{target1}/{target2}",
			Definitions: []definition.Definition{
				{
					Method:   definition.Create,
					Function: Handle,
					Parameters: []definition.Parameter{
						{
							Source: definition.Header,
							Name:   "User-Agent",
						},
						{
							Source: definition.Query,
							Name:   "target1",
						},
						{
							Source:  definition.Query,
							Name:    "target2",
							Default: false,
						},
						{
							Source: definition.Body,
							Name:   "app",
						},
					},
					Results: []definition.Result{
						{Destination: definition.Data},
						{Destination: definition.Error},
					},
				},
			},
		},
	},
}

// Application defines application api model
type Application struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Target    string `json:"target"`
	Target1   int    `json:"target2"`
	Target2   bool   `json:"target1"`
}

// Handle handles http request
func Handle(ctx context.Context, userAgent string, target1 int, target2 bool, app *Application) (*Application, error) {
	app.Target = userAgent
	app.Target1 = target1
	app.Target2 = target2
	return app, nil
}
