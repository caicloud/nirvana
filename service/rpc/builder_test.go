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

package rpc

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
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

type key string

const wordKey key = "word"

var desc1 = definition.RPCDescriptor{
	Path: "/",
	Middlewares: []definition.Middleware{
		func(ctx context.Context, chain definition.Chain) error {
			return chain.Continue(context.WithValue(ctx, wordKey, "hi"))
		},
	},
	Consumes: []string{definition.MIMEJSON},
	Produces: []string{definition.MIMEJSON},
	Actions: []definition.RPCAction{
		{
			Name:    "GetEcho",
			Version: "2020-01-01",
			Function: func(ctx context.Context, name string) (string, error) {
				return fmt.Sprintf("%s, %s", ctx.Value(wordKey), name), nil
			},
			Consumes: []string{definition.MIMEJSON},
			Produces: []string{definition.MIMEText},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "name",
					Description: "Corresponding for text",
				},
			},
			Results: definition.DataErrorResults(""),
		},
		{
			Name:    "GetEcho2",
			Version: "2020-01-01",
			Function: func(ctx context.Context, body struct {
				Name string `json:"name"`
			}) (struct {
				Name string `json:"name"`
				Word string `json:"word"`
			}, error) {
				return struct {
					Name string `json:"name"`
					Word string `json:"word"`
				}{
					Name: body.Name,
					Word: ctx.Value(wordKey).(string),
				}, nil
			},
			Parameters: []definition.Parameter{
				definition.BodyParameterFor("Corresponding for json"),
			},
			Results: definition.DataErrorResults(""),
		},
	},
}

var desc2 = definition.RPCDescriptor{
	Path: "/aaa",
	Middlewares: []definition.Middleware{
		func(ctx context.Context, chain definition.Chain) error {
			return chain.Continue(context.WithValue(ctx, wordKey, "hello"))
		},
	},
	Actions: []definition.RPCAction{
		{
			Name:    "GetEcho",
			Version: "2020-01-01",
			Function: func(ctx context.Context, name string) (string, error) {
				return fmt.Sprintf("%s, %s", ctx.Value(wordKey), name), nil
			},
			Consumes: []string{definition.MIMEJSON},
			Produces: []string{definition.MIMEText},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "name",
					Description: "Corresponding for text",
				},
			},
			Results: definition.DataErrorResults(""),
		},
	},
}

func TestServer(t *testing.T) {
	cases := []struct {
		u           string
		body        string
		method      string
		expCode     int
		expHeader   string
		expRespBody string
	}{
		{
			u:           "/?Action=GetEcho&Version=2020-01-01&name=alice",
			body:        "",
			method:      "POST",
			expCode:     http.StatusOK,
			expHeader:   definition.MIMEText,
			expRespBody: "hi, alice",
		},
		{
			u:           "/?Action=GetEcho&Version=2020-01-01&name=badguy",
			body:        "",
			method:      "POST",
			expCode:     http.StatusForbidden,
			expHeader:   definition.MIMEText,
			expRespBody: "",
		},
		{
			u:           "/?Action=NotExist&Version=2020-01-01",
			body:        "",
			method:      "POST",
			expCode:     http.StatusMethodNotAllowed,
			expHeader:   definition.MIMEText,
			expRespBody: "",
		},
		{
			u:           "/?Action=GetEcho2&Version=2020-01-01",
			body:        `{"name":"bob"}`,
			method:      "POST",
			expCode:     http.StatusOK,
			expHeader:   definition.MIMEJSON,
			expRespBody: `{"name":"bob","word":"hi"}`,
		},
		{
			u:           "/aaa?Action=GetEcho&Version=2020-01-01&name=peter",
			body:        "",
			method:      "POST",
			expCode:     http.StatusOK,
			expHeader:   definition.MIMEText,
			expRespBody: `hello, peter`,
		},
	}

	builder := NewBuilder()
	builder.SetModifier(service.FirstContextParameter())
	builder.AddFilter(func(resp http.ResponseWriter, req *http.Request) bool {
		if req.URL.Query().Get("name") == "badguy" {
			resp.WriteHeader(http.StatusForbidden)
			return false
		}
		return true
	})
	err := builder.AddDescriptor(desc1, desc2)
	if err != nil {
		t.Fatal(err)
	}
	s, err := builder.Build()
	if err != nil {
		t.Fatal(err)
	}

	for i, c := range cases {
		u, _ := url.Parse(c.u)

		req := &http.Request{
			Method: c.method,
			URL:    u,
			Header: http.Header{
				"Content-Type": []string{definition.MIMEJSON},
				"Accept":       []string{c.expHeader},
				"User-Agent":   []string{"nothing"},
			},
		}
		req = req.WithContext(context.Background())
		req.Body = ioutil.NopCloser(bytes.NewBuffer([]byte(c.body)))
		resp := newRW()
		s.ServeHTTP(resp, req)
		t.Logf("%s", resp.buf.Bytes())
		if resp.code != c.expCode {
			t.Fatalf("case %d: Response code should be %d, but got: %d", i, c.expCode, resp.code)
		}
		if c.expCode == http.StatusOK {
			if resp.header == nil || resp.header.Get("Content-Type") != c.expHeader {
				t.Fatalf("case %d: Content-Type should be %s, but got: %s", i, c.expHeader, resp.header.Get("Content-Type"))
			}
			result := resp.buf.String()
			if strings.TrimSuffix(result, "\n") != c.expRespBody {
				t.Fatalf("case %d: Response does not match: %s", i, result)
			}
		}
	}
}

