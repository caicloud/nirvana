# Copyright 2018 Caicloud Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# This will force go to use the vendor files instead of using the `$GOPATH/pkg/mod`. (vendor mode)
# more info: https://github.com/golang/go/wiki/Modules#how-do-i-use-vendoring-with-modules-is-vendoring-going-away
export GOFLAGS := -mod=vendor

VERSION_PKG := github.com/caicloud/nirvana/cmd/nirvana/version
VERSION      ?= $(shell git describe --tags --always --dirty)
BRANCH       ?= $(shell git branch --show-current)
GITCOMMIT    ?= $(shell git rev-parse HEAD)
GITTREESTATE ?= $(if $(shell git status --porcelain),dirty,clean)
BUILDDATE    ?= $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")

.PHONY: test
test:
	hack/verify_boilerplate.py
	@go test -race -cover ./...

.PHONY: license
license:
	@go run ./hack/license/apache.go --go-header-file hack/boilerplate/boilerplate.go.txt

.PHONY: format
format:
	@find . ! -path "./vendor/*" -name "*.go" | xargs gofmt -s -w

.PHONY: refine
refine: format license

.PHONY: gitbook
gitbook:
	@gitbook build ./manuals ./docs

.PHONY: build-local
build-local:
	CGO_ENABLED=0 go build -trimpath $(GOFLAGS) -o bin/nirvana -ldflags "-s -w \
	-X $(VERSION_PKG).version=$(VERSION) \
	-X $(VERSION_PKG).branch=$(BRANCH) \
	-X $(VERSION_PKG).gitCommit=$(GITCOMMIT) \
	-X $(VERSION_PKG).gitTreeState=$(GITTREESTATE) \
	-X $(VERSION_PKG).buildDate=$(BUILDDATE)" ./cmd/nirvana