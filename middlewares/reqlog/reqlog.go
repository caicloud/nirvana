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

package reqlog

import (
	"context"
	"time"

	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
	"github.com/caicloud/nirvana/service"
)

// Default returns a reqlog middleware that uses the default Nirvana logger.
func Default() definition.Middleware {
	return func(ctx context.Context, chain definition.Chain) error {
		start := time.Now()
		httpCtx := service.HTTPContextFrom(ctx)

		err := chain.Continue(ctx)

		request := httpCtx.Request()
		response := httpCtx.ResponseWriter()
		log.Infoln(
			request.Method,
			response.StatusCode(),
			response.ContentLength(),
			time.Since(start).String(),
			request.URL.String(),
		)

		return err
	}
}

// Custom returns a reqlog middleware with a custom logger and designated logging level.
func Custom(logger log.Logger, level log.Level) definition.Middleware {
	return func(ctx context.Context, chain definition.Chain) error {
		start := time.Now()
		httpCtx := service.HTTPContextFrom(ctx)

		err := chain.Continue(ctx)

		request := httpCtx.Request()
		response := httpCtx.ResponseWriter()
		logger.V(level).Infoln(
			request.Method,
			response.StatusCode(),
			response.ContentLength(),
			time.Since(start).String(),
			request.URL.String(),
		)

		return err
	}
}
