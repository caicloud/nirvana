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

package api

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/packages"
)

// Analyzer analyzes go packages.
type Analyzer struct {
	root         string
	packages     map[string]*packages.Package
	files        map[string]*ast.File
	packageFiles map[string][]*ast.File
}

// NewAnalyzer creates a code analyzer.
func NewAnalyzer(root string, paths ...string) (*Analyzer, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedSyntax | packages.NeedDeps | packages.NeedImports,
		Tests: false,
		Dir:   root,
	}, paths...)
	if err != nil {
		return nil, fmt.Errorf("error loading packages: %w", err)
	}

	analyzer := &Analyzer{
		root:         root,
		packages:     make(map[string]*packages.Package),
		files:        make(map[string]*ast.File),
		packageFiles: make(map[string][]*ast.File),
	}

	for _, pkg := range pkgs {
		analyzer.packages[pkg.PkgPath] = pkg
		for _, p := range pkg.Imports {
			analyzer.packages[p.PkgPath] = p
		}
	}

	for _, pkg := range analyzer.packages {
		files := make([]*ast.File, 0, len(pkg.CompiledGoFiles))
		for i, filename := range pkg.CompiledGoFiles {
			files = append(files, pkg.Syntax[i])
			analyzer.files[filename] = pkg.Syntax[i]
		}
		analyzer.packageFiles[pkg.PkgPath] = files
	}
	return analyzer, nil
}

// Paths returns all packages' paths.
func (a *Analyzer) Paths() []string {
	paths := make([]string, 0, len(a.packages))
	for _, pkg := range a.packages {
		paths = append(paths, pkg.PkgPath)
	}
	return paths
}

// PackageComments returns comments above package keyword.
// Import package before calling this method.
func (a *Analyzer) PackageComments(path string) []*ast.CommentGroup {
	files, ok := a.packageFiles[path]
	if !ok {
		return nil
	}
	results := make([]*ast.CommentGroup, 0, len(files))
	for _, file := range files {
		for _, cg := range file.Comments {
			if cg.End() < file.Package {
				results = append(results, cg)
			}
		}
	}
	return results
}

// Comments returns immediate comments above pos.
// Import package before calling this method.
func (a *Analyzer) Comments(pkg string, pos token.Pos) *ast.CommentGroup {
	p := a.packages[pkg]
	position := p.Fset.Position(pos)
	file := a.files[position.Filename]
	for _, cg := range file.Comments {
		cgPos := p.Fset.Position(cg.End())
		if cgPos.Line == position.Line-1 {
			return cg
		}
	}
	return nil
}

// ObjectOf returns declaration object of target.
func (a *Analyzer) ObjectOf(pkg, name string) (types.Object, error) {
	// We need to rewrite analyzer with go/parser rather than go/types.
	p, ok := a.packages[pkg]
	if !ok {
		return nil, fmt.Errorf("can't find package %s", pkg)
	}
	obj := p.Types.Scope().Lookup(name)
	if obj == nil {
		return nil, fmt.Errorf("can't find declearation of %s.%s", pkg, name)
	}
	return obj, nil
}
