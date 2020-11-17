/*
Copyright 2020 Caicloud Authors

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

package definition

import (
	"fmt"
	"net/http"
)

const (
	ActionName  = "Action"
	VersionName = "Version"
)

type RPCDefinition struct {
	Definition

	// Version defines the version this API belongs to.
	// Need to use time format, eg: 2020-10-10
	Version string
	// Name defines the Action name.
	Name string
}

func (rd *RPCDefinition) merge(d RPCDescriptor) *RPCDefinition {
	if rd.Version == "" {
		rd.Version = d.Version
	}
	if rd.Consumes == nil {
		rd.Consumes = d.Consumes
	}
	if rd.Produces == nil {
		rd.Produces = d.Produces
	}
	if rd.Tags == nil {
		rd.Tags = d.Tags
	}
	if rd.PreFunctions == nil {
		rd.PreFunctions = d.PreFunctions
	}
	return rd
}

// RPCDescriptor describes a descriptor for API definition in RPC style.
type RPCDescriptor struct {
	// Path is the url path.
	Path string
	// Version defines the version this API belongs to.
	// Need to use time format, eg: 2020-10-10
	Version string
	// Description describes the usage of the Descriptor.
	Description string
	// PreFunctions describe template for PreFunctions of RPCDefinitions
	PreFunctions []Middleware
	// Tags describe template for Tags of RPCDefinitions
	Tags []string
	// Consumes describe template for Consumes of RPCDefinitions
	Consumes []string
	// Produces describe template for Produces of RPCDefinitions
	Produces []string
	// Definitions contains definitions for current descriptor.
	RPCDefinitions []RPCDefinition
}

func (d RPCDescriptor) GetPath() string {
	return d.Path
}

func (d RPCDescriptor) GetConsumes() []string {
	return d.Consumes
}

func (d RPCDescriptor) GetProduces() []string {
	return d.Produces
}

func (d RPCDescriptor) GetTags() []string {
	return d.Tags
}

func (d RPCDescriptor) GetMiddlewares() []Middleware {
	return nil
}

func (d RPCDescriptor) GetDefinitions() []Definition {
	defs := make([]Definition, 0, len(d.RPCDefinitions))
	for _, rpcDef := range d.RPCDefinitions {
		rpcDef := rpcDef.merge(d)
		rpcDef.Condition = Condition{
			Satisfied: func(req *http.Request) bool {
				return req.URL.Query().Get(ActionName) == rpcDef.Name && req.URL.Query().Get(VersionName) == rpcDef.Version
			},
			UniqID: func() string {
				return rpcDef.Name + rpcDef.Version
			},
			Description: func() string {
				return fmt.Sprintf("?Action=%s&Version=%s", rpcDef.Name, rpcDef.Version)
			},
		}
		defs = append(defs, rpcDef.Definition)
	}
	return defs
}

func (d RPCDescriptor) GetChildren() []Descriptor {
	return nil
}

func (d RPCDescriptor) GetDescription() string {
	return d.Description
}
