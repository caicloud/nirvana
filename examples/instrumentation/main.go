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
	"math/rand"
	"time"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/plugins/metrics"
	"github.com/caicloud/nirvana/plugins/profiling"
)

// This example shows how metrics and profiling plugin work, and the defaults functionality they provide.
// Run `ab -n 100 -H 'Content-type: application/json' http://localhost:8080/hello`, then
// curl `http://localhost:8080/metrics` to see default metrics for http requests.
func main() {
	config := nirvana.NewDefaultConfig("", 8080).
		Configure(
			profiling.Path("/debug/pprof/"),
			profiling.Contention(true),
			// By using metrics.Default configurer, metrics for http requests will be prefixed with the default 'nirvana_app' prefix.
			// If you want use a different prefix, use metrics.Namespace configurer.
			metrics.Default(),
			nirvana.Descriptor(example),
		)

	log.Infof("Listening on %s:%d", config.IP, config.Port)
	if err := nirvana.NewServer(config).Serve(); err != nil {
		log.Fatal(err)
	}
}

var example = definition.Descriptor{
	Path:        "/hello",
	Description: "metrics example",
	Definitions: []definition.Definition{
		{
			Method: definition.Get,
			Function: func(ctx context.Context) (string, error) {
				latency := 20 + rand.Float64()*300
				<-time.After(time.Duration(latency) * time.Millisecond)
				return "success", nil
			},
			Consumes: []string{"application/json"},
			Produces: []string{"application/json"},
			Results: []definition.Result{
				{
					Destination: definition.Data,
				},
				{
					Destination: definition.Error,
				},
			},
		},
	},
}
