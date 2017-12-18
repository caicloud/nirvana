ROOT=$(dirname "${BASH_SOURCE}")/../../..

go run ${ROOT}/cmd/openapi-gen/main.go \
    -i github.com/caicloud/nirvana/examples/openapi/pkg/api/v1 \
    -p github.com/caicloud/nirvana/examples/openapi/api


go run ${ROOT}/examples/openapi/main.go > ${ROOT}/examples/openapi/swagger.json
