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

package service

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"

	"github.com/caicloud/nirvana/definition"
)

type responseWriter struct {
	code   int
	header http.Header
	buf    *bytes.Buffer
}

func newRW() *responseWriter {
	return &responseWriter{0, http.Header{}, bytes.NewBuffer(nil)}
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) Write(d []byte) (int, error) {
	return r.buf.Write(d)
}

func (r *responseWriter) WriteHeader(code int) {
	r.code = code
}

var desc = definition.Descriptor{
	Path:        "/api/v1/",
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

type Application struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Target    string `json:"target"`
	Target1   int    `json:"target2"`
	Target2   bool   `json:"target1"`
}

func Handle(ctx context.Context, userAgent string, target1 int, target2 bool, app *Application) (*Application, error) {
	app.Target = userAgent
	app.Target1 = target1
	app.Target2 = target2
	return app, nil
}

func TestServer(t *testing.T) {
	u, _ := url.Parse("/api/v1/1222/false?target1=1&target2=false")
	data := []byte(`{
	"name": "asdasd",
	"namespace": "system"
}`)

	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"Accept":       []string{"application/json"},
			"User-Agent":   []string{"nothing"},
		},
		ContentLength: int64(len(data)),
	}
	builder := NewBuilder()
	builder.SetModifier(FirstContextParameter())
	builder.AddFilter(RedirectTrailingSlash(), FillLeadingSlash(), ParseRequestForm())
	err := builder.AddDescriptor(desc)
	if err != nil {
		t.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}
	req = req.WithContext(context.Background())
	req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
	resp := newRW()
	s.ServeHTTP(resp, req)
	t.Log(resp.code)
	t.Log(resp.header)
	t.Logf("%s", resp.buf.Bytes())
	if resp.code != 201 {
		t.Fatalf("Response code should be 201, but got: %d", resp.code)
	}
	if resp.header == nil || resp.header.Get("Content-Type") != "application/json" {
		t.Fatalf("Content-Type should be application/json, but got: %s", resp.header.Get("Content-Type"))
	}
	result := resp.buf.String()
	target := `{"name":"asdasd","namespace":"system","target":"nothing","target2":1,"target1":false}` + "\n"
	if result != target {
		t.Fatalf("Response does not match: %s", result)
	}
}

func BenchmarkServer(b *testing.B) {
	u, _ := url.Parse("/api/v1/1222/false?target1=1&target2=false")
	data := []byte(`{
	"name": "asdasd",
	"namespace": "system"
}`)

	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
			"Accept":       []string{"application/json"},
			"User-Agent":   []string{"nothing"},
		},
		ContentLength: int64(len(data)),
	}
	builder := NewBuilder()
	builder.SetModifier(FirstContextParameter())
	builder.AddFilter(RedirectTrailingSlash(), FillLeadingSlash(), ParseRequestForm())
	err := builder.AddDescriptor(desc)
	if err != nil {
		b.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		b.Fatal(err)
	}
	req = req.WithContext(context.Background())
	resp := newRW()
	for n := 0; n < b.N; n++ {
		req.Body = ioutil.NopCloser(bytes.NewBuffer(data))
		s.ServeHTTP(resp, req)
		resp.buf = bytes.NewBuffer(resp.buf.Bytes())
	}
}
