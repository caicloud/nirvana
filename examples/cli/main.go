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
	"fmt"
	"os"

	"github.com/caicloud/nirvana/cli"
	"github.com/spf13/cobra"
)

var (
	// NeverStop may be passed to Until to make it never stop.
	NeverStop <-chan struct{} = make(chan struct{})
)

func main() {

	log := new(string)
	dev := new(string)
	cmd := cli.NewCommand(&cobra.Command{
		Use:  "example",
		Long: "this is an cli example",
		Run: func(cmd *cobra.Command, args []string) {
			flog := cmd.LocalFlags().Lookup("log")
			fdev := cmd.PersistentFlags().Lookup("dev")
			fmt.Printf("log: %v\n", *log)
			fmt.Printf("flag.log: %v\n", flog.Value.String())
			fmt.Printf("viper.log: %v\n", cli.GetString("log"))

			fmt.Printf("dev: %v\n", *dev)
			fmt.Printf("flag.dev: %v\n", fdev.Value.String())
			fmt.Printf("viper.dev: %v\n", cli.GetString("dev"))
		},
	})

	fs := []cli.Flag{
		cli.StringFlag{
			Name:        "log",
			Shorthand:   "l",
			EnvKey:      "LOG",
			Destination: log,
		},
		// hidden
		cli.StringFlag{
			Name:                "dev",
			Shorthand:           "d",
			EnvKey:              "DEV",
			Persistent:          true,
			Deprecated:          "move",
			ShorthandDeprecated: "move2",
			Hidden:              true,
			DefValue:            "default value",
			Destination:         dev,
		},
		cli.StringFlag{
			Name:      "auto",
			Shorthand: "a",
		},
	}

	// set env
	os.Setenv("LOG", "test")

	cli.AutomaticEnv()
	cli.SetEnvPrefix("example")
	cmd.AddFlag(fs...)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
