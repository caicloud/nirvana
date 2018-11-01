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

package service

import (
	"reflect"
	"testing"
)

func TestParseAcceptTypes(t *testing.T) {
	cts := []struct {
		ct      string
		results []string
	}{
		{"text/*",
			[]string{"text/*"}},
		{"text/*;q=0.3",
			[]string{"text/*"}},
		{"text/*;q=0.3, text/html;q=0.7",
			[]string{"text/html", "text/*"}},
		{"text/*;q =0.3, text/html;q=0.7, text/html;level =1, text/html ; level=2;q=0.4, */*;q=0.5",
			[]string{"text/html;level=1", "text/html", "*/*", "text/html;level=2", "text/*"}},
		{"text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,image/apng,*/*;q=0.8",
			[]string{"text/html", "application/xhtml+xml", "image/webp", "image/apng", "application/xml", "*/*"}},
		{"application/openmetrics-text; version=0.0.1,*/*;q=0.1,text/plain;version=0.0.4;q=0.5",
			[]string{"application/openmetrics-text;version=0.0.1", "text/plain;version=0.0.4", "*/*"}},
	}
	for _, ct := range cts {
		results, err := parseAcceptTypes(ct.ct)
		if err != nil {
			t.Fatalf("Generate with error: %s", err.Error())
		}
		if !reflect.DeepEqual(ct.results, results) {
			t.Fatalf("Generate wrong results for %s: %v", ct.ct, results)
		}
	}
}
