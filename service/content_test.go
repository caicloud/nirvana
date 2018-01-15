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

package service

import (
	"bytes"
	"context"
	"io"
	"reflect"
	"testing"

	"github.com/caicloud/nirvana/definition"
)

type vc2 struct {
	vc
	contentType string
	data        string
}

func (v *vc2) Body() (reader io.ReadCloser, contentType string, ok bool) {
	return &file{[]byte(v.data), 0}, v.contentType, true
}

func TestConsumer(t *testing.T) {
	const data = `{"value":"test body"}`
	types := []string{
		definition.MIMEText,
		definition.MIMEJSON,
		definition.MIMEXML,
		definition.MIMEOctetStream,
		definition.MIMEURLEncoded,
		definition.MIMEFormData,
	}
	targets := []reflect.Type{
		reflect.TypeOf(""),
		reflect.TypeOf(([]byte)(nil)),
	}
	defaults := []interface{}{
		"",
		[]byte{},
	}
	g := &BodyParameterGenerator{}
	for _, ct := range types {
		for i, target := range targets {
			def := defaults[i]
			if err := g.Validate("test", def, target); err != nil {
				t.Fatal(err)
			}
			result, err := g.Generate(
				context.Background(),
				&vc2{
					contentType: ct,
					data:        data,
				},
				AllConsumers(),
				"test",
				target,
			)
			if err != nil {
				t.Fatal(err)
			}
			switch r := result.(type) {
			case string:
				if r != data {
					t.Fatalf("Generate wrong data: %v", r)
				}
			case []byte:
				if string(r) != data {
					t.Fatalf("Generate wrong data: %v", r)
				}
			}
		}
	}

}

func TestProducer(t *testing.T) {
	const data = `{"value":"test body"}`
	types := []string{
		definition.MIMEText,
		definition.MIMEJSON,
		definition.MIMEXML,
		definition.MIMEOctetStream,
	}
	values := []interface{}{
		data,
		[]byte(data),
	}
	for _, at := range types {
		producer := ProducerFor(at)
		if producer == nil {
			t.Fatalf("Can't find producer for accept type: %s", at)
		}
		for _, v := range values {
			w := bytes.NewBuffer(nil)
			if err := producer.Produce(w, v); err != nil {
				t.Fatal(err)
			}
			if data != w.String() {
				t.Fatalf("Producer %s writed wrong data: %s", at, w.Bytes())
			}
		}
	}
}
