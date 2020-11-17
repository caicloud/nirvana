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

// RESTfulDescriptor describes a descriptor for RESTfulAPI definitions.
type RESTfulDescriptor struct {
	// Path is the url path. It will inherit parent's path.
	//
	// If parent path is "/api/v1", current is "/some",
	// It means current definitions handles "/api/v1/some".
	Path string
	// Consumes indicates content types that current definitions
	// and child definitions can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates content types that current definitions
	// and child definitions can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Tags indicates tags of current definitions and child definitions.
	// It will override parent descriptor's tags.
	Tags []string
	// Middlewares contains path middlewares.
	Middlewares []Middleware
	// Definitions contains definitions for current path.
	Definitions []Definition
	// Children is used to place sub-descriptors.
	Children []RESTfulDescriptor
	// Description describes the usage of the path.
	Description string
}

func (d RESTfulDescriptor) GetPath() string {
	return d.Path
}

func (d RESTfulDescriptor) GetConsumes() []string {
	return d.Consumes
}

func (d RESTfulDescriptor) GetProduces() []string {
	return d.Produces
}

func (d RESTfulDescriptor) GetTags() []string {
	return d.Tags
}

func (d RESTfulDescriptor) GetMiddlewares() []Middleware {
	return d.Middlewares
}

func (d RESTfulDescriptor) GetDefinitions() []Definition {
	return d.Definitions
}

func (d RESTfulDescriptor) GetChildren() []Descriptor {
	childrenI := make([]Descriptor, 0, len(d.Children))
	for _, c := range d.Children {
		childrenI = append(childrenI, c)
	}
	return childrenI
}

func (d RESTfulDescriptor) GetDescription() string {
	return d.Description
}
