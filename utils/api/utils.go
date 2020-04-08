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

package api

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/caicloud/nirvana/definition"
)

// PathForVersion returns the path of API file.
func PathForVersion(root, version string) string {
	return path.Join(root, fmt.Sprintf("/api.%s.json", version))
}

// WriteFiles writes the API data into files.
func WriteFiles(output string, apis map[string][]byte) error {
	dir, err := filepath.Abs(output)
	if err != nil {
		return err
	}
	if err = os.MkdirAll(dir, 0775); err != nil {
		return err
	}
	for version, data := range apis {
		file := filepath.Join(dir, fmt.Sprintf("api.%s.json", version))
		if err = ioutil.WriteFile(file, data, 0664); err != nil {
			return err
		}
	}
	return nil
}

// GenSwaggerPageData generates a Swagger UI page based on the given API files.
func GenSwaggerPageData(root string, versions []string) ([]byte, error) {
	index := `
<!-- HTML for static distribution bundle build -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
    <title>Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui.css" >
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@3.25.0/favicon-32x32.png" sizes="32x32" />
    <link rel="icon" type="image/png" href="https://unpkg.com/swagger-ui-dist@3.25.0/favicon-16x16.png" sizes="16x16" />
    <style>
      html
      {
        box-sizing: border-box;
        overflow: -moz-scrollbars-vertical;
        overflow-y: scroll;
      }
      *,
      *:before,
      *:after
      {
        box-sizing: inherit;
      }
      body
      {
        margin:0;
        background: #fafafa;
      }
    </style>
  </head>
  <body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-bundle.js"> </script>
    <script src="https://unpkg.com/swagger-ui-dist@3.25.0/swagger-ui-standalone-preset.js"> </script>
    <script>
	  // list of APIS
      var apis = [
        {{ range $i, $v := . }}
        {
            name: '{{ $v.Name }}',
            url: '{{ $v.Path }}'
        },
		{{ end }}
      ];
    window.onload = function() {
      // Begin Swagger UI call region
      const ui = SwaggerUIBundle({
        urls: apis,
        dom_id: '#swagger-ui',
        deepLinking: true,
        presets: [
          SwaggerUIBundle.presets.apis,
          SwaggerUIStandalonePreset
        ],
        plugins: [
          SwaggerUIBundle.plugins.DownloadUrl
        ],
        layout: "StandaloneLayout"
      })
      // End Swagger UI call region
      window.ui = ui
    }
  </script>
  </body>
</html>
`
	tmpl, err := template.New("index.html").Parse(index)
	if err != nil {
		return nil, err
	}
	data := make([]struct {
		Name string
		Path string
	}, 0, len(versions))
	for _, v := range versions {
		data = append(data, struct {
			Name string
			Path string
		}{v, PathForVersion(root, v)})
	}
	buf := bytes.NewBuffer(nil)
	if err = tmpl.Execute(buf, data); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// DescriptorForData generates a Descriptor for API docs page.
func DescriptorForData(path string, data []byte, contentType string) definition.Descriptor {
	return definition.Descriptor{
		Path: path,
		Definitions: []definition.Definition{
			{
				Method:   definition.Get,
				Consumes: []string{definition.MIMENone},
				Produces: []string{contentType},
				Function: func(context.Context) ([]byte, error) {
					return data, nil
				},
				Parameters: []definition.Parameter{},
				Results:    definition.DataErrorResults(""),
			},
		},
	}
}
