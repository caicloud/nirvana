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
	"context"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/errors"
)

// WrapHTTPHandler wraps an http handler to definition function.
func WrapHTTPHandler(h http.Handler) func(ctx context.Context) {
	return func(ctx context.Context) {
		httpCtx := HTTPContextFrom(ctx)
		h.ServeHTTP(httpCtx.ResponseWriter(), httpCtx.Request())
	}
}

// WrapHTTPHandlerFunc wraps an http handler func to definition function.
func WrapHTTPHandlerFunc(f http.HandlerFunc) func(ctx context.Context) {
	return func(ctx context.Context) {
		httpCtx := HTTPContextFrom(ctx)
		f(httpCtx.ResponseWriter(), httpCtx.Request())
	}
}

// FileNotFound is an error factory to show why can't find a file.
// This error may contains private information. Don't return this error to end users directly.
var FileNotFound = errors.NotFound.Build("Nirvana:Service:FileNotFound", "can't find file ${path} because ${reason}")

// FileForbidden is an error factory to show why can't access a file.
// This error may contains private information. Don't return this error to end users directly.
var FileForbidden = errors.Forbidden.Build("Nirvana:Service:FileForbidden", "can't access file ${path} because ${reason}")

// UnreadableFile is an error factory to show why can't read a file.
// This error may contains private information. Don't return this error to end users directly.
var UnreadableFile = errors.InternalServerError.Build("Nirvana:Service:UnreadableFile", "can't read file ${path} because ${reason}")

// UnseekableFile is an error factory to show why can't seek a file.
// This error may contains private information. Don't return this error to end users directly.
var UnseekableFile = errors.InternalServerError.Build("Nirvana:Service:UnseekableFile", "can't seek file ${path} because ${reason}")

// ReadFile reads file and returns mime type.
func ReadFile(path string) (string, io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		switch {
		case os.IsNotExist(err):
			return "", nil, FileNotFound.Error(path, err)
		case os.IsPermission(err):
			return "", nil, FileForbidden.Error(path, err)
		}
		return "", nil, UnreadableFile.Error(path, err)
	}
	ctype := mime.TypeByExtension(filepath.Ext(path))
	if ctype == "" {
		// Read a chunk to decide between utf-8 text and binary
		var buf [512]byte
		n, _ := io.ReadFull(file, buf[:])
		ctype = http.DetectContentType(buf[:n])
		_, err := file.Seek(0, io.SeekStart)
		if err != nil {
			return "", nil, UnseekableFile.Error(path, err)
		}
	}
	if ctype == "" {
		ctype = definition.MIMEOctetStream
	}
	return ctype, file, nil
}

// MetaForContentType returns a meta for content type.
func MetaForContentType(ctype string) map[string]string {
	return map[string]string{"Content-Type": ctype}
}

// Internal error factories:

var (
	// InvalidService represents no response error.
	InvalidService = errors.InternalServerError.Build("Nirvana:Service:NoResponse", "no response")
	// NoContext means can't find http context.
	NoContext = errors.InternalServerError.Build("Nirvana:Service:NoContext", "can't find http context, you should define `ctx context.Context` as the first parameter of your handler function")
	// UnassignableType represents unassignable type error.
	UnassignableType = errors.InternalServerError.Build("Nirvana:Service:UnassignableType", "type ${typeA} can't assign to ${typeB}")
	// NoConverter represents no converter for type error.
	NoConverter = errors.InternalServerError.Build("Nirvana:Service:UnassignableType", "no converter for type ${type}")
	// NoParameterGenerator represents no parameter generator error.
	NoParameterGenerator = errors.InternalServerError.Build("Nirvana:Service:NoParameterGenerator", "no parameter generator for source ${source}")
)

var (
	invalidContentType     = errors.BadRequest.Build("Nirvana:Service:InvalidContentType", "invalid content type ${type}")
	invalidConversion      = errors.BadRequest.Build("Nirvana:Service:InvalidConversion", "can't convert ${data} to ${type}")
	invalidConsumer        = errors.InternalServerError.Build("Nirvana:Service:invalidConsumer", "${type} is invalid for consumer")
	invalidProducer        = errors.InternalServerError.Build("Nirvana:Service:invalidProducer", "${type} is invalid for producer")
	noConnectionHijacker   = errors.InternalServerError.Build("Nirvana:Service:noConnectionHijacker", "underlying http.ResponseWriter does not implement http.Hijacker")
	invalidMetaType        = errors.InternalServerError.Build("Nirvana:Service:invalidMetaType", "can't recognize meta for type ${type}")
	noProducerToWrite      = errors.NotAcceptable.Build("Nirvana:Service:noProducerToWrite", "can't find producer for accept types ${types}")
	invalidMethod          = errors.InternalServerError.Build("Nirvana:Service:invalidMethod", "http method ${method} is invalid")
	invalidStatusCode      = errors.InternalServerError.Build("Nirvana:Service:invalidStatusCode", "http status code must be in [100,599]")
	invalidBodyType        = errors.InternalServerError.Build("Nirvana:Service:invalidBodyType", "${type} is not a valid type for body")
	noPrefab               = errors.InternalServerError.Build("Nirvana:Service:noPrefab", "no prefab named ${name}")
	invalidAutoParameter   = errors.InternalServerError.Build("Nirvana:Service:invalidAutoParameter", "${type} is not a struct or a pointer to struct")
	invalidFieldTag        = errors.InternalServerError.Build("Nirvana:Service:invalidFieldTag", "filed tag ${tag} is invalid")
	noName                 = errors.InternalServerError.Build("Nirvana:Service:noName", "${source} must have a name")
	invalidTypeForConsumer = errors.InternalServerError.Build("Nirvana:Service:invalidTypeForConsumer", "consumer ${content} can't consume data for type ${type}")
	invalidTypeForProducer = errors.InternalServerError.Build("Nirvana:Service:invalidTypeForProducer", "producer ${content} can't produce data for type ${type}")
)
