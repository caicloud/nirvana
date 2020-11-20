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
	// DefinitionNoMethod represents no http method error.
	DefinitionNoMethod = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoMethod", "no http method in [${method}]${path}")
	// DefinitionNoConsumes represents no content type to consume.
	DefinitionNoConsumes = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoConsumes", "no content type to consume in [${method}]${path}")
	// DefinitionNoProduces represents no content type to produce.
	DefinitionNoProduces = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoProduces", "no content type to produce in [${method}]${path}")
	// DefinitionNoErrorProduces represents no content type to produce error.
	DefinitionNoErrorProduces = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoErrorProduces", "no content type to produce error in [${method}]${path}")
	// DefinitionNoFunction represents no function error.
	DefinitionNoFunction = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoFunction", "no function in [${method}]${path}")
	// DefinitionInvalidFunctionType represents invalid function type.
	DefinitionInvalidFunctionType = errors.InternalServerError.Build("Nirvana:Service:DefinitionInvalidFunctionType", "${type} is not function in [${method}]${path}")
	// DefinitionNoConsumer represents no consumer error.
	DefinitionNoConsumer = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoConsumer", "no consumer for content type ${type} in [${method}]${path}")
	// DefinitionNoProducer represents no producer error.
	DefinitionNoProducer = errors.InternalServerError.Build("Nirvana:Service:DefinitionNoProducer", "no producer for content type ${type} in [${method}]${path}")
	// DefinitionConflict represents conflict error.
	DefinitionConflict = errors.InternalServerError.Build("Nirvana:Service:DefinitionConflict", "consumer-producer pair ${key}:${value} conflicts in [http.${method}]${path}")
	// DefinitionUnmatchedParameters represents parameters unmatch.
	DefinitionUnmatchedParameters = errors.InternalServerError.Build(
		"Nirvana:Service:DefinitionUnmatchedParameters",
		"function ${function} has ${count} parameters but want ${desired} in ${path}, you can define it with descriptor->definition[]->parameters[]",
	)
	// DefinitionUnmatchedResults represents results unmatch.
	DefinitionUnmatchedResults = errors.InternalServerError.Build(
		"Nirvana:Service:DefinitionUnmatchedResults",
		"function ${function} has ${count} results but want ${desired} in ${path}, you can define it with descriptor->definition[]->results[]",
	)
	// NoDestinationHandler represents no DestinationHandler error.
	NoDestinationHandler = errors.InternalServerError.Build(
		"Nirvana:Service:NoDestinationHandler",
		"no destination handler for destination ${destination}, you can define it with descriptor->definition[]->results[]->destination",
	)
	// InvalidParameter represents invalid parameter error.
	InvalidParameter = errors.InternalServerError.Build("Nirvana:Service:InvalidParameter", "can't validate ${order} parameter of function ${name}: ${err}")
	// InvalidResult represents invalid parameter error.
	InvalidResult = errors.InternalServerError.Build("Nirvana:Service:InvalidResult", "can't validate ${order} result of function ${name}: ${err}")
	// InvalidOperatorsForParameter represents invalid operators error.
	InvalidOperatorsForParameter = errors.InternalServerError.Build("Nirvana:Service:InvalidOperatorsForParameter", "can't validate operators for ${order} parameter of function ${name}: ${err}")
	// InvalidOperatorsForResult represents invalid operators error.
	InvalidOperatorsForResult = errors.InternalServerError.Build("Nirvana:Service:InvalidOperatorsForResult", "can't validate operators for ${order} result of function ${name}: ${err}")
)

var (
	requiredField          = errors.InternalServerError.Build("Nirvana:Service:RequiredField", "required field ${field} in ${source} but got empty")
	invalidOperatorInType  = errors.InternalServerError.Build("Nirvana:Service:invalidOperatorInType", "the type ${type} is not compatible to the in type of the ${index} operator")
	invalidOperatorOutType = errors.InternalServerError.Build("Nirvana:Service:invalidOperatorOutType", "the out type of the ${index} operator is not compatible to the type ${type}")
)
