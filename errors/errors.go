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

// StatusReason is an enumeration of possible failure causes. Each StatusReason
// must map to a single HTTP status code, but multiple reasons may map to the
// same HTTP status code.
type StatusReason string

// Error is an error intended to be used by all caicloud components to return
// error to clients.
type Error struct {
	// Suggested HTTP return code for this error, 0 if not set. This field is optional.
	// Caller can choose to use this code or choose to use another error code for client.
	statusCode int32

	// Required, a machine-readable description of the error.
	Reason StatusReason `json:"reason"`
	// Required for 4xx but optional for 5xx. Message is a human-readable description
	// of the error. Message can be golang template.
	Message string `json:"message"`
	// Required when template is used in message or i18nMessage.
	Data map[string]string `json:"data,omitempty"`
}