func TestDefinitions(t *testing.T) {
	expDefs := map[string][]definition.Definition{
		"/?Version=2020-01-01&Action=GetEcho": {
			{
				Method:        "Create",
				Consumes:      desc1.Actions[0].Consumes,
				Produces:      desc1.Actions[0].Produces,
				Tags:          nil,
				ErrorProduces: desc1.Actions[0].Produces,
				Function:      nil,
				Parameters:    desc1.Actions[0].Parameters,
				Results:       desc1.Actions[0].Results,
				Summary:       "summary",
				Description:   "",
				Example:       nil,
			},
		},
		"/?Version=2020-01-01&Action=GetEcho2": {
			{
				Method:        "Create",
				Consumes:      desc1.Consumes,
				Produces:      desc1.Produces,
				Tags:          nil,
				ErrorProduces: desc1.Produces,
				Function:      nil,
				Parameters:    desc1.Actions[1].Parameters,
				Results:       desc1.Actions[1].Results,
				Summary:       "summary",
				Description:   "",
				Example:       nil,
			},
		},
		"/aaa?Version=2020-01-01&Action=GetEcho": {
			{
				Method:        "Create",
				Consumes:      desc2.Actions[0].Consumes,
				Produces:      desc2.Actions[0].Produces,
				Tags:          nil,
				ErrorProduces: desc2.Actions[0].Produces,
				Function:      nil,
				Parameters:    desc2.Actions[0].Parameters,
				Results:       desc2.Actions[0].Results,
				Summary:       "summary",
				Description:   "",
				Example:       nil,
			},
		},
	}
	builder := NewBuilder()
	builder.SetModifier(func(d *definition.Definition) {
		d.Summary = "summary"
	})
	err := builder.AddDescriptor(desc1, desc2)
	if err != nil {
		t.Fatal(err)
	}
	defs := builder.Definitions()
	for _, def := range defs {
		def[0].Function = nil
	}
	if !reflect.DeepEqual(defs, expDefs) {
		t.Fatalf("unexpected definition")
	}
}

func TestDuplicatedRPCPath(t *testing.T) {
	const version = "2020-10-10"
	action := definition.RPCAction{
		Name:    "GetFoo",
		Version: version,
		Function: func() (string, error) {
			return "", nil
		},
		Results: definition.DataErrorResults(""),
	}
	desc := definition.RPCDescriptor{
		Path:        "/",
		Description: "Test",
		Consumes:    []string{definition.MIMEAll},
		Produces:    []string{definition.MIMEAll},
		Actions:     []definition.RPCAction{action, action},
	}
	builder := NewBuilder()
	err := builder.AddDescriptor(desc)
	if err == nil {
		t.Fatal("Unexpected success")
	}
	if !strings.Contains(err.Error(), "duplicated rpc path") {
		t.Fatalf("Unexpected error: %v", err)
	}
}

func BenchmarkServer(b *testing.B) {
	u, _ := url.Parse("/?Action=GetEcho&Version=2020-01-01&name=alice")

	req := &http.Request{
		Method: "POST",
		URL:    u,
		Header: http.Header{
			"Content-Type": []string{definition.MIMEJSON},
			"Accept":       []string{fmt.Sprintf("%s,%s", definition.MIMEText, definition.MIMEText)},
			"User-Agent":   []string{"nothing"},
		},
	}
	builder := NewBuilder()
	builder.SetModifier(service.FirstContextParameter())
	builder.AddFilter(service.RedirectTrailingSlash(), service.FillLeadingSlash(), service.ParseRequestForm())
	err := builder.AddDescriptor(desc1, desc2)
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
		s.ServeHTTP(resp, req)
		resp.buf = bytes.NewBuffer(resp.buf.Bytes())
	}
}
