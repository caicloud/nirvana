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

package main

import (
	"context"
	"io/ioutil"
	"mime/multipart"

	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
)

func Upload(ctx context.Context, file multipart.File) (string, error) {
	if file == nil {
		return "no content", nil
	}
	f, err := ioutil.ReadAll(file)
	if err != nil {
		return "", nil
	}
	return string(f), nil
}

func main() {
	descriptors := []definition.Descriptor{
		{
			Path:        "/required",
			Description: "Upload API",
			Definitions: []definition.Definition{
				{
					Method:   definition.Create,
					Function: Upload,
					Consumes: []string{definition.MIMEAll},
					Produces: []string{definition.MIMEAll},
					Parameters: []definition.Parameter{
						{
							Source: definition.File,
							Name:   "file",
						},
					},
					Results: definition.DataErrorResults(""),
				},
			},
		},
		{
			Path:        "/optional",
			Description: "Upload API",
			Definitions: []definition.Definition{
				{
					Method:   definition.Create,
					Function: Upload,
					Consumes: []string{definition.MIMEAll},
					Produces: []string{definition.MIMEAll},
					Parameters: []definition.Parameter{
						{
							Source:   definition.File,
							Name:     "file",
							Optional: true,
						},
					},
					Results: definition.DataErrorResults(""),
				},
			},
		},
	}
	cmd := config.NewDefaultNirvanaCommand()
	if err := cmd.Execute(descriptors...); err != nil {
		log.Fatal(err)
	}
}
