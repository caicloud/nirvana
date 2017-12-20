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

package web

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
)

var (
	// contextKeyUnderlyingHTTPContext is a key for context.
	// It's unique and point to httpCtx.
	contextKeyUnderlyingHTTPContext interface{} = new(byte)
)

// httpCtx contains a http.Request and a http.ResponseWriter for a request.
// It goes through the life cycle of a request.
type httpCtx struct {
	context.Context
	container container
	response  response
}

func newHTTPContext() interface{} {
	ctx := &httpCtx{}
	ctx.container.params = make([]param, 0, 5)
	return ctx
}

// reset resets itself by a http.Request and a http.ResponseWriter.
func (c *httpCtx) reset(w http.ResponseWriter, request *http.Request) {
	// Get context from request.
	c.Context = request.Context()
	// Reset value container.
	c.container.request = request
	c.container.params = c.container.params[:0]
	c.container.query = nil
	// Reset response.
	c.response.writer = w
	c.response.statusCode = 0
	c.response.contentLength = 0
}

// clear clears request and response.
func (c *httpCtx) clear() {
	c.Context = nil
	c.container.request = nil
	c.container.query = nil
	c.response.writer = nil
}

// Value returns itself when key is contextKeyUnderlyingHTTPContext.
func (c *httpCtx) Value(key interface{}) interface{} {
	if key == contextKeyUnderlyingHTTPContext {
		return c
	}
	return c.Context.Value(key)
}

// ValueContainer contains values from a request.
type ValueContainer interface {
	// Path returns path value by key.
	Path(key string) (string, bool)
	// Query returns value from query string.
	Query(key string) ([]string, bool)
	// Header returns value by header key.
	Header(key string) ([]string, bool)
	// Form returns value from request. It is valid when
	// http "Content-Type" is "application/x-www-form-urlencoded"
	// or "multipart/form-data".
	Form(key string) ([]string, bool)
	// File returns a file reader when "Content-Type" is "multipart/form-data".
	File(key string) (multipart.File, bool)
	// Body returns a reader to read data from request body.
	// The reader only can read once.
	Body() (reader io.ReadCloser, contentType string, ok bool)
}

type param struct {
	key   string
	value string
}

// container implements ValueContainer and provides methods to get values.
type container struct {
	request *http.Request
	params  []param
	query   url.Values
}

// Set sets path parameter key-value pairs.
func (c *container) Set(key, value string) {
	c.params = append(c.params, param{key, value})
}

// Get gets path value.
func (c *container) Get(key string) (string, bool) {
	for i := len(c.params) - 1; i >= 0; i-- {
		p := c.params[i]
		if p.key == key {
			return p.value, true
		}
	}
	return "", false
}

// Path returns path value by key. It's same as Get().
func (c *container) Path(key string) (string, bool) {
	return c.Get(key)
}

// Query returns value from query string.
func (c *container) Query(key string) ([]string, bool) {
	if c.query == nil {
		c.query = c.request.URL.Query()
	}
	values := c.query[key]
	return values, len(values) > 0
}

// Header returns value by header key.
func (c *container) Header(key string) ([]string, bool) {
	h := c.request.Header[key]
	return h, len(h) > 0
}

// Form returns value from request. It is valid when
// http "Content-Type" is "application/x-www-form-urlencoded"
// or "multipart/form-data".
func (c *container) Form(key string) ([]string, bool) {
	values := c.request.PostForm[key]
	return values, len(values) > 0
}

// File returns a file reader when "Content-Type" is "multipart/form-data".
func (c *container) File(key string) (multipart.File, bool) {
	file, _, err := c.request.FormFile(key)
	return file, err == nil
}

// Body returns a reader to read data from request body.
// The reader only can read once.
func (c *container) Body() (reader io.ReadCloser, contentType string, ok bool) {
	contentType, err := ContentType(c.request)
	return c.request.Body, contentType, err == nil
}

// ResponseWriter extends http.ResponseWriter.
type ResponseWriter interface {
	http.ResponseWriter
	// HeaderWritable can check whether WriteHeader() has
	// been called. If the method returns false, you should
	// not recall WriteHeader().
	HeaderWritable() bool
	// StatusCode returns status code.
	StatusCode() int
	// ContentLength returns the length of written content.
	ContentLength() int
}

type response struct {
	writer        http.ResponseWriter
	statusCode    int
	contentLength int
}

// For http.HTTPResponseWriter and HTTPResponseInfo
func (c *response) Header() http.Header {
	return c.writer.Header()
}

// Write is a disguise of http.response.Write().
func (c *response) Write(data []byte) (int, error) {
	if c.statusCode <= 0 {
		c.WriteHeader(200)
	}
	length, err := c.writer.Write(data)
	c.contentLength += length
	return length, err
}

// WriteHeader is a disguise of http.response.WriteHeader().
func (c *response) WriteHeader(code int) {
	c.statusCode = code
	c.writer.WriteHeader(code)
}

// Flush is a disguise of http.response.Flush().
func (c *response) Flush() {
	c.writer.(http.Flusher).Flush()
}

// CloseNotify is a disguise of http.response.CloseNotify().
func (c *response) CloseNotify() <-chan bool {
	return c.writer.(http.CloseNotifier).CloseNotify()
}

// Hijack is a disguise of http.response.Hijack().
func (c *response) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if h, ok := c.writer.(http.Hijacker); ok {
		return h.Hijack()
	}
	return nil, nil, fmt.Errorf("underlying http.ResponseWriter does not implement http.Hijacker")
}

// StatusCode returns status code.
func (c *response) StatusCode() int {
	if c.statusCode <= 0 {
		return http.StatusOK
	}
	return c.statusCode
}

// ContentLength returns the length of written content.
func (c *response) ContentLength() int {
	return c.contentLength
}

// HeaderWritable can check whether WriteHeader() has
// been called. If the method returns false, you should
// not recall WriteHeader().
func (c *response) HeaderWritable() bool {
	return c.statusCode <= 0
}

// HTTPRequest gets http.Request from context.
func HTTPRequest(ctx context.Context) *http.Request {
	if c := httpContext(ctx); c != nil {
		return c.container.request
	}
	return nil
}

// HTTPResponseWriter gets ResponseWriter from context.
func HTTPResponseWriter(ctx context.Context) ResponseWriter {
	if c := httpContext(ctx); c != nil {
		return &c.response
	}
	return nil
}

// HTTPValueContainer gets ValueContainer from context.
func HTTPValueContainer(ctx context.Context) ValueContainer {
	if c := httpContext(ctx); c != nil {
		return &c.container
	}
	return nil
}

func httpContext(ctx context.Context) *httpCtx {
	value := ctx.Value(contextKeyUnderlyingHTTPContext)
	if value == nil {
		return nil
	}
	if c, ok := value.(*httpCtx); ok {
		return c
	}
	return nil
}
