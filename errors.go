/*
Copyright 2017 caicloud authors. All rights reserved.
*/

package nirvana

// StatusReason is an enumeration of possible failure causes. Each StatusReason
// must map to a single HTTP status code, but multiple reasons may map to the
// same HTTP status code.
type StatusReason string

// Error is an error intended to be used by all caicloud components
// to return error to clients.
type Error struct {
	// Required, a human-readable description of the error. Message can be a template.
	Message string `json:"message"`
	// Required for 4xx, a machine-readable description of the error. Reason
	Reason StatusReason `json:"reason,omitempty"`
	// Required for 4xx when template is used in message or i18nMessage.
	Data map[string]string `json:"data,omitempty"`
	// Suggested HTTP return code for this error, 0 if not set. This field is optional.
	// Caller can choose to use this code or choose to use another error code for client.
	Code int32 `json:"code,omitempty"`
}
