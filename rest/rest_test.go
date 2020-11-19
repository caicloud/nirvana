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

package rest

import (
	"reflect"
	"testing"
)

func TestClient_parseURL(t *testing.T) {
	cases := []struct {
		rawurl    string
		parsedURL parsedURL
	}{
		{
			rawurl: "/apis/v1/messages/{message}",
			parsedURL: parsedURL{
				path: &path{
					path: "/apis/v1/messages/{message}",
					names: map[string]int{
						"message": 1,
					},
					segments: []string{
						"/apis/v1/messages/",
						"message",
					},
				},
				queries: map[string][]string{},
			},
		},
		{
			rawurl: "/?Action=GetMessage&Version=2020-01-01",
			parsedURL: parsedURL{
				path: &path{
					path:     "/",
					names:    map[string]int{},
					segments: []string{"/"},
				},
				queries: map[string][]string{
					"Action":  {"GetMessage"},
					"Version": {"2020-01-01"},
				},
			},
		},
		{
			rawurl: "/apis/v1/messages/{message}?Action=GetMessage&Version=2020-01-01",
			parsedURL: parsedURL{
				path: &path{
					path: "/apis/v1/messages/{message}",
					names: map[string]int{
						"message": 1,
					},
					segments: []string{
						"/apis/v1/messages/",
						"message",
					},
				},
				queries: map[string][]string{
					"Action":  {"GetMessage"},
					"Version": {"2020-01-01"},
				},
			},
		},
		// test cache return
		{
			rawurl: "/apis/v1/messages/{message}?Action=GetMessage&Version=2020-01-01",
			parsedURL: parsedURL{
				path: &path{
					path: "/apis/v1/messages/{message}",
					names: map[string]int{
						"message": 1,
					},
					segments: []string{
						"/apis/v1/messages/",
						"message",
					},
				},
				queries: map[string][]string{
					"Action":  {"GetMessage"},
					"Version": {"2020-01-01"},
				},
			},
		},
	}

	cli, err := NewClient(&Config{
		Scheme:   "",
		Host:     "test",
		Executor: nil,
	})
	if err != nil {
		t.Fatalf("new client: %v", err)
	}
	for _, cs := range cases {
		parsedURL, err := cli.parseURL(cs.rawurl)
		if err != nil {
			t.Errorf("parse url: %v", err)
			continue
		}
		if !reflect.DeepEqual(parsedURL, cs.parsedURL) {
			t.Errorf("unexpected result, parsedURL is %v", parsedURL)
			continue
		}
	}
}
