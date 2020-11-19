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

package rest

import (
	"github.com/caicloud/nirvana/errors"
)

var (
	noExecutorForMethod      = errors.MethodNotAllowed.Build("Nirvana:Service:NoExecutorForMethod", "method not allowed")
	noExecutorForContentType = errors.UnsupportedMediaType.Build("Nirvana:Service:NoExecutorForContentType", "unsupported media type")
	noExecutorToProduce      = errors.NotAcceptable.Build("Nirvana:Service:NoExecutorToProduce", "not acceptable")
	noRouter                 = errors.InternalServerError.Build("Nirvana:Service:NoRouter", "no router to build service")
)
