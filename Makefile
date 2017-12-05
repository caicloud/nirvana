
PKGS := $(shell go list ./... | grep -v /vendor | grep -v /tests)

.PHONY: test
test:
	@go test -cover $(PKGS)

.PHONY: flag-gen 
flag-gen: 
	go run ./cmd/flag-gen/main.go -i github.com/caicloud/nirvana/cmd/flag-gen/types \
	  -p github.com/caicloud/nirvana/cli \
	  -v 5
	@for generated in $(shell ls cli | grep generated); do \
		echo "run goimports on cli/$${generated}"; \
		goimports -w cli/$${generated}; \
	done


.PHONY: license
license: 
	go run ./cmd/license/apache.go --go-header-file boilerplate/boilerplate.go.txt --logtostderr

