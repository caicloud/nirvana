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
	"path/filepath"

	"github.com/caicloud/nirvana/cmd/openapi-gen/generators"
	"github.com/golang/glog"
	"k8s.io/gengo/args"
)

func main() {
	arguments := args.Default()

	arguments.OutputFileBaseName = "openapi_generated"
	arguments.GoHeaderFilePath = filepath.Join(args.DefaultSourceTree(), "github.com/caicloud/nirvana/hack/boilerplate/boilerplate.go.txt")

	if err := arguments.Execute(
		generators.NameSystems(),
		generators.DefaultNameSystem(),
		generators.Packages,
	); err != nil {
		glog.Fatalf("Error: %v", err)
	}

	glog.V(2).Info("Completed successfully.")
}
