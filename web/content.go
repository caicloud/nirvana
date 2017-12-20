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
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strconv"
)

// acceptTypeAll indicates a accept type from http request.
// It means client can receive any content.
// Request content type in header "Content-Type" must not set to "*/*".
// It only can exist in request header "Accept".
// In most time, it locate at the last element of "Accept".
// It's default value if client have not set "Accept" header.
const acceptTypeAll = "*/*"

// Consumer handles specifically typed data from a reader and unmarshals it into an object.
type Consumer interface {
	// ContentType returns a HTTP MIME type.
	ContentType() string
	// Consume unmarshals data from r into v.
	Consume(r io.Reader, v interface{}) error
}

// Producer marshals an object to specifically typed data and write it into a writer.
type Producer interface {
	// ContentType returns a HTTP MIME type.
	ContentType() string
	// Produce marshals v to data and write to w.
	Produce(w io.Writer, v interface{}) error
}

var consumers = map[string]Consumer{}
var producers = map[string]Producer{}

// ConsumerFor gets a consumer for specified content type.
func ConsumerFor(contentType string) Consumer {
	return consumers[contentType]
}

// ProducerFor gets a producer for specified content type.
func ProducerFor(contentType string) Producer {
	return producers[contentType]
}

// RegisterConsumer register a consumer. A consumer must not handle "*/*".
func RegisterConsumer(c Consumer) error {
	if c.ContentType() == acceptTypeAll {
		return fmt.Errorf("must not register a consumer for %s", acceptTypeAll)
	}
	if _, ok := consumers[c.ContentType()]; ok {
		return fmt.Errorf("consumer %s has been registered", c.ContentType())
	}
	consumers[c.ContentType()] = c
	return nil
}

// RegisterProducer register a producer. A producer must not handle "*/*".
func RegisterProducer(p Producer) error {
	if p.ContentType() == acceptTypeAll {
		return fmt.Errorf("must not register a producer for %s", acceptTypeAll)
	}
	if _, ok := producers[p.ContentType()]; ok {
		return fmt.Errorf("producer %s has been registered", p.ContentType())
	}
	producers[p.ContentType()] = p
	return nil
}

// JSONSerializer implements Consumer and Producer for content type "application/json".
type JSONSerializer struct{}

// ContentType returns json MIME type.
func (s *JSONSerializer) ContentType() string {
	return "application/json"
}

// Consume unmarshals json from r into v.
func (s *JSONSerializer) Consume(r io.Reader, v interface{}) error {
	return json.NewDecoder(r).Decode(v)
}

// Produce marshals v to json and write to w.
func (s *JSONSerializer) Produce(w io.Writer, v interface{}) error {
	return json.NewEncoder(w).Encode(v)
}

// XMLSerializer implements Consumer and Producer for content type "application/xml".
type XMLSerializer struct{}

// ContentType returns xml MIME type.
func (s *XMLSerializer) ContentType() string {
	return "application/xml"
}

// Consume unmarshals xml from r into v.
func (s *XMLSerializer) Consume(r io.Reader, v interface{}) error {
	return xml.NewDecoder(r).Decode(v)
}

// Produce marshals v to xml and write to w.
func (s *XMLSerializer) Produce(w io.Writer, v interface{}) error {
	return xml.NewEncoder(w).Encode(v)
}

const (
	// MimeOctet is the mime type for unknown byte stream
	MimeOctet = "application/octet-stream"
	// MimeText is the mime type for text
	MimeText = "text/plain"
)

type nopProducer string

func (p nopProducer) Produce(w io.Writer, v interface{}) error { return nil }

func (p nopProducer) ContentType() string {
	return string(p)
}

// Prefab creates instances for internal type. These instances are not
// unmarshaled form http request data.
type Prefab interface {
	// Name returns prefab name.
	Name() string
	// Type is instance type.
	Type() reflect.Type
	// Make makes an instance.
	Make(ctx context.Context) (interface{}, error)
}

var prefabs = map[string]Prefab{}

// PrefabFor gets a prefab by name.
func PrefabFor(name string) Prefab {
	return prefabs[name]
}

// RegisterPrefab registers a prefab.
func RegisterPrefab(prefab Prefab) error {
	if _, ok := prefabs[prefab.Name()]; ok {
		return fmt.Errorf("prefab %s has been registered", prefab.Name())
	}
	prefabs[prefab.Name()] = prefab
	return nil
}

// ContextPrefab returns context from parameter of Make().
// It's usually used for generating the first parameter of api handler.
type ContextPrefab struct{}

