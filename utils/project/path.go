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

package project

import (
	"fmt"
	"go/build"
	"os"
	"path/filepath"
	"strings"
)

// goPaths contains all valid go paths.
var goPaths []string
var goSrcPaths []string

func init() {
	paths := strings.Split(build.Default.GOPATH, string(os.PathListSeparator))
	for _, path := range paths {
		path, err := filepath.Abs(path)
		if err != nil {
			continue
		}
		info, err := os.Stat(path)
		if err != nil || !info.IsDir() {
			// Ignore go path which is nonexistent or non-readable.
			continue
		}
		goPaths = append(goPaths, path)
		goSrcPaths = append(goSrcPaths, srcPath(path))
	}
}

// srcPath returns the path of src in go path.
func srcPath(goPath string) string {
	return filepath.Join(goPath, "src")
}

// GoPath finds the go path and absolute path for specified directory.
func GoPath(directory string) (goPath string, absPath string, err error) {
	absPath, err = filepath.Abs(directory)
	if err != nil {
		return "", "", err
	}
	for i, path := range goSrcPaths {
		if strings.HasPrefix(absPath, path) {
			return goPaths[i], absPath, nil
		}
	}
	return "", "", fmt.Errorf("%s is not in GOPATH", absPath)
}

// PackageForPath gets package path for a path.
func PackageForPath(directory string) (string, error) {
	absPath, err := filepath.Abs(directory)
	if err != nil {
		return "", err
	}
	for i, path := range goSrcPaths {
		if strings.HasPrefix(absPath, path) {
			// in GOPATH
			return filepath.ToSlash(strings.Trim(absPath[len(srcPath(goPaths[i])):], string(os.PathSeparator))), nil
		}
	}

	// not in GOPATH, use the directory directly
	return filepath.Base(directory), nil
}

// Subdirectories walkthroughs all subdirectories. The results contains itself.
// The path will be ignored if:
// 1. The path is non-existent
// 2. Go module is disabled and the path is not in GOPATH
// 3. Go modules is enabled and the path doesn't belong to a go module
func Subdirectories(vendor bool, paths ...string) []string {
	walked := map[string]bool{}
	goDir := map[string]bool{}
	for _, path := range paths {
		var (
			absPath string
			err     error
		)
		useModule := os.Getenv("GO111MODULE")
		if useModule == "off" {
			_, absPath, err = GoPath(path)
		} else {
			root := findModuleRoot(path)
			if len(root) > 0 {
				absPath, err = filepath.Abs(path)
			} else {
				if useModule == "on" {
					continue
				}
				_, absPath, err = GoPath(path)
			}
		}
		if err != nil {
			// Ignore the err and go next.
			continue
		}
		_ = filepath.Walk(absPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return filepath.SkipDir
			}
			if info.IsDir() {
				if !vendor && info.Name() == "vendor" {
					return filepath.SkipDir
				}
				if walked[path] {
					return filepath.SkipDir
				}
				walked[path] = true
				return nil
			}
			if strings.HasSuffix(path, ".go") {
				dir := filepath.Dir(path)
				goDir[dir] = true
			}
			return nil
		})
	}
	results := make([]string, 0, len(goDir))
	for path := range goDir {
		results = append(results, path)
	}
	return results
}

// findModuleRoot returns the path of main module.
func findModuleRoot(dir string) string {
	if dir == "" {
		return ""
	}
	dir = filepath.Clean(dir)

	// Look for enclosing go.mod.
	for {
		if fi, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil && !fi.IsDir() {
			return dir
		}
		d := filepath.Dir(dir)
		if d == dir {
			break
		}
		dir = d
	}
	return ""
}
