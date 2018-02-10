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

package config

import (
	"fmt"
	"reflect"
	"testing"
	"time"

	"github.com/caicloud/nirvana"
)

type AnotherOption struct {
	Int8        int8          `desc:"Int8"`
	Int16       int16         `desc:"Int16"`
	Int32       int32         `desc:"Int32"`
	Int64       int64         `desc:"Int64"`
	Int         int           `desc:"Int"`
	Uint8       uint8         `desc:"Uint8"`
	Uint16      uint16        `desc:"Uint16"`
	Uint32      uint32        `desc:"Uint32"`
	Uint64      uint64        `desc:"Uint64"`
	Uint        uint          `desc:"Uint"`
	Float32     float32       `desc:"Float32"`
	Float64     float64       `desc:"Float64"`
	String      string        `desc:"String"`
	StringSlice []string      `desc:"StringSlice"`
	Bool        bool          `desc:"Bool"`
	Duration    time.Duration `desc:"Duration"`
}

type TestOption struct {
	AnotherOption
	HTTP          string `desc:"HTTP"`
	HTTPPort      int    `desc:"HTTPPort"`
	HTTPPort2     int    `desc:"HTTPPort2"`
	HTTPPort233a  int    `desc:"HTTPPort233a"`
	HTTPPort233Aa int    `desc:"HTTPPort233Aa"`
	HTTP2         string `desc:"HTTP2"`
	HTTP2Port     int    `desc:"HTTP2Port"`
	SomeHTTP      string `desc:"SomeHTTP"`
	SomeHTTPDesc  string `desc:"SomeHTTPDesc"`
}

// Name returns plugin name.
func (p *TestOption) Name() string {
	return "test"
}

// Configure configures nirvana config via current option.
func (p *TestOption) Configure(cfg *nirvana.Config) error {
	return nil
}

