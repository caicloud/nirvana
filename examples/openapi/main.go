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

package main

import (
	"encoding/json"
	"os"

	"github.com/caicloud/nirvana/cmd/openapi-gen/builder"
	"github.com/caicloud/nirvana/examples/openapi/api"
	"github.com/caicloud/nirvana/examples/openapi/pkg/api/v1"
	"github.com/caicloud/nirvana/utils/openapi/common"
	"github.com/go-openapi/spec"
)

func main() {
	// v1.Desc need to be changed to your own Desc struct.
	swagger, err := builder.BuildOpenAPISpec(&v1.Desc, &common.Config{
		Info: &spec.Info{
			InfoProps: spec.InfoProps{
				Title:       "openapi example",
				Description: "This is an example of openapi-gen for nirvana",
				Contact: &spec.ContactInfo{
					Name:  "caicloud",
					URL:   "https://caicloud.io",
					Email: "caicloud@caicloud.io",
				},
				License: &spec.License{
					Name: "Apache License, Version 2.0",
					URL:  "http://www.apache.org/licenses/LICENSE-2.0",
				},
				Version: "v1.0.0",
			},
		},
		// Your own generated struct.
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
