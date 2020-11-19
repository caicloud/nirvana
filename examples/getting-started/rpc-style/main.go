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
	"github.com/caicloud/nirvana/middlewares/reqlog"
	"github.com/caicloud/nirvana/service/builder"
)

var echo = definition.RPCDescriptor{
	Description: "Echo API",
	Middlewares: []definition.Middleware{reqlog.Default()},
	Consumes:    []string{definition.MIMEAll},
	Produces:    []string{definition.MIMEJSON},
	Actions: []definition.RPCAction{
		{
			Name:    "Echo",
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
			Name:    "Echo1",
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
	Prefix:      "/aaa",
	Description: "Echo API",
	Actions: []definition.RPCAction{
		{
			Name:    "Echo2",
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
		nirvana.APIStyle(builder.APIStyleRPC),
		nirvana.Descriptor(echo, echo2),
	)
	if err := cmd.ExecuteWithConfig(conf); err != nil {
		log.Fatal(err)
	}
}
