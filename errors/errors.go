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
	"net/http"
)

type Error interface {
	Reason() Reason
	Error() string
}

// Reason is an enumeration of possible failure causes. Each Reason
// must map to a single HTTP status code, but multiple reasons may map to the
// same HTTP status code.
type Reason string

type message struct {
	// Required, a machine-readable description of the error.
	Reason Reason `json:"reason"`
	// Required when template is used in message or i18nMessage.
	Data []interface{} `json:"data,omitempty"`
	// Required for 4xx but optional for 5xx. Message is a human-readable description
	// of the error. Message can be golang template.
	Message string `json:"message"`
}

// err is an error intended to be used by all APIs to return error to clients.
type err struct {
	// Suggested HTTP return code for this error, 0 if not set. This field is optional.
	// Caller can choose to use this code or choose to use another error code for client.
	code int
	// Format is used to generate message.
	format string
	// Useful message.
	message message
}

func (e *err) Code() int {
	return e.code
}

func (e *err) Message() interface{} {
	return &e.message
}

func (e *err) Reason() Reason {
	return e.message.Reason
}

func (e *err) Error() string {
	return e.message.Message
}

type formatter struct {
	code   int
	reason Reason
	// It should like
	// 1. something named {0} is not found
	// 2. something named {name} is not found
	// Which one will win the battle?
	format string
}

func (f *formatter) Format(a ...interface{}) Error {
	return &err{
		code:   f.code,
		format: f.format,
		message: message{
			Reason:  f.reason,
			Data:    a,
			Message: fmt.Sprintf(f.format, a...),
		},
	}
}

func ResourceNotFound(resource string) Error {
	return (&formatter{
		code:   http.StatusNotFound,
		reason: Reason(http.StatusText(http.StatusNotFound)),
		format: "%s is not found",
	}).Format(resource)
}
