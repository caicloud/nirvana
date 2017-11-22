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
	"bytes"
	goflag "flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/caicloud/nirvana/cli"
	"github.com/golang/glog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var sentinels = []string{
	"Copyright",
	"Caicloud",
	`Licensed under the Apache License, Version 2.0 (the "License");`,
}

// Run ...
func Run() {
	root := cli.Viper.GetString("root")

	licenseBytes, err := ioutil.ReadFile(root + "/LICENSE")
	if err != nil {
		glog.Fatal(err)
		return
	}

	license := []byte(fmt.Sprintf("/*\n%s*/\n\n", licenseBytes))

	err = filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// skip vendor
		if info.IsDir() &&
			(strings.Contains(path, "vendor") || strings.Contains(path, ".git")) {
			return filepath.SkipDir
		}

		if info.IsDir() {
			return nil
		}

		// skip not go file
		if ext := filepath.Ext(path); ext != ".go" {
			// log.Infof("Skip file: %s", path)
			return nil
		}

		allFile, err := ioutil.ReadFile(path)
		if err != nil {
			return nil
		}

		src := allFile[:150]

		needLicense := false

		for _, sentinel := range sentinels {
			if !bytes.Contains(src, []byte(sentinel)) {
				needLicense = true
			}
		}

		if needLicense {
			glog.Infof("Add License to file: %s", path)

			i := bytes.Index(allFile, []byte("package"))

			if !cli.Viper.GetBool("dryRun") {
				_ = ioutil.WriteFile(path, append(license, allFile[i:]...), 0655)
			}
			return nil
		}

		glog.Infof("Skip file: %s", path)

		return nil
	})

	if err != nil {
		glog.Error(err)
	}
}

func main() {
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	_ = goflag.CommandLine.Parse(nil)

	cmd := cli.NewCommand(&cobra.Command{
		Use:  "license",
		Long: "Read Apache 2.0 LICENSE content and add to to all go source code header",
		Run: func(cmd *cobra.Command, args []string) {
			Run()
		},
	})

	_ = cmd.AddFlag(
		cli.BoolFlag{
			Name:      "dryRun",
			Shorthand: "d",
		},
		cli.StringFlag{
			Name:      "root",
			Shorthand: "r",
			DefValue:  "./",
		},
	)

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