// Name returns prefab name.
func (p *ContextPrefab) Name() string {
	return "context"
}

// Type is type of context.Context.
func (p *ContextPrefab) Type() reflect.Type {
	return reflect.TypeOf((*context.Context)(nil)).Elem()
}

// Make returns context simply.
func (p *ContextPrefab) Make(ctx context.Context) (interface{}, error) {
	return ctx, nil
}

// Converter is used to convert []string to specific type. Data must have one
// element at least or it will panic.
type Converter func(ctx context.Context, data []string) (interface{}, error)

var converters = map[reflect.Type]Converter{
	reflect.TypeOf(bool(false)): ConvertToBool,
	reflect.TypeOf(int(0)):      ConvertToInt,
	reflect.TypeOf(int8(0)):     ConvertToInt8,
	reflect.TypeOf(int16(0)):    ConvertToInt16,
	reflect.TypeOf(int32(0)):    ConvertToInt32,
	reflect.TypeOf(int64(0)):    ConvertToInt64,
	reflect.TypeOf(uint(0)):     ConvertToUint,
	reflect.TypeOf(uint8(0)):    ConvertToUint8,
	reflect.TypeOf(uint16(0)):   ConvertToUint16,
	reflect.TypeOf(uint32(0)):   ConvertToUint32,
	reflect.TypeOf(uint64(0)):   ConvertToUint64,
	reflect.TypeOf(float32(0)):  ConvertToFloat32,
	reflect.TypeOf(float64(0)):  ConvertToFloat64,
	reflect.TypeOf(string("")):  ConvertToString,
}

// ConverterFor gets converter for specified type.
func ConverterFor(typ reflect.Type) Converter {
	return converters[typ]
}

// SetConverter sets a converter for specified type. New converter
// overrides old one.
func SetConverter(typ reflect.Type, converter Converter) {
	converters[typ] = converter
}

// ConvertToBool converts []string to bool. Only the first data is used.
func ConvertToBool(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseBool(origin)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to bool", origin)
	}
	return target, nil
}

// ConvertToInt converts []string to int. Only the first data is used.
func ConvertToInt(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to int", reflect.TypeOf(origin).String())
	}
	return int(target), nil
}

// ConvertToInt8 converts []string to int8. Only the first data is used.
func ConvertToInt8(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to int8", reflect.TypeOf(origin).String())
	}
	return int8(target), nil
}

// ConvertToInt16 converts []string to int16. Only the first data is used.
func ConvertToInt16(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to int16", reflect.TypeOf(origin).String())
	}
	return int16(target), nil
}

// ConvertToInt32 converts []string to int32. Only the first data is used.
func ConvertToInt32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to int32", reflect.TypeOf(origin).String())
	}
	return int32(target), nil
}

// ConvertToInt64 converts []string to int64. Only the first data is used.
func ConvertToInt64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to int64", reflect.TypeOf(origin).String())
	}
	return target, nil
}

// ConvertToUint converts []string to uint. Only the first data is used.
func ConvertToUint(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 0)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to uint", reflect.TypeOf(origin).String())
	}
	return uint(target), nil
}

// ConvertToUint8 converts []string to uint8. Only the first data is used.
func ConvertToUint8(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 8)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to uint8", reflect.TypeOf(origin).String())
	}
	return uint8(target), nil
}

// ConvertToUint16 converts []string to uint16. Only the first data is used.
func ConvertToUint16(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 16)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to uint16", reflect.TypeOf(origin).String())
	}
	return uint16(target), nil
}

// ConvertToUint32 converts []string to uint32. Only the first data is used.
func ConvertToUint32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to uint32", reflect.TypeOf(origin).String())
	}
	return uint32(target), nil
}

// ConvertToUint64 converts []string to uint64. Only the first data is used.
func ConvertToUint64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to uint64", reflect.TypeOf(origin).String())
	}
	return target, nil
}

// ConvertToFloat32 converts []string to float32. Only the first data is used.
func ConvertToFloat32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseFloat(origin, 32)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to float32", reflect.TypeOf(origin).String())
	}
	return float32(target), nil
}

// ConvertToFloat64 converts []string to float64. Only the first data is used.
func ConvertToFloat64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseFloat(origin, 64)
	if err != nil {
		return nil, fmt.Errorf("can't convert %s to float64", reflect.TypeOf(origin).String())
	}
	return target, nil
}

// ConvertToString return the first element in []string.
func ConvertToString(ctx context.Context, data []string) (interface{}, error) {
	return data[0], nil
}
