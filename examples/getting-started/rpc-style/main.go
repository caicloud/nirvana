/*
Copyright 2018 Caicloud Authors

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

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/config"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/middlewares/metrics"
	"github.com/caicloud/nirvana/middlewares/reqlog"
	"github.com/caicloud/nirvana/plugins/apidocs"
	"github.com/caicloud/nirvana/plugins/healthcheck"
	"github.com/caicloud/nirvana/plugins/profiling"
	"github.com/caicloud/nirvana/service"
)

var echo = definition.RPCDescriptor{
	Description: "Echo API",
	Middlewares: []definition.Middleware{reqlog.Default()},
	Consumes:    []string{definition.MIMEAll},
	Produces:    []string{definition.MIMEJSON},
	Actions: []definition.RPCAction{
		{
			Name:    "GetEcho",
			Version: "2020-10-10",
			Function: func(ctx context.Context, msg string) (string, error) {
				return msg, nil
			},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "msg",
					Description: "Corresponding to the second parameter",
				},
			},
			Results: definition.DataErrorResults(""),
		},
		{
			Name:    "CreateEcho",
			Version: "2020-10-10",
			Function: func(ctx context.Context, msg string) (string, error) {
				return msg, nil
			},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "msg",
					Description: "Corresponding to the second parameter",
				},
			},
			Results: definition.DataErrorResults(""),
		},
	},
}

var echo2 = definition.RPCDescriptor{
	Path:        "/aaa",
	Description: "Echo API",
	Actions: []definition.RPCAction{
		{
			Name:    "DeleteEcho",
			Version: "2020-10-10",
			Function: func(ctx context.Context, msg string) (string, error) {
				return msg, nil
			},
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEJSON},
			Parameters: []definition.Parameter{
				{
					Source:      definition.Query,
					Name:        "msg",
					Description: "Corresponding to the second parameter",
				},
			},
			Results: definition.DataErrorResults(""),
		},
	},
}

func main() {
	cmd := config.NewDefaultNirvanaCommand()
	conf := nirvana.NewDefaultConfig()
	conf.Configure(
		nirvana.APIStyle(service.APIStyleRPC),
		nirvana.Descriptor(echo, echo2, metrics.RPCDescriptor("/metrics")),

		profiling.Path("/debug/pprof"),
		healthcheck.CheckerWithType(func(ctx context.Context, checkType string) error {
			switch checkType {
			case healthcheck.LivenessCheck:
				return nil
			case healthcheck.ReadinessCheck:
				// add specific check logic here if needed
				return nil
			}
			return nil
		}),
		// enable the API docs plugin
		func(c *nirvana.Config) error {
			plugin := &apidocs.Option{FilesPath: "apis", Path: "/docs"}
			return plugin.Configure(c)
		},
	)
	if err := cmd.ExecuteWithConfig(conf); err != nil {
		log.Fatal(err)
	}
}
