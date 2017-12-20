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

package cli

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// make directories for testing
func initDirs(t *testing.T) (string, string, func()) {

	var (
		testDirs = []string{`a a`, `b`, `c\c`, `D_`}
		config   = `improbable`
	)

	root, err := ioutil.TempDir("", "")

	cleanup := true
	defer func() {
		if cleanup {
			os.Chdir("..")
			os.RemoveAll(root)
		}
	}()

	assert.Nil(t, err)

	err = os.Chdir(root)
	assert.Nil(t, err)

	for _, dir := range testDirs {
		err = os.Mkdir(dir, 0750)
		assert.Nil(t, err)

		err = ioutil.WriteFile(
			path.Join(dir, config+".toml"),
			[]byte("key = \"value is "+dir+"\"\n"),
			0640)
		assert.Nil(t, err)
	}

	cleanup = false
	return root, config, func() {
		os.Chdir("..")
		os.RemoveAll(root)
	}
}
func TestIsSet(t *testing.T) {
	Reset()
	tests := []struct {
		name string
		key  string
		set  bool
		want bool
	}{
		{"", "test", true, true},
		{"", "test2", false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.set {
				v.Set(tt.key, "aaa")
			}
			assert.Equal(t, tt.want, IsSet(tt.key))
		})
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
		want  interface{}
	}{
		{"", "test", true, true},
		{"", "test2", 1, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Set(tt.key, tt.value)
			assert.Equal(t, tt.want, Get(tt.key))
		})
	}
}

func TestSetConfigFile(t *testing.T) {
	type args struct {
		in string
	}
	var tests []struct {
		name string
		args args
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetConfigFile(tt.args.in)
		})
	}
}

func TestReadConfig(t *testing.T) {
	tests := []struct {
		name    string
		content string
		key     string
		value   interface{}
		wantErr bool
	}{
		{"json", `{"key": "json"}`, "key", "json", false},
		{"toml", `key = "toml"`, "key", "toml", false},
		{"yaml", `key: yaml`, "key", "yaml", false},
		{"hcl", `key = "hcl"`, "key", "hcl", false},
		{"props", `key: props`, "key", "props", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetConfigType(tt.name)
			err := ReadConfig(strings.NewReader(tt.content))
			t.Log(tt.content)
			got := Get(tt.key)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, got, tt.value)
		})
	}
}

func TestReadInConfig(t *testing.T) {
	Reset()
	root, config, cleanup := initDirs(t)
	defer cleanup()

	entries, _ := ioutil.ReadDir(root)
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			paths = append(paths, e.Name())
		}
	}
	SetConfigPaths(config, paths...)
	v.SetDefault("key", "default")

	err := ReadInConfig()
	assert.Nil(t, err)
	assert.Equal(t, "value is "+path.Base(paths[0]), v.GetString("key"))

	SetConfigFile(filepath.Join(paths[1], config+".toml"))
	err = ReadInConfig()
	assert.Nil(t, err)
	assert.Equal(t, "value is "+path.Base(paths[1]), v.GetString("key"))

}

func TestMergeConfig(t *testing.T) {
	Reset()
	tests := []struct {
		name    string
		content string
		key     string
		value   interface{}
		wantErr bool
	}{
		{"json", `{"key": "json"}`, "key", "json", false},
		{"toml", `key = "toml"`, "key", "toml", false},
		{"yaml", `key: yaml`, "key", "yaml", false},
		{"hcl", `key = "hcl"`, "key", "hcl", false},
		{"props", `key: props`, "key", "props", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetConfigType(tt.name)
			err := MergeConfig(strings.NewReader(tt.content))
			t.Log(tt.content)
			got := Get(tt.key)
			if tt.wantErr {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
			assert.Equal(t, got, tt.value)
		})
	}
}

func TestMergeInConfig(t *testing.T) {
	Reset()
	root, config, cleanup := initDirs(t)
	defer cleanup()

	entries, _ := ioutil.ReadDir(root)
	var paths []string
	for _, e := range entries {
		if e.IsDir() {
			paths = append(paths, e.Name())
		}
	}
	SetConfigPaths(config, paths...)
	v.SetDefault("key", "default")

	err := MergeInConfig()
	assert.Nil(t, err)
	assert.Equal(t, "value is "+path.Base(paths[0]), v.GetString("key"))

	SetConfigFile(filepath.Join(paths[1], config+".toml"))
	err = ReadInConfig()
	assert.Nil(t, err)
	assert.Equal(t, "value is "+path.Base(paths[1]), v.GetString("key"))
}

func TestAllKeys(t *testing.T) {
	Reset()
	Set("k1", "value")
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"k2": "v", "k": {"k3": 1}}`))
	want := sort.StringSlice{"k1", "k2", "k.k3"}
	want.Sort()
	allKeys := sort.StringSlice(AllKeys())
	allKeys.Sort()
	assert.Equal(t, want, allKeys)
}

func TestAllSettings(t *testing.T) {
	Reset()
	Set("k1", "value")
	SetConfigType("json")
	ReadConfig(strings.NewReader(`{"k2": "v", "k": {"k3": 1}}`))

	assert.Equal(t, map[string]interface{}{"k": map[string]interface{}{"k3": 1.0}, "k1": "value", "k2": "v"}, AllSettings())
}
