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

package metrics

import (
	"context"
	"net/http"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics installs the default prometheus metrics handler
type Metrics struct{}

func (d Metrics) Install(s service.Server, metricsPath string) {
	if metricsPath == "" {
		metricsPath = "/metrics"
	}
	if err := s.AddDescriptors(convertHandlerToDescriptor(metricsPath, promhttp.Handler())); err != nil {
		panic(err)
	}
}

func convertHandlerToFunction(h http.Handler) interface{} {
	return func(ctx context.Context) {
		r := service.HTTPRequest(ctx)
		w := service.HTTPResponseWriter(ctx)
		h.ServeHTTP(w, r)
	}
}

func convertHandlerToDescriptor(path string, h http.Handler) definition.Descriptor {
	return definition.Descriptor{
		Path: path,
		Definitions: []definition.Definition{
			{
				Method:   definition.Get,
				Function: convertHandlerToFunction(h),
				Produces: []string{"text/plain", "application/octet-stream"},
			},
		},
	}
}
