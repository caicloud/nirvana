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

package api

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/cmd/nirvana/buildutils"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/api"
	"github.com/caicloud/nirvana/utils/generators/swagger"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func init() {
	err := service.RegisterProducer(service.NewSimpleSerializer("text/html"))
	if err != nil {
		log.Fatalln(err)
	}
}

func newAPICommand() *cobra.Command {
	options := &apiOptions{}
	cmd := &cobra.Command{
		Use:   "api /path/to/apis",
		Short: "Generate API documents for your project",
		Long:  options.Manuals(),
		Run: func(cmd *cobra.Command, args []string) {
			if err := options.Validate(cmd, args); err != nil {
				log.Fatalln(err)
			}
			if err := options.Run(cmd, args); err != nil {
				log.Fatalln(err)
			}
		},
	}
	options.Install(cmd.PersistentFlags())
	return cmd
}

type apiOptions struct {
	Serve  string
	Output string
}

func (o *apiOptions) Install(flags *pflag.FlagSet) {
	flags.StringVar(&o.Serve, "serve", "127.0.0.1:8080", "Start a server to host api docs")
	flags.StringVar(&o.Output, "output", "", "Directory to output api specifications")
}

func (o *apiOptions) Validate(cmd *cobra.Command, args []string) error {
	return nil
}

func (o *apiOptions) Run(cmd *cobra.Command, args []string) error {
	if len(args) <= 0 {
		defaultAPIsPath := "pkg"
		args = append(args, defaultAPIsPath)
		log.Infof("No packages are specified, defaults to %s", defaultAPIsPath)
	}

	config, definitions, err := buildutils.Build(args...)
	if err != nil {
		return err
	}

	log.Infof("Project root directory is %s", config.Root)

	generator := swagger.NewDefaultGenerator(config, definitions)
	swaggers, err := generator.Generate()
	if err != nil {
		return err
	}

	files := make(map[string][]byte, len(swaggers))
	for filename, s := range swaggers {
		data, err := json.MarshalIndent(s, "", "  ")
		if err != nil {
			return err
		}

		files[filename] = data
	}

	if o.Output != "" {
		if err = api.WriteFiles(o.Output, files); err != nil {
			return err
		}
	}

	if o.Serve != "" {
		err = o.serve(files)
	}
	return err
}

func (o *apiOptions) serve(apis map[string][]byte) error {
	hosts := strings.Split(o.Serve, ":")
	ip := strings.TrimSpace(hosts[0])
	if ip == "" {
		ip = "127.0.0.1"
	}
	port := uint16(8080)
	if len(hosts) >= 2 {
		p := strings.TrimSpace(hosts[1])
		if p != "" {
			pt, err := strconv.Atoi(p)
			if err != nil {
				return err
			}
			port = uint16(pt)
		}
	}
	log.SetDefaultLogger(log.NewStdLogger(0))
	cfg := nirvana.NewDefaultConfig()
	versions := make([]string, 0, len(apis))
	for v, data := range apis {
		versions = append(versions, v)
		cfg.Configure(nirvana.Descriptor(
			api.DescriptorForData(api.PathForVersion("/", v), data, definition.MIMEJSON),
		))
	}
	data, err := api.GenSwaggerPageData("/", versions)
	if err != nil {
		return err
	}
	cfg.Configure(
		nirvana.Descriptor(api.DescriptorForData("/", data, definition.MIMEHTML)),
		nirvana.IP(ip),
		nirvana.Port(port),
	)
	log.Infof("Listening on %s:%d. Please open your browser to view api docs", cfg.IP(), cfg.Port())
	return nirvana.NewServer(cfg).Serve()
}

func (o *apiOptions) Manuals() string {
	return ""
}
