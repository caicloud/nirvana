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