func TestCommand(t *testing.T) {
	o := &TestOption{
		AnotherOption{
			-1, -2, -3, -4, -5,
			6, 7, 8, 9, 10,
			100.123, 200.345,
			"test", []string{"eee", "123"},
			true, time.Second,
		},
		"xxx",
		1, 2, 3, 4,
		"xxx2",
		5,
		"sxxx", "sxxxd",
	}
	fields := []configField{
		{&o.HTTP, o.HTTP, "test.http", "TEST_HTTP", "", "test-http", "HTTP"},
		{&o.HTTPPort, o.HTTPPort, "test.httpPort", "TEST_HTTP_PORT", "", "test-http-port", "HTTPPort"},
		{&o.HTTPPort2, o.HTTPPort2, "test.httpPort2", "TEST_HTTP_PORT2", "", "test-http-port2", "HTTPPort2"},
		{&o.HTTPPort233a, o.HTTPPort233a, "test.httpPort233a", "TEST_HTTP_PORT233A", "", "test-http-port233a", "HTTPPort233a"},
		{&o.HTTPPort233Aa, o.HTTPPort233Aa, "test.httpPort233Aa", "TEST_HTTP_PORT233_AA", "", "test-http-port233-aa", "HTTPPort233Aa"},
		{&o.HTTP2, o.HTTP2, "test.http2", "TEST_HTTP2", "", "test-http2", "HTTP2"},
		{&o.HTTP2Port, o.HTTP2Port, "test.http2Port", "TEST_HTTP2_PORT", "", "test-http2-port", "HTTP2Port"},
		{&o.SomeHTTP, o.SomeHTTP, "test.someHTTP", "TEST_SOME_HTTP", "", "test-some-http", "SomeHTTP"},
		{&o.SomeHTTPDesc, o.SomeHTTPDesc, "test.someHTTPDesc", "TEST_SOME_HTTP_DESC", "", "test-some-http-desc", "SomeHTTPDesc"},

		{&o.Int8, o.Int8, "test.int8", "TEST_INT8", "", "test-int8", "Int8"},
		{&o.Int16, o.Int16, "test.int16", "TEST_INT16", "", "test-int16", "Int16"},
		{&o.Int32, o.Int32, "test.int32", "TEST_INT32", "", "test-int32", "Int32"},
		{&o.Int64, o.Int64, "test.int64", "TEST_INT64", "", "test-int64", "Int64"},
		{&o.Int, o.Int, "test.int", "TEST_INT", "", "test-int", "Int"},
		{&o.Uint8, o.Uint8, "test.uint8", "TEST_UINT8", "", "test-uint8", "Uint8"},
		{&o.Uint16, o.Uint16, "test.uint16", "TEST_UINT16", "", "test-uint16", "Uint16"},
		{&o.Uint32, o.Uint32, "test.uint32", "TEST_UINT32", "", "test-uint32", "Uint32"},
		{&o.Uint64, o.Uint64, "test.uint64", "TEST_UINT64", "", "test-uint64", "Uint64"},
		{&o.Uint, o.Uint, "test.uint", "TEST_UINT", "", "test-uint", "Uint"},
		{&o.Float32, o.Float32, "test.float32", "TEST_FLOAT32", "", "test-float32", "Float32"},
		{&o.Float64, o.Float64, "test.float64", "TEST_FLOAT64", "", "test-float64", "Float64"},
		{&o.String, o.String, "test.string", "TEST_STRING", "", "test-string", "String"},
		{&o.StringSlice, o.StringSlice, "test.stringSlice", "TEST_STRING_SLICE", "", "test-string-slice", "StringSlice"},
		{&o.Bool, o.Bool, "test.bool", "TEST_BOOL", "", "test-bool", "Bool"},
		{&o.Duration, o.Duration, "test.duration", "TEST_DURATION", "", "test-duration", "Duration"},
	}
	cmd := NewDefaultNirvanaCommand()
	cmd.EnablePlugin(o)
	flags := cmd.Command(nirvana.NewDefaultConfig("", 8080)).Flags()
	c := cmd.(*command)
	for _, f := range fields {
		cf, ok := c.fields[f.key]
		if !ok {
			t.Logf("%+v", c.fields)
			t.Fatalf("Can't find key %s", f.key)
		}
		if cf.key != f.key {
			t.Fatalf("Inequality key for %s: %v", f.key, cf.key)
		}
		if cf.pointer != f.pointer {
			t.Fatalf("Inequality pointer for %s", f.key)
		}
		if !reflect.DeepEqual(cf.desired, f.desired) {
			t.Fatalf("Inequality default value for %s: %v != %v", f.key, cf.desired, f.desired)
		}
		if cf.env != f.env {
			t.Fatalf("Inequality env for %s: %v", f.key, cf.env)
		}
		if cf.shortFlag != f.shortFlag {
			t.Fatalf("Inequality short flag for %s: %v", f.key, cf.shortFlag)
		}
		if cf.longFlag != f.longFlag {
			t.Fatalf("Inequality long flag for %s: %v", f.key, cf.longFlag)
		}
		if cf.description != f.description {
			t.Fatalf("Inequality description for %s: %v", f.key, cf.description)
		}
		flag := flags.Lookup(f.longFlag)
		if flag.Name != f.longFlag {
			t.Fatalf("Inequality flag name for %s: %v", f.key, flag.Name)
		}
		if flag.Usage != fmt.Sprintf("%s (env %s) (cfg %s)", f.description, f.env, f.key) {
			t.Fatalf("Inequality flag usage for %s: %v", f.key, flag.Usage)
		}
	}
}

func TestChooseValue(t *testing.T) {
	values := []struct {
		key          string
		value        interface{}
		env          string
		envValue     interface{}
		defaultValue interface{}
		desc         string
		result       interface{}
	}{
		{"test.key", 1, "TEST_KEY", 2, 3, " (env TEST_KEY=2) (cfg test.key=1)", 2},
		{"test.key", 1, "TEST_KEY", nil, 3, " (env TEST_KEY) (cfg test.key=1)", 1},
		{"test.key", nil, "TEST_KEY", 2, 3, " (env TEST_KEY=2) (cfg test.key)", 2},
		{"test.key", nil, "TEST_KEY", nil, 3, " (env TEST_KEY) (cfg test.key)", 3},
	}
	for _, v := range values {
		desc, result := chooseValue(v.key, v.value, v.env, v.envValue, v.defaultValue)
		if desc != v.desc || result != v.result {
			t.Fatalf("chooseValue generates wrong results: %s, %v", desc, result)
		}
	}
}
