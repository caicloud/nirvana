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

package golang

import (
	"bytes"
	"fmt"
	"go/format"
	"path"
	"strings"
	"text/template"

	"github.com/caicloud/nirvana/utils/api"
	"github.com/caicloud/nirvana/utils/generators/utils"
	"github.com/caicloud/nirvana/utils/project"
)

// Generator is for generating golang client.
type Generator struct {
	config  *project.Config
	apis    *api.Definitions
	rest    string
	pkg     string
	rootPkg string
}

// NewGenerator creates a golang client generator.
func NewGenerator(
	config *project.Config,
	apis *api.Definitions,
	rest string,
	pkg string,
	rootPkg string,
) *Generator {
	return &Generator{
		config:  config,
		apis:    apis,
		rest:    rest,
		pkg:     pkg,
		rootPkg: rootPkg,
	}
}

// Generate generate files
func (g *Generator) Generate() (map[string][]byte, error) {
	definitions, err := utils.SplitDefinitions(g.apis, g.config)
	if err != nil {
		return nil, err
	}
	codes := make(map[string][]byte)
	versions := make([]utils.Version, 0, len(definitions))
	for _, d := range definitions {
		versions = append(versions, d.Version)
		helper, err := newHelper(g.rootPkg, d.Defs)
		if err != nil {
			return nil, err
		}
		// all lower case string
		packageName := d.Version.Module + d.Version.Name
		types, imports := helper.Types()
		typeCodes, err := g.typeCodes(packageName, types, imports)
		if err != nil {
			return nil, err
		}
		functions, imports := helper.Functions()
		functionCodes, err := g.functionCodes(packageName, functions, imports)
		if err != nil {
			return nil, err
		}
		codes[packageName+"/types"] = typeCodes
		codes[packageName+"/client"] = functionCodes
	}
	client, err := g.aggregationClientCode(versions)
	if err != nil {
		return nil, err
	}
	codes["client"] = client
	return codes, nil
}

func (g *Generator) typeCodes(version string, types []Type, imports []string) ([]byte, error) {
	data := bytes.NewBufferString(fmt.Sprintf("package %s\n", version))
	writeln := func(str string) {
		_, err := fmt.Fprintln(data, str)
		// Ignore this error.
		_ = err
	}

	if len(imports) > 0 {
		writeln("import (")
		for _, pkg := range imports {
			writeln(pkg)
		}
		writeln(")")
	}

	for _, typ := range types {
		writeln("")
		writeln(string(typ.Generate()))
	}
	return format.Source(data.Bytes())
}

