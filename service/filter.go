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
	"mime"
	"net/http"
	"strconv"
	"strings"
)

// Filter can filter request. It has the highest priority in a request
// lifecycle. It runs before router matching.
// If a filter return false, that means the request should be filtered.
// If a filter want to filter a request, it should handle the request
// by itself.
type Filter func(resp http.ResponseWriter, req *http.Request) bool

// RedirectTrailingSlash returns a filter to redirect request.
// If a request has trailing slash like `some-url/`, the filter will
// redirect the request to `some-url`.
func RedirectTrailingSlash() Filter {
	return func(resp http.ResponseWriter, req *http.Request) bool {
		path := req.URL.Path
		if len(path) > 1 && path[len(path)-1] == '/' {
			req.URL.Path = strings.TrimRight(path, "/")
			// Redirect to path without trailing slash.
			http.Redirect(resp, req, req.URL.String(), http.StatusTemporaryRedirect)
			return false
		}
		return true
	}
}

// FillLeadingSlash is a pseudo filter. It only fills a leading slash when
// a request path does not have a leading slash.
// The filter won't filter anything.
func FillLeadingSlash() Filter {
	return func(resp http.ResponseWriter, req *http.Request) bool {
		path := req.URL.Path
		if len(path) <= 0 || path[0] != '/' {
			// Relative path may omit leading slash.
			req.URL.Path = "/" + path
		}
		return true
	}
}

// ParseRequestForm is a pseudo filter. It parse request form when content
// type is "application/x-www-form-urlencoded" or "multipart/form-data".
// The filter won't filter anything unless some error occurs in parsing.
func ParseRequestForm() Filter {
	return func(resp http.ResponseWriter, req *http.Request) bool {
		ct, err := ContentType(req)
		if err == nil {
			switch ct {
			case "application/x-www-form-urlencoded":
				err = req.ParseForm()
			case "multipart/form-data":
				err = req.ParseMultipartForm(32 << 20)
			default:
				req.Form = req.URL.Query()
			}
		}
		if err != nil {
			http.Error(resp, err.Error(), http.StatusBadRequest)
			return false
		}
		return true
	}
}

// ContentType is a util to get content type from a request.
func ContentType(req *http.Request) (string, error) {
	ct := req.Header.Get("Content-Type")
	if ct == "" {
		return "application/octet-stream", nil
	}
	ct, _, err := mime.ParseMediaType(ct)
	if err != nil {
		return "", err
	}
	return ct, nil
}

// AcceptTypes is a util to get accept types from a request.
// Accept types are sorted by q.
func AcceptTypes(req *http.Request) ([]string, error) {
	ct := req.Header.Get("Accept")
	if ct == "" {
		return []string{acceptTypeAll}, nil
	}
	return parseAcceptTypes(ct)
}

func parseAcceptTypes(v string) ([]string, error) {
	types := []string{}
	factors := []float64{}
	strs := strings.Split(v, ",")
	for _, str := range strs {
		str := strings.Trim(str, " ")
		tf := strings.Split(str, ";")
		types = append(types, tf[0])
		factor := 1.0
		if len(tf) == 2 {
			qp := strings.Split(tf[1], "=")
			q, err := strconv.ParseFloat(qp[1], 32)
			if err != nil {
				return nil, err
			}
			factor = q
		}
		factors = append(factors, factor)
	}
	if len(types) <= 1 {
		return types, nil
	}
	// In most cases, bubble sort is enough.
	// May be can optimize.
	exchanged := true
	for exchanged {
		exchanged = false
		for i := 1; i < len(factors); i++ {
			if factors[i] > factors[i-1] {
				types[i-1], types[i] = types[i], types[i-1]
				factors[i-1], factors[i] = factors[i], factors[i-1]
				exchanged = true
			}
		}
	}
	return types, nil
}
