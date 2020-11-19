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

package executor

import (
	"github.com/caicloud/nirvana/errors"
)

var (
	DefinitionNoMethod            = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoMethod", "no http method in [${method}]${path}")
	DefinitionNoConsumes          = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoConsumes", "no content type to consume in [${method}]${path}")
	DefinitionNoProduces          = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoProduces", "no content type to produce in [${method}]${path}")
	DefinitionNoErrorProduces     = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoErrorProduces", "no content type to produce error in [${method}]${path}")
	DefinitionNoFunction          = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoFunction", "no function in [${method}]${path}")
	DefinitionInvalidFunctionType = errors.InternalServerError.Build("Nirvana:Service:DefinitionInvalidFunctionType",
		"${type} is not function in [${method}]${path}")
	DefinitionNoConsumer = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoConsumer",
		"no consumer for content type ${type} in [${method}]${path}")
	DefinitionNoProducer = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoProducer",
		"no producer for content type ${type} in [${method}]${path}")
	DefinitionConflict = errors.InternalServerError.Build("Nirvana:Service:DefinitionConflict",
		"consumer-producer pair ${key}:${value} conflicts in [http.${method}]${path}")
	DefinitionUnmatchedParameters = errors.InternalServerError.Build("Nirvana:Service:DefinitionUnmatchedParameters",
		"function ${function} has ${count} parameters but want ${desired} in ${path}, "+
			"you can define it with descriptor->definition[]->parameters[]")
	DefinitionUnmatchedResults = errors.InternalServerError.Build("Nirvana:Service:DefinitionUnmatchedResults",
		"function ${function} has ${count} results but want ${desired} in ${path}, "+
			"you can define it with descriptor->definition[]->results[]")
	NoDestinationHandler = errors.InternalServerError.Build("Nirvana:Service:NoDestinationHandler", "no destination handler for destination ${destination}, "+
		"you can define it with descriptor->definition[]->results[]->destination")
	InvalidOperatorInType = errors.InternalServerError.Build("Nirvana:Service:InvalidOperatorInType",
		"the type ${type} is not compatible to the in type of the ${index} operator")
	InvalidOperatorOutType = errors.InternalServerError.Build("Nirvana:Service:InvalidOperatorOutType",
		"the out type of the ${index} operator is not compatible to the type ${type}")
)

var (
	requiredField = errors.InternalServerError.Build("Nirvana:Service:RequiredField", "required field ${field} in ${source} but got empty")
)