func (g *Generator) functionCodes(version string, functions []function, imports []string) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	template, err := template.New("codes").Parse(`
package {{ .Version }}

import (
	"context"

	{{- range .Imports }}
	{{.}}
	{{- end }}

	rest "{{ .Rest }}"
)

// Interface describes {{ .Version }} client.
type Interface interface {
{{- range .Functions }}
{{ .Comments -}}
	{{ .Name }}(ctx context.Context{{- if eq .Method "Any" }}, method string, responseCode int{{- end }}{{ range .Parameters }},{{ .ProposedName }} {{ .Typ }}{{- end }}) (
	{{- range .Results }}{{ .ProposedName }} {{ .Typ }}, {{ end }}err error)
{{- end }}
}

// Client for version {{ .Version }}.
type Client struct {
	rest *rest.Client
}

// NewClient creates a new client.
func NewClient(cfg *rest.Config) (*Client, error) {
	client, err := rest.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	return &Client{client}, nil
}

// MustNewClient creates a new client or panic if an error occurs.
func MustNewClient(cfg *rest.Config) *Client {
	client, err := NewClient(cfg)
	if err != nil {
		panic(err)
	}
	return client
}

{{ range .Functions }}
{{ .Comments -}}
func (c *Client) {{ .Name }}(ctx context.Context{{- if eq .Method "Any" }}, method string, responseCode int{{- end }}{{ range .Parameters }},{{ .ProposedName }} {{ .Typ }}{{- end }}) (
	{{- range .Results }}{{ .ProposedName }} {{ .Typ }}, {{ end }}err error) {
	{{- range .Results }}
	{{- if ne .Creator "" }}
	{{ .ProposedName }} = {{ .Creator }}
    {{- end }}
    {{- end }}
	err = c.rest.Request({{- if eq .Method "Any" }}method, responseCode{{- else }}"{{ .Method }}", {{ .Code }}{{- end }}, "{{ .Path }}").
	{{ range .Parameters }}
	{{ $param := .ProposedName }}
	{{ if not .Extensions }}
	{{ .Source }}("{{ .Name }}", {{ $param }}).
	{{ end }}
	{{ range .Extensions }}
	{{ .Source }}("{{ .Name }}", {{ $param }}.{{ .Key }}).
	{{ end }}
    {{ end }}

	{{ range .Results }}
	{{- if ne .Creator "" }}
	{{ .Destination }}({{ .ProposedName }}).
    {{- else }}
	{{ .Destination }}(&{{ .ProposedName }}).
    {{- end }}
    {{ end }}
	Do(ctx)
	return 
}
{{ end }}
		`)
	if err != nil {
		return nil, err
	}
	err = template.Execute(buf, map[string]interface{}{
		"Version":   version,
		"Rest":      g.rest,
		"Functions": functions,
		"Imports":   imports,
	})
	if err != nil {
		return nil, err
	}
	return format.Source(buf.Bytes())
}

type versionedPackage struct {
	Alias    string
	Version  string
	Path     string
	Function string
}

func (g *Generator) aggregationClientCode(versions []utils.Version) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	template, err := template.New("codes").Parse(`
package {{ .PackageName }}

import (
	{{ range .Pakcages }}
	"{{ .Path }}"
	{{- end }}

	rest "{{ .Rest }}"
)

// Interface describes a versioned client.
type Interface interface {
{{- range .Pakcages }}
	// {{ .Function }} returns {{ .Alias }} client.
	{{ .Function }}() {{ .Alias }}.Interface
{{- end }}
}

// Client contains versioned clients.
type Client struct {
	{{ range .Pakcages }}
	{{ .Version }} *{{ .Alias }}.Client
	{{- end }}
}

// NewClient creates a new client.
func NewClient(cfg *rest.Config) (Interface, error) {
	c := &Client{}
	var err error
	{{ range .Pakcages }}
	c.{{ .Version }}, err =  {{ .Alias }}.NewClient(cfg)
	if err != nil {
		return nil, err
	}
	{{ end -}}
	return c, nil
}

// MustNewClient creates a new client or panic if an error occurs.
func MustNewClient(cfg *rest.Config) Interface {
	return &Client{
	{{- range .Pakcages }}
	{{ .Version }}: {{ .Alias }}.MustNewClient(cfg),
	{{- end }}
	}
}

{{ range .Pakcages }}
// {{ .Function }} returns a versioned client.
func (c *Client) {{ .Function }}() {{ .Alias }}.Interface {
	return c.{{ .Version }}
}
{{ end }}
		`)
	if err != nil {
		return nil, err
	}
	packages := make([]versionedPackage, 0, len(versions))
	for _, version := range versions {
		alias := version.Module + version.Name
		var v string
		if version.Module != "" {
			v = version.Module + strings.Title(version.Name)
		} else {
			v = version.Name
		}
		packages = append(packages, versionedPackage{
			Alias:    alias,
			Version:  v,
			Path:     path.Join(g.pkg, alias),
			Function: strings.Title(version.Module) + strings.Title(version.Name),
		})
	}
	err = template.Execute(buf, map[string]interface{}{
		"PackageName": path.Base(g.pkg),
		"Pakcages":    packages,
		"Rest":        g.rest,
	})
	if err != nil {
		return nil, err
	}
	return format.Source(buf.Bytes())
}
