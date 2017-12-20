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
	"context"
	"fmt"
	"net/http"
	"reflect"

	"github.com/caicloud/nirvana/definition"
)

// Error is a common interface for error.
// If an error implements the interface, type handlers can
// use Code() to get a specified HTTP status code.
type Error interface {
	// Code is a HTTP status code.
	Code() int
	// Message is an object which contains information of the error.
	Message() interface{}
}

// MetaOperatorForKey is a util to generate http.Header for a string.
// If your handler return a string and you want set it to the header of a response,
// you can use the util to specify the key of header.
//
// ex.
// func SomeHandler(ctx context.Context) (contentType string, err error)
// You can add the operator into the field "definition.Definition.Results[].Operators[]":
// MetaOperatorForKey("Content-Type")
func MetaOperatorForKey(key string) definition.Operator {
	return definition.OperatorFunc(func(ctx context.Context, object interface{}) (interface{}, error) {
		if str, ok := object.(string); ok {
			return map[string]string{key: str}, nil
		}
		return nil, fmt.Errorf("can't convert meta for type %s", reflect.TypeOf(object))
	})
}

const (
	// HighPriority for error type.
	// If an error occurs, ignore meta and data.
	HighPriority int = 100
	// MediumPriority for meta type.
	MediumPriority int = 200
	// LowPriority for data type.
	LowPriority int = 300
)

// TypeHandler is used to handle the results from API handlers.
type TypeHandler interface {
	// Type returns definition.Type which the type handler can handle.
	Type() definition.Type
	// Priority returns priority of the type handler. Type handler with higher priority will prior execute.
	Priority() int
	// Validate validates whether the type handler can handle the target type.
	Validate(target reflect.Type) error
	// Handle handles a value. If the handler has something wrong, it should return an error.
	// The handler descides how to deal with value by producers and status code.
	// The status code is a success status code. If everything is ok, the handler should use the status code.
	//
	// There are three cases for return values (goon means go on or continue):
	// 1. go on is true, err is nil.
	//    It means that current type handler did nothing (or looks like did nothing) and next type handler
	//    should take the context.
	// 2. go on is false, err is nil.
	//    It means that current type handler has finished the context and next type handler should not run.
	// 3. err is not nil
	//    It means that current type handler handled the context but something wrong. All subsequent type
	//    handlers should not run.
	Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error)
}

var handlers = map[definition.Type]TypeHandler{}

// TypeHandlerFor gets a type handler for specified type.
func TypeHandlerFor(typ definition.Type) TypeHandler {
	return handlers[typ]
}

// RegisterTypeHandler registers a type handler.
func RegisterTypeHandler(handler TypeHandler) error {
	if _, ok := handlers[handler.Type()]; ok {
		return fmt.Errorf("type handler of type %s has been registered", handler.Type())
	}
	handlers[handler.Type()] = handler
	return nil
}

// MetaTypeHandler writes metadata to http.ResponseWriter.Header and value type should be map[string]string.
// If value type is not map, the handler will stop the handlers chain and return an error.
// If there is no error, it always expect that the next handler goes on.
type MetaTypeHandler struct{}

func (h *MetaTypeHandler) Type() definition.Type              { return definition.Meta }
func (h *MetaTypeHandler) Priority() int                      { return MediumPriority }
func (h *MetaTypeHandler) Validate(target reflect.Type) error { return nil }
func (h *MetaTypeHandler) Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error) {
	if value == nil {
		return true, nil
	}
	if values, ok := value.(map[string]string); ok {
		headers := HTTPResponseWriter(ctx).Header()
		for key, value := range values {
			headers.Set(key, value)
		}
		return true, nil
	}
	return false, fmt.Errorf("can't recognize meta for type %s", reflect.TypeOf(value))
}

// DataTypeHandler writes value to http.ResponseWriter. The type handler handle object value.
// If value is nil, the handler does nothing.
type DataTypeHandler struct{}

func (h *DataTypeHandler) Type() definition.Type              { return definition.Data }
func (h *DataTypeHandler) Priority() int                      { return LowPriority }
func (h *DataTypeHandler) Validate(target reflect.Type) error { return nil }
func (h *DataTypeHandler) Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error) {
	if value == nil {
		return true, nil
	}
	err = WriteData(ctx, producers, code, value)
	return err == nil, err
}

// ErrorTypeHandler writes error to http.ResponseWriter.
// If there is no error, the handler does nothing.
type ErrorTypeHandler struct{}

func (h *ErrorTypeHandler) Type() definition.Type              { return definition.Error }
func (h *ErrorTypeHandler) Priority() int                      { return HighPriority }
func (h *ErrorTypeHandler) Validate(target reflect.Type) error { return nil }
func (h *ErrorTypeHandler) Handle(ctx context.Context, producers []Producer, code int, value interface{}) (goon bool, err error) {
	if value == nil {
		return true, nil
	}
	if e, ok := value.(Error); ok {
		err := WriteData(ctx, producers, e.Code(), e.Message())
		return false, err
	}
	return false, WriteData(ctx, producers, http.StatusInternalServerError, value)
}

// WriteData chooses right producer by "Accrpt" header and writes data to context.
// You should never call the function except you are writing a type handler.
func WriteData(ctx context.Context, producers []Producer, code int, data interface{}) error {
	httpCtx := httpContext(ctx)
	ats, err := AcceptTypes(httpCtx.container.request)
	if err != nil {
		return err
	}
	producer := chooseProducer(ats, producers)
	if producer == nil {
		return fmt.Errorf("can't find producer for accept types %+v", ats)
	}
	if httpCtx.response.HeaderWritable() {
		httpCtx.response.Header().Set("Content-Type", producer.ContentType())
		httpCtx.response.WriteHeader(code)
	}
	return producer.Produce(&httpCtx.response, data)
}

func chooseProducer(acceptTypes []string, producers []Producer) Producer {
	if len(acceptTypes) <= 0 || len(producers) <= 0 {
		return nil
	}
	for _, v := range acceptTypes {
		if v == acceptTypeAll {
			return producers[0]
		}
		for _, p := range producers {
			if p.ContentType() == v {
				return p
			}
		}
	}
	return nil
}
