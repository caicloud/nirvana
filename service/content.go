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
	"encoding/json"
	"encoding/xml"
	"io"
	"reflect"
	"strconv"

	"github.com/caicloud/nirvana/definition"
)

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

var consumers = map[string]Consumer{
	definition.MIMENone:        &NoneSerializer{},
	definition.MIMEText:        &NoneSerializer{},
	definition.MIMEJSON:        &JSONSerializer{},
	definition.MIMEXML:         &XMLSerializer{},
	definition.MIMEOctetStream: &NoneSerializer{},
	definition.MIMEURLEncoded:  &NoneSerializer{},
	definition.MIMEFormData:    &NoneSerializer{},
}

var producers = map[string]Producer{
	definition.MIMENone:        &NoneSerializer{},
	definition.MIMEText:        &NoneSerializer{},
	definition.MIMEJSON:        &JSONSerializer{},
	definition.MIMEXML:         &XMLSerializer{},
	definition.MIMEOctetStream: &NoneSerializer{},
	definition.MIMEURLEncoded:  &NoneSerializer{},
	definition.MIMEFormData:    &NoneSerializer{},
}

// AllConsumers returns all consumers.
func AllConsumers() []Consumer {
	cs := make([]Consumer, 0, len(consumers))
	for _, c := range consumers {
		cs = append(cs, c)
	}
	return cs
}

// ConsumerFor gets a consumer for specified content type.
func ConsumerFor(contentType string) Consumer {
	return consumers[contentType]
}

// AllProducers returns all producers.
func AllProducers() []Producer {
	ps := make([]Producer, 0, len(producers))
	// JSON always the first one in producers.
	// The first one will be choosed when accept types
	// are not recognized.
	if p := producers[definition.MIMEJSON]; p != nil {
		ps = append(ps, p)
	}
	for _, p := range producers {
		if p.ContentType() == definition.MIMEJSON {
			continue
		}
		ps = append(ps, p)
	}
	return ps
}

// ProducerFor gets a producer for specified content type.
func ProducerFor(contentType string) Producer {
	return producers[contentType]
}

// RegisterConsumer register a consumer. A consumer must not handle "*/*".
func RegisterConsumer(c Consumer) error {
	if c.ContentType() == definition.MIMEAll {
		return invalidConsumer.Error(definition.MIMEAll)
	}
	consumers[c.ContentType()] = c
	return nil
}

// RegisterProducer register a producer. A producer must not handle "*/*".
func RegisterProducer(p Producer) error {
	if p.ContentType() == definition.MIMEAll {
		return invalidProducer.Error(definition.MIMEAll)
	}
	producers[p.ContentType()] = p
	return nil
}

// NoneSerializer implements Consumer and Producer for content types
// which can only receive data by io.Reader.
type NoneSerializer struct{}

// ContentType returns none MIME type.
func (s *NoneSerializer) ContentType() string {
	return definition.MIMENone
}

// Consume does nothing.
func (s *NoneSerializer) Consume(r io.Reader, v interface{}) error {
	return nil
}

// Produce does nothing.
func (s *NoneSerializer) Produce(w io.Writer, v interface{}) error {
	return nil
}

// JSONSerializer implements Consumer and Producer for content type "application/json".
type JSONSerializer struct{}

// ContentType returns json MIME type.
func (s *JSONSerializer) ContentType() string {
	return definition.MIMEJSON
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
	return definition.MIMEXML
}

// Consume unmarshals xml from r into v.
func (s *XMLSerializer) Consume(r io.Reader, v interface{}) error {
	return xml.NewDecoder(r).Decode(v)
}

// Produce marshals v to xml and write to w.
func (s *XMLSerializer) Produce(w io.Writer, v interface{}) error {
	return xml.NewEncoder(w).Encode(v)
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

var prefabs = map[string]Prefab{
	"context": &ContextPrefab{},
}

// PrefabFor gets a prefab by name.
func PrefabFor(name string) Prefab {
	return prefabs[name]
}

// RegisterPrefab registers a prefab.
func RegisterPrefab(prefab Prefab) error {
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

// RegisterConverter registers a converter for specified type. New converter
// overrides old one.
func RegisterConverter(typ reflect.Type, converter Converter) {
	converters[typ] = converter
}

// ConvertToBool converts []string to bool. Only the first data is used.
func ConvertToBool(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseBool(origin)
	if err != nil {
		return nil, invalidConversion.Error(origin, "bool")
	}
	return target, nil
}

// ConvertToInt converts []string to int. Only the first data is used.
func ConvertToInt(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 0)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int")
	}
	return int(target), nil
}

// ConvertToInt8 converts []string to int8. Only the first data is used.
func ConvertToInt8(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 8)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int8")
	}
	return int8(target), nil
}

// ConvertToInt16 converts []string to int16. Only the first data is used.
func ConvertToInt16(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 16)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int16")
	}
	return int16(target), nil
}

// ConvertToInt32 converts []string to int32. Only the first data is used.
func ConvertToInt32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 32)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int32")
	}
	return int32(target), nil
}

// ConvertToInt64 converts []string to int64. Only the first data is used.
func ConvertToInt64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseInt(origin, 10, 64)
	if err != nil {
		return nil, invalidConversion.Error(origin, "int64")
	}
	return target, nil
}

// ConvertToUint converts []string to uint. Only the first data is used.
func ConvertToUint(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 0)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint")
	}
	return uint(target), nil
}

// ConvertToUint8 converts []string to uint8. Only the first data is used.
func ConvertToUint8(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 8)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint8")
	}
	return uint8(target), nil
}

// ConvertToUint16 converts []string to uint16. Only the first data is used.
func ConvertToUint16(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 16)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint16")
	}
	return uint16(target), nil
}

// ConvertToUint32 converts []string to uint32. Only the first data is used.
func ConvertToUint32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 32)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint32")
	}
	return uint32(target), nil
}

// ConvertToUint64 converts []string to uint64. Only the first data is used.
func ConvertToUint64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseUint(origin, 10, 64)
	if err != nil {
		return nil, invalidConversion.Error(origin, "uint64")
	}
	return target, nil
}

// ConvertToFloat32 converts []string to float32. Only the first data is used.
func ConvertToFloat32(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseFloat(origin, 32)
	if err != nil {
		return nil, invalidConversion.Error(origin, "float32")
	}
	return float32(target), nil
}

// ConvertToFloat64 converts []string to float64. Only the first data is used.
func ConvertToFloat64(ctx context.Context, data []string) (interface{}, error) {
	origin := data[0]
	target, err := strconv.ParseFloat(origin, 64)
	if err != nil {
		return nil, invalidConversion.Error(origin, "float64")
	}
	return target, nil
}

// ConvertToString return the first element in []string.
func ConvertToString(ctx context.Context, data []string) (interface{}, error) {
	return data[0], nil
}
