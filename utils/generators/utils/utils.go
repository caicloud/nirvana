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
