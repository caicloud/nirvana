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

package api

import (
	"github.com/caicloud/nirvana/examples/swapi/pkg/api/v1"
	"github.com/caicloud/nirvana/examples/swapi/pkg/loader"
	"github.com/caicloud/nirvana/service"
)

func CreateWebServer(model loader.ModelLoader) service.Server {
	if err := service.RegisterDefaultEnvironment(); err != nil {
		panic(err)
	}
	s := service.NewDefaultServer()
	if err := s.AddDescriptors(v1.API(model)); err != nil {
		panic(err)
	}
	return s
}
