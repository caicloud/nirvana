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

package generators

import (
	"reflect"
	"strings"

	"k8s.io/gengo/types"
)

const (
	openAPITag    = "nirvana:openapi"
	tagValueFalse = "false"
	tagValueTrue  = "true"
)

func getOpenAPITagValue(comments []string) []string {
	return types.ExtractCommentTags("+", comments)[openAPITag]
}

func hasOpenAPITagValue(comments []string, value string) bool {
	vs := getOpenAPITagValue(comments)
	for _, v := range vs {
		if v == value {
			return true
		}
	}
	return false
}

func hasOptionalTag(m *types.Member) bool {
	hasOptionalJSONTag := strings.Contains(
		reflect.StructTag(m.Tags).Get("json"), "omitempty")
	return hasOptionalJSONTag
}

func getJSONTags(m *types.Member) []string {
	jsonTag := reflect.StructTag(m.Tags).Get("json")
	if jsonTag == "" {
		return nil
	}
	return strings.Split(jsonTag, ",")
}

func isInline(m *types.Member) bool {
	jsonTags := getJSONTags(m)
	return len(jsonTags) > 1 && jsonTags[1] == "inline"
}
