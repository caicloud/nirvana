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

package router

import "github.com/caicloud/nirvana/errors"

// Router `Match` method only can return errors from these error factory:
// routerNotFound, noInspector, noExecutor.

var (
	// routerNotFound means there no router matches path.
	routerNotFound = errors.NotFound.Build("Nirvana:Router:routerNotFound", "can't find router")
	// noInspector means there is no inspector in router.
	noInspector = errors.NotFound.Build("Nirvana:Router:noInspector", "no inspector to generate executor")
	// noExecutor means can't pack middlewares for nil executor.
	noExecutor = errors.NotFound.Build("Nirvana:Router:noExecutor", "no executor to pack middlewares")
	// conflictInspectors can build errors for failure of merging routers.
	// If attempts to merge two router and they all have inspector, an error
	// should be returned. A merged router can't have two inspectors.
	conflictInspectors = errors.Conflict.Build("Nirvana:Router:conflictInspectors", "can't merge two routers that all have inspector")
	// emptyRouterTarget means a router node has an invalid empty target.
	emptyRouterTarget = errors.UnprocessableEntity.Build("Nirvana:Router:emptyRouterTarget", "router ${kind} has no target")
	// unknownRouterType means a node type is unprocessable.
	unknownRouterType = errors.UnprocessableEntity.Build("Nirvana:Router:unknownRouterType", "router ${kind} has unknown type ${type}")
	// unmatchedRouterKey means a router's key is not matched with another.
	unmatchedRouterKey = errors.UnprocessableEntity.Build("Nirvana:Router:unmatchedRouterKey", "router key ${keyA} is not matched with ${keyB}")
	// unmatchedRouterRegexp means a router's regexp is not matched with another.
	unmatchedRouterRegexp = errors.UnprocessableEntity.Build("Nirvana:Router:unmatchedRouterRegexp", "router regexp ${regexpA} is not matched with ${regexpA}")
	// noCommonPrefix means two routers have no common prefix.
	noCommonPrefix = errors.UnprocessableEntity.Build("Nirvana:Router:noCommonPrefix", "there is no common prefix for the two routers")
	// invalidPath means router path is invalid.
	invalidPath = errors.UnprocessableEntity.Build("Nirvana:Router:invalidPath", "invalid path")
	// invalidParentRouter means router node has no method to add child routers.
	invalidParentRouter = errors.UnprocessableEntity.Build("Nirvana:Router:invalidParentRouter", "router ${type} has no method to add children")
	// unmatchedSegmentKeys means segment has unmatched keys.
	unmatchedSegmentKeys = errors.UnprocessableEntity.Build("Nirvana:Router:unmatchedSegmentKeys", "segment ${value} has unmatched keys")
	// unknownSegment means can't recognize segment.
	unknownSegment = errors.UnprocessableEntity.Build("Nirvana:Router:unknownSegment", "unknown segment ${value}")
	// unmatchedPathBrace means path has unmatched brace.
	unmatchedPathBrace = errors.UnprocessableEntity.Build("Nirvana:Router:unmatchedPathBrace", "unmatched braces")
	// invalidPathKey means path key must be the last element.
	invalidPathKey = errors.UnprocessableEntity.Build("Nirvana:Router:invalidPathKey", "key ${key} should be last element in the path")
	// invalidRegexp means regexp is not notmative.
	invalidRegexp = errors.UnprocessableEntity.Build("Nirvana:Router:invalidRegexp", "regexp ${regexp} does not have normative format")
)
