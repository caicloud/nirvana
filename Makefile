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

PKGS := $(shell go list ./... | grep -v /vendor | grep -v /tests)

.PHONY: test
test:
	@go test -cover $(PKGS)
	hack/verify_boilerplate.py
	hack/verify-govet.sh
	hack/verify-gofmt.sh

.PHONY: flag-gen
flag-gen:
	go run ./hack/flag-gen/main.go -i github.com/caicloud/nirvana/hack/flag-gen/types \
	  -p github.com/caicloud/nirvana/cli \
	  -v 5
	@for generated in $(shell ls cli | grep generated); do \
		echo "run goimports on cli/$${generated}"; \
		goimports -w cli/$${generated}; \
	done


.PHONY: license
license:
	go run ./hack/license/apache.go --go-header-file hack/boilerplate/boilerplate.go.txt --logtostderr
