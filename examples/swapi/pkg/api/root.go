package api

import (
	"github.com/caicloud/nirvana/web"
	"github.com/caicloud/nirvana/examples/swapi/pkg/loader"
	"github.com/caicloud/nirvana/examples/swapi/pkg/api/v1"
)

func CreateWebServer(model loader.ModelLoader) web.Server {
	if err := web.RegisterDefaultEnvironment(); err != nil {
		panic(err)
	}
	s := web.NewDefaultServer()
	if err := s.AddDescriptors(v1.API(model)); err != nil {
		panic(err)
	}
	return s
}
