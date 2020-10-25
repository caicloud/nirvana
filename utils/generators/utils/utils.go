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

package utils

import (
	"encoding/json"

	"github.com/caicloud/nirvana/utils/api"
	"github.com/caicloud/nirvana/utils/generators/swagger"
	"github.com/caicloud/nirvana/utils/project"
)

// GenSwaggerData generates swagger data for definitions.
func GenSwaggerData(config *project.Config, definitions *api.Definitions) (map[string][]byte, error) {
	generator := swagger.NewDefaultGenerator(config, definitions)
	swaggers, err := generator.Generate()
	if err != nil {
		return nil, err
	}

	files := make(map[string][]byte, len(swaggers))
	for filename, s := range swaggers {
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return nil, err
		}

		files[filename] = data
	}
	return files, nil
}
