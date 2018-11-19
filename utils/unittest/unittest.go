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

package unittest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

// ResponseWriter mocks http.ResponseWriter and records response for testing
type ResponseWriter interface {
	http.ResponseWriter
	Code() int
	Bytes() []byte
}

// NewResponseWriter creates a new ResponseWriter
func NewResponseWriter() ResponseWriter {
	return newRW()
}

type responseWriter struct {
	code   int
	header http.Header
	buf    *bytes.Buffer
}

func newRW() *responseWriter {
	return &responseWriter{0, http.Header{}, bytes.NewBuffer(nil)}
}

func (r *responseWriter) Write(d []byte) (int, error) {
	return r.buf.Write(d)
}

func (r *responseWriter) WriteHeader(code int) {
	r.code = code
}

func (r *responseWriter) Header() http.Header {
	return r.header
}

func (r *responseWriter) Code() int {
	return r.code
}

func (r *responseWriter) Bytes() []byte {
	return r.buf.Bytes()
}

// NewTestService creates a service.Service for testing.
func NewTestService(desc ...definition.Descriptor) (service.Service, error) {
	builder := service.NewBuilder()
	builder.SetModifier(service.FirstContextParameter())
	builder.AddFilter(service.RedirectTrailingSlash(), service.FillLeadingSlash(), service.ParseRequestForm())
	if err := builder.AddDescriptor(desc...); err != nil {
		return nil, err
	}
	return builder.Build()
}

// NewTestServiceWithConfig creates a service.Service for testing with user specified modifier and
// filters. If modifier or filters is nil, default option will be used.
func NewTestServiceWithConfig(desc []definition.Descriptor, modifier service.DefinitionModifier,
	filters []service.Filter) (service.Service, error) {
	builder := service.NewBuilder()
	if modifier == nil {
		modifier = service.FirstContextParameter()
	}
	builder.SetModifier(modifier)
	if filters == nil {
		filters = []service.Filter{service.RedirectTrailingSlash(), service.FillLeadingSlash(), service.ParseRequestForm()}
	}
	builder.AddFilter(filters...)
	if err := builder.AddDescriptor(desc...); err != nil {
		return nil, err
	}
	return builder.Build()
}

// NewJSONRequest creates a http.Request with json Content-Type. The data parameter can be io.Reader, []byte or a struct.
func NewJSONRequest(ctx context.Context, method, url string, data interface{}) (*http.Request, error) {
	var r io.Reader
	if data != nil {
		switch t := data.(type) {
		case io.Reader:
			r = t
		case []byte:
			r = bytes.NewBuffer(t)
		default:
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("error encode data: %v", err)
			}
			r = bytes.NewBuffer(jsonBytes)
		}
	}
	req, err := http.NewRequest(method, url, r)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	req.Header.Add("Content-Type", "application/json")
	return req, nil
}
