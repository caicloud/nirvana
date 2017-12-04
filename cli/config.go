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
	"io"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

var (
	// global viper
	v = viper.New()

	// ignoreFsEvent ingores file system event
	ignoreFsEvent = func(fsnotify.Event) {}
)

// Reset intends for testing, will reset all to default settings.
// In the public interface for the viper package so applications
// can use it in their testing as well.
func Reset() {
	v = viper.New()
}

// IsSet checks to see if the key has been set in any of the data locations.
// IsSet is case-insensitive for a key
func IsSet(key string) bool {
	return v.IsSet(key)
}

// Get can retrieve any value given the key to use.
// Get is case-insensitive for a key.
// Get has the behavior of returning the value associated with the first
// place from where it is set. Viper will check in the following order:
// override, flag, env, config file, key/value store, default
//
// Get returns an interface. For a specific value use one of the Get____ methods.
func Get(key string) interface{} {
	return v.Get(key)
}

// Set sets the value for the key in the override regiser.
// Set is case-insensitive for a key.
// Will be used instead of values obtained via
// flags, config file, ENV, default, or key/value store.
func Set(key string, value interface{}) {
	v.Set(key, value)
}

// SetConfigFile explicitly defines the path, name and extension of the config file.
// Viper will use this and not check any of the config paths.
func SetConfigFile(in string) {
	v.SetConfigFile(in)
}

// SetConfigPaths adds paths for Viper to search for the config file in.
// The given file name should not contain a extension, e.g. 'json'.
func SetConfigPaths(noExtName string, paths ...string) {
	for _, p := range paths {
		v.AddConfigPath(p)
	}
	v.SetConfigName(noExtName)
}

// SetConfigType sets the type of the configuration, e.g. "json".
func SetConfigType(in string) error {
	if !stringInSlice(in, viper.SupportedExts) {
		return viper.UnsupportedConfigError(in)
	}
	v.SetConfigType(in)
	return nil
}

// ReadConfig will read a configuration file, setting existing keys to nil if the
// key does not exist in the file.
// You should SetConfigType before read config
func ReadConfig(in io.Reader) error {
	return v.ReadConfig(in)
}

// ReadInConfig will discover and load the configuration file from disk
// and key/value stores, searching in one of the defined paths.
func ReadInConfig() error {
	return v.ReadInConfig()
}

// MergeConfig merges a new configuration with an existing config.
// You should SetConfigType before read config
func MergeConfig(in io.Reader) error {
	return v.MergeConfig(in)
}

// MergeInConfig merges a new configuration with an existing config.
func MergeInConfig() error {
	return v.MergeInConfig()
}

// WatchConfig watches the configuration file change
func WatchConfig(onChange func(in fsnotify.Event)) {
	if onChange == nil {
		// avoid nil pointer error
		onChange = ignoreFsEvent
	}
	v.OnConfigChange(onChange)
	v.WatchConfig()
}

// AllKeys returns all keys holding a value, regardless of where they are set.
// Nested keys are returned with a v.keyDelim (= ".") separator
func AllKeys() []string {
	return v.AllKeys()
}

// AllSettings merges all settings and returns them as a map[string]interface{}.
func AllSettings() map[string]interface{} {
	return v.AllSettings()
}
