# How to run example

## Run by hack/run.sh

```
./examples/openapi/hack/run.sh

```

`swagger.json` will be output as `./examples/openapi/swagger.json`

## Run manually

workspace: github.com/caicloud/nirvana

### Generate definition

```
go run ./cmd/openapi-gen/main.go \
    -i github.com/caicloud/nirvana/examples/openapi/pkg/api/v1 \
    -p github.com/caicloud/nirvana/examples/openapi/api
```

A file named `openapi_generated.go` will be output

### Generate swagger.json

```
go run ./examples/openapi/main.go > ./examples/openapi/swagger.json
```

