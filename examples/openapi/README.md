# How to get json file for swagger-ui

1. Decide where types coming from

    Put a doc.go in the pkg containing structs you want generate definition for.

    For example, pkg named v1 will have a doc.go as follow.

    ```
    // Package v1 is definition of api
    // +nirvana:openapi=true
    package v1
    ```

    In this example, the pkg is `github.com/caicloud/nirvana/examples/openapi/pkg/api/v1`

2. Decide a pkg to put your openapi_generated.go

    If this pkg folder is empty, you might need to put a doc.go there.

    In this example:
    It is [doc.go](./api/doc.go).

    And the pkg is `github.com/caicloud/nirvana/examples/openapi/api`

3. Generate openapi_generated.go

    ```
    go run ${GOPATH}/src/github.com/caicloud/nirvana/cmd/openapi-gen/main.go \
    -i github.com/caicloud/nirvana/examples/openapi/pkg/api/v1 \
    -p github.com/caicloud/nirvana/examples/openapi/api
    ```

4. Generate json file for swagger-ui server

    ```
    go run main.go > swagger.json
    ```
    You might need to embed code piece in main.go in your own project.
    Check [main.go](./main.go) for detail.

5. Serve your json file

    ```
    python svr.py
    ```

6. Start your swagger-ui server

    `docker run -p 8080:8080 --rm swaggerapi/swagger-ui:v2.2.9`

7. Go to `http://127.0.0.1:8080/`

8. Use `http://127.0.0.1:8000/swagger.json` instead of `http://petstore.swagger.io/v2/swagger.json` and Explore
