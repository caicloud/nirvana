/*
Copyright 2020 Caicloud Authors

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

package main

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"

	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/middlewares/reqlog"
	"github.com/caicloud/nirvana/service"
)

func main() {
	backendServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "this call was relayed by the reverse proxy")
	}))
	defer backendServer.Close()

	rpURL, err := url.Parse(backendServer.URL)
	if err != nil {
		log.Fatal(err)
	}

	descriptors := []definition.Descriptor{
		definition.RESTfulDescriptor{
			Path:        "/",
			Description: "hello API",
			Definitions: []definition.Definition{
				{
					Method:   definition.Get,
					Consumes: []string{definition.MIMEAll},
					Produces: []string{definition.MIMEJSON},
					Function: func(ctx context.Context) (string, error) {
						return "hello", nil
					},
					Results: definition.DataErrorResults(""),
				},
			},
		},
		definition.RESTfulDescriptor{
			Path:        "/proxy",
			Description: "proxy API",
			Middlewares: []definition.Middleware{
				reqlog.Default(),
			},
			Definitions: []definition.Definition{
				{
					Method:   definition.Get,
					Consumes: []string{definition.MIMEAll},
					Produces: []string{definition.MIMEAll},
					Function: service.WrapHTTPHandler(httputil.NewSingleHostReverseProxy(rpURL)),
				},
			},
		},
	}

	cmd := config.NewDefaultNirvanaCommand()
	if err = cmd.Execute(descriptors...); err != nil {
		log.Fatal(err)
	}
}
