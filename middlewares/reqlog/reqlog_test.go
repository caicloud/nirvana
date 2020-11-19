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

package reqlog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/service/rest"
)

const tenantHeader = "X-Tenant"

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

func mustURL(rawURL string) *url.URL {
	u, _ := url.Parse(rawURL)
	return u
}

func buildService(middleware definition.Middleware) (service.Service, error) {
	builder := rest.NewBuilder()
	builder.SetModifier(service.FirstContextParameter())
	if err := builder.AddDescriptor(
		definition.Descriptor{
			Path:        "/api/v1/foo",
			Consumes:    []string{"*/*"},
			Produces:    []string{"application/json"},
			Middlewares: []definition.Middleware{middleware},
			Definitions: []definition.Definition{
				{
					Method: definition.Create,
					Parameters: []definition.Parameter{
						definition.HeaderParameterFor(tenantHeader, ""),
						definition.QueryParameterFor("query", ""),
						definition.BodyParameterFor(""),
					},
					Function: func(ctx context.Context, tenant, query string, body []byte) (string, error) {
						request := service.HTTPContextFrom(ctx).Request()
						return fmt.Sprintf(
							"%s %s %s %s %s",
							request.Method, request.URL.String(),
							tenant, query, string(body),
						), nil
					},
					Results: definition.DataErrorResults(""),
				},
				{
					Method: definition.Get,
					Parameters: []definition.Parameter{
						definition.HeaderParameterFor(tenantHeader, ""),
						definition.QueryParameterFor("query", ""),
					},
					Function: func(ctx context.Context, tenant, query string) (string, error) {
						request := service.HTTPContextFrom(ctx).Request()
						return fmt.Sprintf(
							"%s %s %s %s",
							request.Method, request.URL.String(), tenant, query,
						), nil
					},
					Results: definition.DataErrorResults(""),
				},
			},
		},
	); err != nil {
		return nil, err
	}
	return builder.Build()
}

type testCase struct {
	name      string
	req       http.Request
	wantRegex string
	respStr   string
	respCode  int
}

func buildTestCases() []testCase {
	return []testCase{
		{
			name: "POST",
			req: http.Request{
				Method: http.MethodPost,
				URL:    mustURL("/api/v1/foo?query=bar"),
				Header: http.Header{
					"Content-Type": []string{"application/json"},
					"Accept":       []string{"application/json"},
					"X-Tenant":     []string{"system-tenant"},
				},
				Body: ioutil.NopCloser(strings.NewReader("{\"key\": \"value\"}")),
			},
			wantRegex: `^INFO\s(.+) | POST 201 61 [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+ /api/v1/foo?query=bar$`,
			respStr:   `POST /api/v1/foo?query=bar system-tenant bar {"key": "value"}`,
			respCode:  201,
		},
		{
			name: "GET",
			req: http.Request{
				Method: http.MethodGet,
				URL:    mustURL("/api/v1/foo?query=bar"),
				Header: http.Header{
					"X-Tenant": []string{"system-tenant"},
				},
			},
			wantRegex: `^INFO\s(.+) | GET 200 43 [-+]?([0-9]*(\.[0-9]*)?[a-z]+)+ /api/v1/foo?query=bar$`,
			respStr:   `GET /api/v1/foo?query=bar system-tenant bar`,
			respCode:  200,
		},
	}
}

func TestDefault(t *testing.T) {
	for _, tc := range buildTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			old := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w
			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, r)
				outC <- buf.String()
			}()

			log.SetDefaultLogger(log.NewStdLogger(log.LevelDebug))
			svc, err := buildService(Default())
			if err != nil {
				t.Fatal(err)
			}
			rw := newRW()
			svc.ServeHTTP(rw, &tc.req)

			w.Close()
			out := <-outC
			os.Stderr = old

			if rw.buf.String() != tc.respStr {
				t.Fatalf("%s != %s", rw.buf.String(), tc.respStr)
			}
			if rw.code != tc.respCode {
				t.Fatal("unexpected response code, the middleware might have altered the request/response")
			}
			if !regexp.MustCompile(tc.wantRegex).MatchString(out) {
				t.Fatal("unexpected log content")
			}
		})
	}
}

func TestCustom(t *testing.T) {
	for _, tc := range buildTestCases() {
		t.Run(tc.name, func(t *testing.T) {
			old := os.Stderr
			r, w, _ := os.Pipe()
			os.Stderr = w
			outC := make(chan string)
			go func() {
				var buf bytes.Buffer
				io.Copy(&buf, r)
				outC <- buf.String()
			}()

			svc, err := buildService(Custom(log.NewStdLogger(0), 0))
			if err != nil {
				t.Fatal(err)
			}
			rw := newRW()
			svc.ServeHTTP(rw, &tc.req)

			w.Close()
			out := <-outC
			os.Stderr = old

			if rw.buf.String() != tc.respStr {
				t.Fatalf("%s != %s", rw.buf.String(), tc.respStr)
			}
			if rw.code != tc.respCode {
				t.Fatal("unexpected response code, the middleware might have altered the request/response")
			}
			if !regexp.MustCompile(tc.wantRegex).MatchString(out) {
				t.Fatal("unexpected log content")
			}
		})
	}
}
