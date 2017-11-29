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
	goflag "flag"
	"os"
	"path/filepath"

	"github.com/caicloud/nirvana/cmd/flag-gen/generators"
	"github.com/spf13/pflag"

	"github.com/golang/glog"
	"k8s.io/gengo/args"
)

func main() {
	arguments := &args.GeneratorArgs{
		OutputBase:       args.DefaultSourceTree(),
		GoHeaderFilePath: filepath.Join(args.DefaultSourceTree(), "github.com/caicloud/nirvana/boilerplate/boilerplate.go.txt"),
	}
	arguments.AddFlags(pflag.CommandLine)
	pflag.CommandLine.AddGoFlagSet(goflag.CommandLine)
	if err := pflag.Set("logtostderr", "true"); err != nil {
		panic(err)
	}
	pflag.Parse()
	if err := goflag.CommandLine.Parse(nil); err != nil {
		panic(err)
	}

	// Run it
	if err := arguments.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	); err != nil {
		glog.Fatalf("Error: %v", err)
		os.Exit(1)
	}

	glog.V(2).Info("Completed successfully.")
}
