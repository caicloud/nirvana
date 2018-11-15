package unittest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
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

// NewTestService creates a service.Service for testing
func NewTestService(desc ...definition.Descriptor) (service.Service, error) {
	builder := service.NewBuilder()
	builder.SetModifier(service.FirstContextParameter())
	builder.AddFilter(service.RedirectTrailingSlash(), service.FillLeadingSlash(), service.ParseRequestForm())
	if err := builder.AddDescriptor(desc...); err != nil {
		return nil, err
	}
	return builder.Build()
}

// NewJSONRequest creates a http.Request with json Content-Type. The data parameter can be io.Reader,
// []bytes or a struct.
func NewJSONRequest(ctx context.Context, method, url string, data interface{}) (*http.Request, error) {
	var r io.ReadCloser
	if data != nil {
		switch t := data.(type) {
		case io.ReadCloser:
			r = t
		case io.Reader:
			r = ioutil.NopCloser(t)
		case []byte:
			r = ioutil.NopCloser(bytes.NewBuffer(t))
		default:
			jsonBytes, err := json.Marshal(data)
			if err != nil {
				return nil, fmt.Errorf("error encode data: %v", err)
			}
			ioutil.NopCloser(bytes.NewBuffer(jsonBytes))
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
