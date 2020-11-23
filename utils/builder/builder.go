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

package builder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/types"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"text/template"

	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/utils/api"
)

var buildTagRegexp = regexp.MustCompile(`^[ \t]*\+nirvana:api[ \t]*=(.*)\n`)

// APIBuilder builds api definitions by specified package.
type APIBuilder struct {
	root  string
	paths []string
}

// NewAPIBuilder creates an api builder.
func NewAPIBuilder(root string, paths ...string) *APIBuilder {
	return &APIBuilder{
		root:  root,
		paths: paths,
	}
}

// Build builds api definitions.
func (b *APIBuilder) Build() (*api.Definitions, error) {
	analyzer, err := api.NewAnalyzer(b.root, b.paths...)
	if err != nil {
		return nil, err
	}

	descriptors := make([]function, 0)
	modifiers := make([]function, 0)

	apiStyle := string(service.APIStyleREST)

	for _, pkg := range analyzer.Paths() {
		groups := analyzer.PackageComments(pkg)
		for _, group := range groups {
			matches := buildTagRegexp.FindAllStringSubmatch(group.Text(), -1)
			for _, match := range matches {
				if len(match) == 2 {
					tag := reflect.StructTag(match[1])
					descriptorFunc := tag.Get("descriptors")
					if descriptorFunc != "" {
						f, err := getFunction(analyzer, pkg, descriptorFunc)
						if err != nil {
							return nil, err
						}
						descriptors = append(descriptors, *f)
					}
					modifierFunc := tag.Get("modifiers")
					if modifierFunc != "" {
						f, err := getFunction(analyzer, pkg, modifierFunc)
						if err != nil {
							return nil, err
						}
						modifiers = append(modifiers, *f)
					}
					style := tag.Get("style")
					if style != "" {
						apiStyle = style
					}
				}
			}
		}
	}
	if len(descriptors) <= 0 {
		return nil, fmt.Errorf("can't find descriptors from %v", b.paths)
	}
	return b.runMain(descriptors, modifiers, b.root, b.paths, apiStyle)
}

type function struct {
	Pkg   string
	Name  string
	Array bool
}

func getFunction(analyzer *api.Analyzer, pkg, name string) (*function, error) {
	f := &function{
		Pkg:  pkg,
		Name: name,
	}
	obj, err := analyzer.ObjectOf(pkg, name)
	if err != nil {
		return nil, err
	}
	key := fmt.Sprintf("%s.%s", pkg, name)
	if !obj.Exported() {
		return nil, fmt.Errorf("%s is not exported", key)
	}
	ft, ok := obj.Type().(*types.Signature)
	if !ok {
		return nil, fmt.Errorf("%s is not a function", key)
	}
	if ft.Params().Len() > 0 {
		return nil, fmt.Errorf("%s should not have parameters", key)
	}
	results := ft.Results()
	if results.Len() != 1 {
		return nil, fmt.Errorf("%s should have one result", key)
	}
	result := results.At(0)
	switch result.Type().(type) {
	case *types.Named:
		f.Array = false
	case *types.Slice:
		f.Array = true
	default:
		return nil, fmt.Errorf("%s should return an object or a slice", key)
	}
	return f, nil
}

func (b *APIBuilder) runMain(descriptors, modifiers []function, root string, paths []string, apiStyle string) (*api.Definitions, error) {
	tempDir, err := ioutil.TempDir(root, "nirvana-generated")
	if err != nil {
		return nil, err
	}
	defer func() {
		// Clean temp dir.
		err := os.RemoveAll(tempDir)
		_ = err
	}()
	data, err := b.file(descriptors, modifiers, root, paths, apiStyle)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(tempDir, "main.go")
	if err := ioutil.WriteFile(path, data, 0664); err != nil {
		return nil, err
	}
	cmd := exec.Command("go", "run", path)
	cmd.Stderr = os.Stderr
	buf := bytes.NewBuffer(nil)
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return nil, err
	}
	definitions := &api.Definitions{}
	if err := json.NewDecoder(buf).Decode(definitions); err != nil {
		return nil, err
	}
	return definitions, nil
}

func (b *APIBuilder) file(descriptors, modifiers []function, root string, paths []string, apiStyle string) ([]byte, error) {
	const tpl = `
package main

import (
	"fmt"
	"encoding/json"

	{{ range $i,$m := .modifiers }}
	m{{ $i }} "{{ $m.Pkg }}"
	{{ end }}
	{{ range $i,$d := .descriptors }}
	d{{ $i }} "{{ $d.Pkg }}"
	{{ end }}

	"github.com/caicloud/nirvana/utils/api"
	"github.com/caicloud/nirvana/log"
)

func main() {
	container, err := api.NewContainer({{ .root }}{{ range .paths }}, "{{ . }}"{{ end }})
	if err != nil {
		log.Fatal(err)
	}
	{{ range $i,$m := .modifiers }}
	container.AddModifier(m{{ $i }}.{{ $m.Name }}(){{ if $m.Array }}...{{ end }})
	{{ end }}
	{{ range $i,$d := .descriptors }}
	container.AddDescriptor(d{{ $i }}.{{ $d.Name }}(){{ if $d.Array }}...{{ end }})
	{{ end }}
	definitions, err := container.Generate({{ .apiStyle }})
	if definitions == nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(definitions)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("%s", data)
}
`
	tmpl, err := template.New("main.go").Parse(tpl)
	if err != nil {
		return nil, err
	}
	buf := bytes.NewBuffer(nil)
	if err := tmpl.Execute(buf, map[string]interface{}{
		"modifiers":   modifiers,
		"descriptors": descriptors,
		"root":        strconv.Quote(root),
		"paths":       paths,
		"apiStyle":    strconv.Quote(apiStyle),
	}); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
