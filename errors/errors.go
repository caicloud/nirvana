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

package errors

import (
	"fmt"
)

// Reason is an enumeration of possible failure causes. Each Reason
// must map to a format which is a string containing ${formatArgu1}.
// Exp:
// Reason "kind:NotFound" may map to Format "${kindName} was not found".
// Reason "Status:Sleep" may map to Format "${Name} is sleeping now"
type Reason string

// message was the result returned to client.
// Exp:
// {
//   "message": "name japari is too short",
//   "reason": "monitoring:CreateDashboardNameTooShort",
//   "data": {
//     "name": "japari"
//   }
// }
type message struct {
	Reason  Reason            `json:"reason,omitempty"`
	Message string            `json:"message"`
	Data    map[string]string `json:"data,omitempty"`
}

type err struct {
	message
	code   int
	format string
}

func (e *err) Code() int {
	return e.code
}

func (e *err) Message() interface{} {
	return e.message
}

func (e *err) Error() string {
	return e.message.Message
}

// Factory is an error factory.
type Factory struct {
	code   int
	reason Reason
	format string
}

// New generates an error.
func (f *Factory) New(v ...interface{}) error {
	msg := message{Reason: f.reason}
	msg.Message, msg.Data = expand(f.format, v...)
	return &err{
		message: msg,
		code:    f.code,
		format:  f.format,
	}
}

// CanNew checks whether f is able to New an error which has the same code, reason and format with e.
func (f *Factory) CanNew(e error) bool {
	x, ok := e.(*err)
	if !ok {
		return false
	}
	return f.code == x.code && f.reason == x.message.Reason && f.format == x.format
}

// Type maps to http code.
// And it can be used to make an error factory.
type Type struct {
	code int
}

// NewType creates a new Type with code.
func NewType(code int) *Type {
	return &Type{code: code}
}

// NewFactory creates a factory to generate errors with predefined format.
func (t *Type) NewFactory(reason Reason, format string) *Factory {
	return &Factory{code: t.code, reason: reason, format: format}
}

// NewRaw creates an error which composed by code, reason and formated message in one call.
func NewRaw(code int, reason Reason, format string, v ...interface{}) error {
	return NewType(code).NewFactory(reason, format).New(v...)
}

// expand expands a format string like "name ${name} is too short" to "name japari is too short"
// by replacing ${} with v... one by one.
// Note that if len(v) < count of ${}, it will panic.
func expand(format string, v ...interface{}) (msg string, data map[string]string) {
	n := 0
	var m map[string]string
	buf := make([]byte, 0, len(format))

	for i := 0; i < len(format); {
		if format[i] == '$' && (i+1) < len(format) && format[i+1] == '{' {
			b := make([]byte, 0, len(format)-i)
			if i+2 == len(format) { // check "...${"
				panic("unexpected EOF while looking for matching }")
			}
			ii := i + 2
			for ii < len(format[i+2:])+i+2 {
				if format[ii] != '}' {
					b = append(b, format[ii])
				} else {
					break
				}
				ii++
				if ii == len(format[i+2:])+i+2 { // check "...${..."
					panic("unexpected EOF while looking for matching }")
				}
			}
			i = ii + 1
			if n == len(v) {
				panic("not enough args")
			}
			if m == nil {
				m = map[string]string{}
			}
			m[string(b)] = fmt.Sprint(v[n])
			buf = append(buf, fmt.Sprint(v[n])...)
			n++
		} else {
			buf = append(buf, format[i])
			i++
		}
	}
	return string(buf), m
}
