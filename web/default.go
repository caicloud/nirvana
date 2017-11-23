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

package web

import "sync/atomic"

// It's used as a lock in RegisterDefaultEnvironment().
var registered = int32(0)

// RegisterDefaultEnvironment registers default plugins.
func RegisterDefaultEnvironment() error {
	if !atomic.CompareAndSwapInt32(&registered, 0, 1) {
		return nil
	}
	if err := registerJSONConsumerAndProducer(); err != nil {
		return err
	}
	if err := registerXMLConsumerAndProducer(); err != nil {
		return err
	}
	if err := registerContextPrefab(); err != nil {
		return err
	}
	if err := registerDefaultParameterGenerators(); err != nil {
		return err
	}
	if err := registerDefaultTypeHandlers(); err != nil {
		return err
	}
	return nil
}

func registerJSONConsumerAndProducer() error {
	// Register JSON
	err := RegisterConsumer(&JSONSerializer{})
	if err != nil {
		return err
	}
	err = RegisterProducer(&JSONSerializer{})
	if err != nil {
		return err
	}
	return nil
}

func registerXMLConsumerAndProducer() error {
	// Register XML
	err := RegisterConsumer(&XMLSerializer{})
	if err != nil {
		return err
	}
	err = RegisterProducer(&XMLSerializer{})
	if err != nil {
		return err
	}
	return nil
}

func registerContextPrefab() error {
	return RegisterPrefab(&ContextPrefab{})
}

func registerDefaultParameterGenerators() error {
	if err := RegisterParameterGenerator(&PathParameterGenerator{}); err != nil {
		return err
	}
	if err := RegisterParameterGenerator(&QueryParameterGenerator{}); err != nil {
		return err
	}
	if err := RegisterParameterGenerator(&HeaderParameterGenerator{}); err != nil {
		return err
	}
	if err := RegisterParameterGenerator(&FormParameterGenerator{}); err != nil {
		return err
	}
	if err := RegisterParameterGenerator(&FileParameterGenerator{}); err != nil {
		return err
	}
	if err := RegisterParameterGenerator(&BodyParameterGenerator{}); err != nil {
		return err
	}
	if err := RegisterParameterGenerator(&PrefabParameterGenerator{}); err != nil {
		return err
	}
	if err := RegisterParameterGenerator(&AutoParameterGenerator{}); err != nil {
		return err
	}
	return nil
}

func registerDefaultTypeHandlers() error {
	if err := RegisterTypeHandler(&MetaTypeHandler{}); err != nil {
		return err
	}
	if err := RegisterTypeHandler(&DataTypeHandler{}); err != nil {
		return err
	}
	if err := RegisterTypeHandler(&ErrorTypeHandler{}); err != nil {
		return err
	}
	return nil
}
