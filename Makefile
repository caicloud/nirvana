
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
