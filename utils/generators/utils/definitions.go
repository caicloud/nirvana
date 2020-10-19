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

package utils

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/caicloud/nirvana/utils/api"
	"github.com/caicloud/nirvana/utils/project"
)

// NameReplacer uses to replace non-alphabetic and non-numeric characters.
var NameReplacer = regexp.MustCompile(`[^a-zA-Z0-9]`)

// Version defines the version data structure.
type Version struct {
	Module string
	Name   string
}

// DefinitionWithVersion defines Definitions and the version data of it.
type DefinitionWithVersion struct {
	Defs *api.Definitions
	Version
}

func encodeVersion(v string) string {
	return strings.ToLower(NameReplacer.ReplaceAllString(v, ""))
}

// SplitDefinitions splits definitions by versions.
func SplitDefinitions(apis *api.Definitions, config *project.Config) ([]DefinitionWithVersion, error) {
	definitions := make([]DefinitionWithVersion, 0, len(config.Versions))
	matched := map[string]bool{}
	for _, version := range config.Versions {
		defs := apis.Subset(func(path string, def *api.Definition) bool {
			for _, rule := range version.PathRules {
				if rule.Check(path) {
					return true
				}
			}
			return false
		})
		if defs != nil {
			for path := range defs.Definitions {
				matched[path] = true
			}
			definitions = append(definitions, DefinitionWithVersion{
				Defs: defs,
				Version: Version{
					Module: encodeVersion(version.Module),
					Name:   encodeVersion(version.Name),
				},
			})
		}
	}
	if len(matched) != len(apis.Definitions) {
		paths := make([]string, 0, len(apis.Definitions))
		for path := range apis.Definitions {
			if !matched[path] {
				paths = append(paths, path)
			}
		}
		return nil, fmt.Errorf("can't match any version for %v", paths)
	}
	return definitions, nil
}
