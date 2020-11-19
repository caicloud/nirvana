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

package builder

import (
	"github.com/caicloud/nirvana/service"
	"github.com/caicloud/nirvana/service/rest"
	"github.com/caicloud/nirvana/service/rpc"
)

const (
	// APIStyleREST represents the RESTful API style.
	APIStyleREST = "rest"
	// APIStyleRPC represents the RPC API style.
	APIStyleRPC = "rpc"
)

// New creates a service builder based on the given API style.
func New(apiStyle string) service.Builder {
	switch apiStyle {
	case APIStyleRPC:
		return rpc.NewBuilder()
	default:
		return rest.NewBuilder()
	}
}
