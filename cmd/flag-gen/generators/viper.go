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

package generators

import (
	"io"

	"github.com/golang/glog"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
)

var _ generator.Generator = &viperGenerator{}

type viperGenerator struct {
	generator.DefaultGen
	outputPackage string
	imports       namer.ImportTracker
	typeToMatch   *types.Type
}

func (g *viperGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.typeToMatch
}

func (g *viperGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *viperGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	imports = append(imports, "github.com/spf13/viper")
	imports = append(imports, "github.com/spf13/cast")
	return
}

func (g *viperGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	glog.V(5).Infof("processing type %v", t)
	m := map[string]interface{}{
		"type": t,
	}
	sw.Do(viperCode, m)

	return sw.Error()
}

var viperCode = `
// Get$.type|public$ returns the value associated with the key as a $.type|raw$.
func Get$.type|public$(key string) $.type|raw$ {
	return cast.To$.type|public$(v.Get(key))
}
`

var _ generator.Generator = &viperTestGenerator{}

type viperTestGenerator struct {
	generator.DefaultGen
	outputPackage string
	imports       namer.ImportTracker
	typeToMatch   *types.Type
}

func (g *viperTestGenerator) Filter(c *generator.Context, t *types.Type) bool {
	return t == g.typeToMatch
}

func (g *viperTestGenerator) Namers(c *generator.Context) namer.NameSystems {
	return namer.NameSystems{
		"raw": namer.NewRawNamer(g.outputPackage, g.imports),
	}
}

func (g *viperTestGenerator) Imports(c *generator.Context) (imports []string) {
	imports = append(imports, g.imports.ImportLines()...)
	imports = append(imports, "testing")
	imports = append(imports, "reflect")
	return
}

func (g *viperTestGenerator) GenerateType(c *generator.Context, t *types.Type, w io.Writer) error {
	sw := generator.NewSnippetWriter(w, c, "$", "$")
	glog.V(5).Infof("processing type %v", t)
	m := map[string]interface{}{
		"type": t,
	}
	sw.Do(viperTestCode, m)

	return sw.Error()
}

var viperTestCode = `
func TestGet$.type|public$(t *testing.T) {
	Reset()
	testcase := getTestCase("$.type|public$")
	key := "dev"
	v.Set(key, testcase.want)
	want, err := cast.To$.type|public$E(testcase.want)
	assert.Nil(t, err)
	if got := Get$.type|public$(key); !reflect.DeepEqual(got, want) {
		assert.Equal(t, want, got)
	}
}
`
