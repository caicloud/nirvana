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
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/http/pprof"
	"path/filepath"
	"runtime"
	rpprof "runtime/pprof"
	"sort"
	"strconv"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/service"
)

func init() {
	nirvana.RegisterConfigInstaller(&profilingInstaller{})
}

// ExternalConfigName is the external config name of profiling.
const ExternalConfigName = "profiling"

// config is profiling config.
type config struct {
	path string
}

type profilingInstaller struct{}

// Name is the external config name.
func (i *profilingInstaller) Name() string {
	return ExternalConfigName
}

// Install installs config to builder.
func (i *profilingInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {
	var err error
	wrapper(cfg, func(c *config) {
		if err = builder.AddDescriptor(descriptor(c.path)); err != nil {
			return
		}
	})
	return err
}

// Uninstall uninstalls stuffs after server terminating.
func (i *profilingInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {
	return nil
}

// Disable returns a configurer to disable profiling.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}

func wrapper(c *nirvana.Config, f func(c *config)) {
	conf := c.Config(ExternalConfigName)
	var cfg *config
	if conf == nil {
		// Default config.
		cfg = &config{
			path: "/debug/pprof/",
		}
	} else {
		// Panic if config type is wrong.
		cfg = conf.(*config)
	}
	f(cfg)
	c.Set(ExternalConfigName, cfg)
}

// Path returns a configurer to set metrics path.
// Default path is /debug/pprof.
// Then these path is used:
//   /debug/pprof/cmdline
//   /debug/pprof/profile
//   /debug/pprof/symbol
//   /debug/pprof/trace
//   /debug/pprof/{prof}
func Path(path string) nirvana.Configurer {
	if path == "" {
		path = "/debug/pprof/"
	}
	return func(c *nirvana.Config) error {
		wrapper(c, func(c *config) {
			c.path = path
		})
		return nil
	}
}

// descriptor creates descriptor for profiling.
func descriptor(path string) definition.Descriptor {
	return definition.Descriptor{
		Path:     path,
		Consumes: []string{definition.MIMEAll},
		Produces: []string{definition.MIMEAll},
		Definitions: []definition.Definition{{
			Method:   definition.Get,
			Function: service.WrapHTTPHandlerFunc(index),
		}},
		Children: []definition.Descriptor{
			{
				Path: "cmdline",
				Definitions: []definition.Definition{{
					Method:   definition.Get,
					Function: service.WrapHTTPHandlerFunc(pprof.Cmdline),
				}},
			},
			{
				Path: "profile",
				Definitions: []definition.Definition{{
					Method:   definition.Get,
					Function: service.WrapHTTPHandlerFunc(pprof.Profile),
				}},
			},
			{
				Path: "symbol",
				Definitions: []definition.Definition{{
					Method:   definition.Get,
					Function: service.WrapHTTPHandlerFunc(pprof.Symbol),
				}},
			},
			{
				Path: "trace",
				Definitions: []definition.Definition{{
					Method:   definition.Get,
					Function: service.WrapHTTPHandlerFunc(pprof.Trace),
				}},
			},
			{
				Path: "{prof}",
				Definitions: []definition.Definition{{
					Method: definition.Get,
					Function: service.WrapHTTPHandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						name := filepath.Base(r.URL.Path)
						w.Header().Set("X-Content-Type-Options", "nosniff")
						p := rpprof.Lookup(name)
						if p == nil {
							w.WriteHeader(http.StatusNotFound)
							_, _ = w.Write([]byte("Unknown profile"))
							return
						}
						gc, _ := strconv.Atoi(r.FormValue("gc"))
						if name == "heap" && gc > 0 {
							runtime.GC()
						}
						debug, _ := strconv.Atoi(r.FormValue("debug"))
						if debug != 0 {
							w.Header().Set("Content-Type", "text/plain; charset=utf-8")
						} else {
							w.Header().Set("Content-Type", "application/octet-stream")
							w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, name))
						}
						_ = p.WriteTo(w, debug)
					}),
				}},
			},
		},
	}
}

var profileDescriptions = map[string]string{
	"allocs":    "A sampling of all past memory allocations",
	"block":     "Stack traces that led to blocking on synchronization primitives",
	"cmdline":   "The command line invocation of the current program",
	"goroutine": "Stack traces of all current goroutines",
	"heap":      "A sampling of memory allocations of live objects. You can specify the gc GET parameter to run GC before taking the heap sample.",
	"mutex":     "Stack traces of holders of contended mutexes",
	"profile": "CPU profile. You can specify the duration in the seconds GET parameter." +
		" After you get the profile file, use the go tool pprof command to investigate the profile.",
	"threadcreate": "Stack traces that led to the creation of new OS threads",
	"trace": "A trace of execution of the current program. You can specify the duration in the seconds GET parameter." +
		" After you get the trace file, use the go tool trace command to investigate the trace.",
	"symbol": "Symbols of program counters",
}

func index(w http.ResponseWriter, r *http.Request) {
	type profile struct {
		Name  string
		Href  string
		Desc  string
		Count int
	}
	var profiles []profile
	for _, p := range rpprof.Profiles() {
		profiles = append(profiles, profile{
			Name:  p.Name(),
			Href:  p.Name() + "?debug=1",
			Desc:  profileDescriptions[p.Name()],
			Count: p.Count(),
		})
	}

	// Adding other profiles exposed from within this package
	for _, p := range []string{"cmdline", "symbol", "profile", "trace"} {
		profiles = append(profiles, profile{
			Name: p,
			Href: p,
			Desc: profileDescriptions[p],
		})
	}

	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Name < profiles[j].Name
	})

	if err := indexTmpl.Execute(w, profiles); err != nil {
		log.Print(err)
	}
}

// indexTmpl is modified from http/pprof/pprof.go, adding 'pprof/' prefix for all href content
// go pprof http index page served the path '/debug/pprof/' which will be redirected to '/debug/pprof'
// by nirvana
var indexTmpl = template.Must(template.New("index").Parse(`<html>
<head>
<title>Profiling</title>
<style>
.profile-name{
	display:inline-block;
	width:6rem;
}
</style>
</head>
<body>
<br>
Types of profiles available:
<table>
<thead><td>Count</td><td>Profile</td></thead>
{{range .}}
	<tr>
	<td>{{.Count}}</td><td><a href="javascript:window.location.href=window.location.pathname+'/{{.Href}}';">{{.Name}}</a></td>
	</tr>
{{end}}
</table>
<a href="javascript:window.location.href=window.location.pathname+'/goroutine?debug=2'">full goroutine stack dump</a>
<br/>
<p>
Profile Descriptions:
<ul>
{{range .}}
<li><div class=profile-name>{{.Name}}:</div> {{.Desc}}</li>
{{end}}
</ul>
</p>
</body>
</html>
`))

// Option contains basic configurations of profiling.
type Option struct {
	// Path is profiling path.
	Path string `desc:"Profiling path"`
}

// NewDefaultOption creates default option.
func NewDefaultOption() *Option {
	return &Option{
		Path: "/debug/pprof/",
	}
}

// Name returns plugin name.
func (p *Option) Name() string {
	return ExternalConfigName
}

// Configure configures nirvana config via current options.
func (p *Option) Configure(cfg *nirvana.Config) error {
	cfg.Configure(
		Path(p.Path),
	)
	return nil
}
