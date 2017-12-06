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

package main

import (
	"context"
	"errors"
	"net/http"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/router"
	"github.com/caicloud/nirvana/trace"
	"github.com/caicloud/nirvana/web"
)

func main() {
	if err := web.RegisterDefaultEnvironment(); err != nil {
		panic(err)
	}

	s := web.NewDefaultServer()

	tracer, ioClose := trace.NewDefaultTracerClient("example", "127.0.0.1:6831")
	defer ioClose.Close()

	cfg := &trace.Config{
		Tracer: tracer,
	}

	example := definition.Descriptor{
		Path:        "/",
		Description: "trace example",
		Middlewares: []router.Middleware{cfg.New()},
		Definitions: []definition.Definition{
			{
				Method: definition.Get,
				Function: func(ctx context.Context) (string, error) {
					msg := web.HTTPRequest(ctx).URL.Query().Get("msg")
					if msg != "" {
						return "", errors.New(msg)
					}
					return "success", nil
				},
				Consumes: []string{"application/json"},
				Produces: []string{"application/json"},
				Results: []definition.Result{
					{
						Type: definition.Data,
					},
					{
						Type: definition.Error,
					},
				},
			},
		},
	}
	s.AddDescriptors(example)
	http.ListenAndServe(":8080", s)
}
