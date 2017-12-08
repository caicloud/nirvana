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
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

// LoadGoBoilerplate loads the boilerplate file passed to --go-header-file.
func loadGoBoilerplate(filepath string) ([]byte, error) {
	b, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	b = bytes.Replace(b, []byte("YEAR"), []byte(strconv.Itoa(time.Now().Year())), -1)
	return b, nil
}

// Run ...
func Run() {
	root := cli.GetString("root")

	boilerplate, err := loadGoBoilerplate(cli.GetString("go-header-file"))
	if err != nil {
		glog.Fatal(err)
		return
	}
	boilerplate = bytes.TrimSpace(boilerplate)
	// add one empty line
	license := append(boilerplate, '\n', '\n')

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
		// TODO: support bash and python file
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

			if !cli.GetBool("dryRun") {
				if err := ioutil.WriteFile(path, append(license, allFile[i:]...), 0655); err != nil {
					panic(err)
				}
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
	if err := goflag.CommandLine.Parse(nil); err != nil {
		panic(err)
	}

	cmd := cli.NewCommand(&cobra.Command{
		Use:  "license",
		Long: "Read Apache 2.0 LICENSE content and add to to all go source code header",
		Run: func(cmd *cobra.Command, args []string) {
			Run()
		},
	})

	if err := cmd.AddFlag(
		cli.BoolFlag{
			Name:      "dryRun",
			Shorthand: "d",
		},
		cli.StringFlag{
			Name:      "root",
			Shorthand: "r",
			DefValue:  "./",
		},
		cli.StringFlag{
			Name: "go-header-file",
		},
	); err != nil {
		panic(err)
	}

	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
