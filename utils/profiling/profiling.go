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

package profiling

import (
	"context"
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
	rpprof "runtime/pprof"
	"strings"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

// Profiling adds handlers for pprof under /debug/pprof.
type Profiling struct{}

func pprofHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/debug/pprof/") {
		name := strings.TrimPrefix(r.URL.Path, "/debug/pprof/")
		switch name {
		case "profile":
			pprof.Profile(w, r)
		case "symbol":
			pprof.Symbol(w, r)
		case "trace":
			pprof.Trace(w, r)
		default:
			pprof.Index(w, r)
		}
		return
	}
	pprofIndex(w, r)
}

// Install adds the Profiling service to the given mux.
func (d Profiling) Install(s service.Server) {
	h := http.HandlerFunc(pprofHandler)
	if err := s.AddDescriptors(
		convertHandlerToDescriptor("/debug/pprof", h),
		convertHandlerToDescriptor("/debug/pprof/{path}", h),
	); err != nil {
		panic(err)
	}
}

func pprofIndex(w http.ResponseWriter, r *http.Request) {
	profiles := rpprof.Profiles()
	if err := indexTmpl.Execute(w, profiles); err != nil {
		log.Print(err)
	}
}

// indexTmpl is modified from http/pprof/pprof.go, adding 'pprof/' prefix for all href content
// go pprof http index page served the path '/debug/pprof/' which will be redirected to '/debug/pprof'
// by nirvana
var indexTmpl = template.Must(template.New("index").Parse(`<html>
	<head>
	<title>/debug/pprof/</title>
	</head>
	<body>
	/debug/pprof/<br>
	<br>
	profiles:<br>
	<table>
	{{range .}}
	<tr><td align=right>{{.Count}}<td><a href="pprof/{{.Name}}?debug=1">{{.Name}}</a>
	{{end}}
	</table>
	<br>
	<a href="pprof/goroutine?debug=2">full goroutine stack dump</a><br>
	</body>
	</html>
	`))

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
