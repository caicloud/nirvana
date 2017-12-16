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
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

var (
	// UnderlineReplacer replace dash of underline
	UnderlineReplacer = strings.NewReplacer("-", "_")
)

var (
	automaticEnvApplied bool
	envKeyReplacer      *strings.Replacer
	envPrefix           string
)

// Flag describes a flag interface
type Flag interface {
	// IsPersistent specify whether the flag is persistent
	IsPersistent() bool
	// GetName returns the flag's name
	GetName() string
	// ApplyTo adds the flag to a given FlagSet
	ApplyTo(*pflag.FlagSet) error
}

// AutomaticEnv has Mamba check ENV variables for all.
// keys set in config, default & flags
func AutomaticEnv() {
	automaticEnvApplied = true
}

// SetEnvKeyReplacer sets the strings.Replacer on the viper object
// Useful for mapping an environmental variable to a key that does
// not match it.
func SetEnvKeyReplacer(r *strings.Replacer) {
	envKeyReplacer = r
}

// SetEnvPrefix defines a prefix that ENVIRONMENT variables will use.
// E.g. if your prefix is "spf", the env registry will look for env
// variables that start with "SPF_". Only work for automatic env
func SetEnvPrefix(in string) {
	if in != "" {
		envPrefix = in
	}
}

func mergeWithEnvPrefix(key string) string {
	if envKeyReplacer != nil {
		key = envKeyReplacer.Replace(key)
	}

	if envPrefix != "" {
		connector := "_"
		if strings.HasSuffix(envPrefix, "_") {
			connector = ""
		}
		return strings.ToUpper(envPrefix + connector + key)
	}

	return strings.ToUpper(key)
}

// getEnv tries to get envKey from env. otherwise returns defValue.
// you must convert the return value to the type you want.
//
// if env key is "", and AutomaticEnv is set, mamba will try to generate
// env key by merging name with envPrefix.
// finally, if the key is "" or key is not set in env, returns the defValue.
func getEnv(name, envKey string, defValue interface{}) (string, interface{}) {

	if envKey == "" && automaticEnvApplied {
		envKey = mergeWithEnvPrefix(name)
	}

	if envKey == "" {
		return "", defValue
	}

	e, ok := os.LookupEnv(envKey)
	if ok {
		return envKey, e
	}

	return envKey, defValue

}

func appendEnvToUsage(usage, key string) string {
	if key == "" {
		return usage
	}

	if usage == "" {
		return fmt.Sprintf("(env $%v)", key)
	}

	return fmt.Sprintf("%v (env $%v)", usage, key)

}
