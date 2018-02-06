# Nirvana

<img align="right" width="225px" src="https://user-images.githubusercontent.com/2191361/35839723-e9e5cdfa-0b2c-11e8-853a-8d3870f9e7ac.png">

[![Build Status](https://travis-ci.org/caicloud/nirvana.svg?branch=master)](https://travis-ci.org/caicloud/nirvana)
[![Coverage Status](https://coveralls.io/repos/github/caicloud/nirvana/badge.svg?branch=master)](https://coveralls.io/github/caicloud/nirvana?branch=master)
[![GoDoc](http://godoc.org/github.com/caicloud/nirvana?status.svg)](http://godoc.org/github.com/caicloud/nirvana)
[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/nirvana)](https://goreportcard.com/report/github.com/caicloud/nirvana)
[![OpenTracing Badge](https://img.shields.io/badge/OpenTracing-enabled-blue.svg)](http://opentracing.io)

Nirvana is a golang API framework designed for productivity and usability. It aims to be the building block for
all golang services in Caicloud. The high-level goals and features include:

* consistent API behavior, structure and layout across all golang projects
* improve engineering productivity with openAPI and client generation, etc
* validation can be added by declaring validation method as part of API definition
* out-of-box instrumentation support, e.g. metrics, profiling, tracing, etc
* easy and standard configuration management, as well as standard cli interface

Nirvana is also extensible and performant, with the goal to support fast developmenet velocity.

## Installation

```
go get -u github.com/caicloud/nirvana

# for openapi generation
go get -u github.com/caicloud/nirvana/cmd/openapi-gen
```

## Getting Started

### API Quick Start

In Nirvana, APIs are defined via `definition.Descriptor`. We will not introduce details of the concept `Descriptor`,
instead, let's take a look at a contrived example:

```go
// API descriptor.
var echo = definition.Descriptor{
	Path:        "/echo",
	Description: "Echo API",
	Definitions: []definition.Definition{
		{
			Method: definition.Get,
			Function: Echo,
			Consumes: []string{definition.MIMEAll},
			Produces: []string{definition.MIMEText},
			Parameters: []definition.Parameter{
				{
					Source: definition.Query,
					Name: "msg",
					Description: "Corresponding to the second parameter",
				},
			},
			Results: []definition.Result{
				{
					Destination: definition.Data,
					Description: "Corresponding to the first result",
				},
				{
					Destination: definition.Error,
					Description: "Corresponding to the second result",
				},
			},
		},
	},
}
```

This is an echo server API descriptor. The descriptor is a bit complex at first glance, but is actually quite
simple. Below is a partially translated HTTP language:

```
HTTP Path: /echo[?msg=]
HTTP Method: Get
HTTP Headers:
    Content-Type: Any Type
    Accept: text/plain or */*
```

The request handler `Echo` receives two parameters and returns two results, as defined in our descriptor.
Note the first parameter is always `context.Context` - it is injected by default config.

```go
// API function.
func Echo(ctx context.Context, msg string) (string, error) {
	return msg, nil
}
```

Nirvana will parse incoming request and generate function parameters for `Echo` function as defined via
`Definition.Parameters` - parameters will be converted into the exact type defined in `Echo`. Once done,
Nirvana collects the results and sends back response.

With our API descriptors ready, we can now create a server to serve requests:

```go
package main

import (
	"context"

	"github.com/caicloud/nirvana"
	"github.com/caicloud/nirvana/definition"
	"github.com/caicloud/nirvana/log"
)

func main() {
	config := nirvana.NewDefaultConfig("", 8080)
	config.Configure(nirvana.Descriptor(echo))
	log.Infof("Listening on %s:%d", config.IP, config.Port)
	if err := nirvana.NewServer(config).Serve(); err != nil {
		log.Fatal(err)
	}
}
```

Now run the server and test it:

```
go run ./examples/getting-started/basics/echo.go
INFO  0202-16:34:38.663+08 echo.go:65 | Listening on :8080
INFO  0202-16:34:38.663+08 builder.go:163 | Definitions: 1 Middlewares: 0 Path: /echo
INFO  0202-16:34:38.663+08 builder.go:178 |   Method: Get Consumes: [*/*] Produces: [text/plain]
```

In another terminal:

```
$ curl "http://localhost:8080/echo?msg=test"
test
```

For full example code, see [basics](./examples/getting-started/basics).

### Validate it!

Now you are tired of echoing non-sense testing message and want to only reply message longer than 10 characters, such
validation can be easily added when defining your descriptor:

```go
Parameters: []definition.Parameter{
	{
		Source:      definition.Query,
		Name:        "msg",
		Description: "Corresponding to the second parameter",
		Operators:   []definition.Operator{validator.String("gt=10")},
	},
},
```

`Operator` is a concept in Nirvana to allow framework user to operate on input request; validation is one of several
pre-defined operators. Another example of `operator` is `convertor`, which allows user to convert between different
versions of an input.

Under the hood, Nirvana uses [go-playground/validator.v9](https://github.com/go-playground/validator) for validation,
which defines a list of useful tags. It also supports custom validation. Nirvana integrates smoothly with the package,
see user guide for more advanced usage.

Now run our new echo server and verify validation works:

```
$ go run ./examples/getting-started/validator/echo.go
INFO  0202-11:18:50.235+08 echo.go:67 | Listening on :8080
INFO  0202-11:18:50.235+08 builder.go:163 | Definitions: 1 Middlewares: 0 Path: /echo
INFO  0202-11:18:50.235+08 builder.go:178 |   Method: Get Consumes: [*/*] Produces: [text/plain]
```

In another terminal:

```
$ curl "http://localhost:8080/echo?msg=test"
Key: '' Error:Field validation for '' failed on the 'gt' tag

$ curl "http://localhost:8080/echo?msg=testtesttest"
testtesttest
```

It works! The above example teaches us two facts:

1. Adding validation support with Nirvana is very simple
2. 10 characters validation is not enough to prevent spam :) (checkout guide below to add your own validation)

For full example code, see [validator](./examples/getting-started/validator).

### Is it popular?

It's time to expose some metrics to help understand and diagnose our service! Nirvana has out-of-box support for
instrumentation. To enable exposing request metrics, just add one more configuration:

```go
config := nirvana.NewDefaultConfig("", 8080).
	Configure(
		metrics.Path("/metrics"),
	)
```

The actual configuration is done with `metrics` plugin. `plugin` is another concept in Nirvana - we can always
add more functionalities to Nirvana via plugin, and each plugin can be individually enabled or disabled. How
plugins are implemented depends on plugin author. For example, some plugins are simply static configuration,
while some are more complex middlewares. All plugins are registered into config. The server will install them
when the server starts.

Now if we start our server, we'll see a wealth of information exposed as [prometheus](https://prometheus.io) format.
The metrics are exposed via `/metrics` endpoint.

```
$ go run ./examples/getting-started/metrics/echo.go
```

Use ab (ApacheBench) to simulate some user load; in another terminal, run:

```
ab -n 1000 -H 'Content-type: application/json' 'http://localhost:8080/echo?msg=testtesttest'
```

Once done, let's checkout some default metrics from metrics plugin. The metric `nirvana_request_count` tells
us how many requests we've seen in total. Since we use `-n 1000`, there will be 1000 requests to `/echo` endpoint.

```
$ curl http://localhost:8080/metrics 2>&1 | grep nirvana_request_count
# HELP nirvana_request_count Counter of server requests broken out for each verb, API resource, client, and HTTP response contentType and code.
# TYPE nirvana_request_count counter
nirvana_request_count{client="ApacheBench/2.3",code="200",contentType="application/json",method="GET",path="/echo"} 1000
```

The metric `nirvana_request_latencies` shows distribution of our service latencies. We've added a random sleep
between [0, 100) in our service; therefore, p90 is around 90m.

```
$ curl http://localhost:8080/metrics 2>&1 | grep "nirvana_request_latencies"
# HELP nirvana_request_latencies Response latency distribution in microseconds for each verb, resource and client.
# TYPE nirvana_request_latencies histogram
nirvana_request_latencies_bucket{method="GET",path="/echo",le="125000"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="250000"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="500000"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="1e+06"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="2e+06"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="4e+06"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="8e+06"} 1000
nirvana_request_latencies_bucket{method="GET",path="/echo",le="+Inf"} 1000
nirvana_request_latencies_sum{method="GET",path="/echo"} 48533
nirvana_request_latencies_count{method="GET",path="/echo"} 1000
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="125000"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="250000"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="500000"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="1e+06"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="2e+06"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="4e+06"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="8e+06"} 1
nirvana_request_latencies_bucket{method="GET",path="/metrics",le="+Inf"} 1
nirvana_request_latencies_sum{method="GET",path="/metrics"} 0
nirvana_request_latencies_count{method="GET",path="/metrics"} 1
# HELP nirvana_request_latencies_summary Response latency summary in microseconds for each verb and resource.
# TYPE nirvana_request_latencies_summary summary
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.5"} 53
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.9"} 89
nirvana_request_latencies_summary{method="GET",path="/echo",quantile="0.99"} 98
nirvana_request_latencies_summary_sum{method="GET",path="/echo"} 48533
nirvana_request_latencies_summary_count{method="GET",path="/echo"} 1000
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.5"} 0
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.9"} 0
nirvana_request_latencies_summary{method="GET",path="/metrics",quantile="0.99"} 0
nirvana_request_latencies_summary_sum{method="GET",path="/metrics"} 0
nirvana_request_latencies_summary_count{method="GET",path="/metrics"} 1
```

See user guide for more information about metrics plugin (and others). For full example code, see [metrics](./examples/getting-started/metrics).

### Show me the docs

You've upgraded your service to provide a new endpoint to create an echo message, i.e.

```
curl -H "Content-Type: application/json" -X POST -d '{"name": "alice", "message": "echo to myself"}' http://localhost:8080/echo
```

This is a complicated enpoint. To make it easy for your user, you decide to provide API documentation.
Nirvana has built-in support to generate openapi documentation. To generate the docs, you need to first
define where types come from. In our example, it's in the `api` package:

```go
package api

// Message defines the message to echo and to whom the message will be sent.
// +caicloud:openapi=true
type Message struct {
	Name    string `json:"name" validate:"required"`
	Message string `json:"message" validate:"gt=10"`
}
```

Next step is to generate openapi definitions:

```
openapi-gen \
  -i github.com/caicloud/nirvana/examples/getting-started/openapi/pkg/api \
  -p github.com/caicloud/nirvana/examples/getting-started/openapi/pkg/api
```

Finally, we can build our openapi specification:

```go
swagger, err := builder.BuildOpenAPISpec(&echo, &common.Config{
	Info: &spec.Info{
		InfoProps: spec.InfoProps{
			Title:       "echo server openAPI",
			Description: "This is open API documentation of echo server",
			Contact: &spec.ContactInfo{
				Name: "nirvana",
				URL:  "https://gonirvana.io",
			},
			License: &spec.License{
				Name: "Apache License, Version 2.0",
				URL:  "http://www.apache.org/licenses/LICENSE-2.0",
			},
			Version: "v1.0.0",
		},
	},
	GetDefinitions: api.GetOpenAPIDefinitions,
})
if err != nil {
	panic(err)
}
encoder := json.NewEncoder(os.Stdout)
if err := encoder.Encode(swagger); err != nil {
	panic(err)
}
```

Now run the following command, we can generate our swagger.json file. Put it into https://editor.swagger.io/,
we'll be able to view our generated API docs.

```
go run ./examples/getting-started/openapi/echo.go > /tmp/swagger.json
```

For full example code, see [openapi](./examples/getting-started/openapi).

## User Guide

### API Descriptor

```go
// Descriptor describes a descriptor for API definitions.
type Descriptor struct {
	// Path is the url path. It will inherit parent's path.
	//
	// If parent path is "/api/v1", current is "/some",
	// It means current definitions handles "/api/v1/some".
	Path string
	// Consumes indicates content types that current definitions
	// and child definitions can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates content types that current definitions
	// and child definitions can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Middlewares contains path middlewares.
	Middlewares []Middleware
	// Definitions contains definitions for current path.
	Definitions []Definition
	// Children is used to place sub-descriptors.
	Children []Descriptor
	// Description describes the usage of the path.
	Description string
}
```

An descriptor can contains many API definitions for a path. It consumes `Content-Type` and produces the data with format of `Accept`. It also can have children.
If a child descriptor does not have its own `Consumes` or `Produces`, it would inherit corresponding
field from its parent.
e.g.

```go
definition.Descriptor{
	Path:        "/path",
	Consumes: []string{definition.MIMEAll},
	Produces: []string{definition.MIMEText},
	Definitions: SomeDefinitions,
	Children: []definition.Descriptor{
		{
			Path:        "/child",
			Produces: []string{definition.MIMEJSON},
			Definitions: SomeDefinitions,
		},
	},
}
```

The child of above descriptor is actually identical to:

```go
definition.Descriptor{
	Path:        "/path/child",
	Consumes: []string{definition.MIMEAll},
	Produces: []string{definition.MIMEJSON},
	Definitions: SomeDefinitions,
}
```

There are supported MIME types and their data types:

| MIME            | Consume                        | Produce                        | Note                                                               |
| --------------- | ------------------------------ | ------------------------------ | ------------------------------------------------------------------ |
| MIMENone        | nil                            | nil                            | Can be used into `Consumes` of Get/List and `Produces` of `Delete` |
| MIMEText        | string/[]byte/io.Reader        | string/[]byte/io.Reader        |                                                                    |
| MIMEJSON        | string/[]byte/io.Reader/struct | string/[]byte/io.Reader/struct |                                                                    |
| MIMEXML         | string/[]byte/io.Reader/struct | string/[]byte/io.Reader/struct |                                                                    |
| MIMEOctetStream | string/[]byte/io.Reader        | string/[]byte/io.Reader        |                                                                    |
| MIMEURLEncoded  | nil                            | nil                            | Depends on `Source`. Only be used in `Consumes`                    |
| MIMEFormData    | nil                            | nil                            | Depends on `Source`. Only be used in `Consumes`                    |

`Middlewares` is not like `Consumes` or `Produces`. It impacts paths rather than descriptors. That means a middleware for `/some/path` will impact all paths have prefix `/some/path`. Even though they are in different descriptors.
e.g.

```go
definition.Descriptor{
	Path:        "/path",
	Middlewares: SomeMiddlewares,
}
definition.Descriptor{
	Path:        "/path/child",
}
```

The two descriptors does not have any relationship but their path have common prefix. The first path `/path` is the prefix of the second one. So middlewares is also valid for the second descriptor. For more details, check the design doc of router.

```go
// Definition defines an API handler.
type Definition struct {
	// Method is definition method.
	Method Method
	// Consumes indicates how many content types the handler can consume.
	// It will override parent descriptor's consumes.
	Consumes []string
	// Produces indicates how many content types the handler can produce.
	// It will override parent descriptor's produces.
	Produces []string
	// Function is a function handler. It must be func type.
	Function interface{}
	// Parameters describes function parameters.
	Parameters []Parameter
	// Results describes function retrun values.
	Results []Result
	// Description describes the API handler.
	Description string
	// Examples contains many examples for the API handler.
	Examples []Example
}
```

An definition describes an API handler. `Method` determines the action of the handler:

| Method      | HTTP Method | Success Status Code |
| ----------- | ----------- | ------------------- |
| List        | GET         | 200                 |
| Get         | GET         | 200                 |
| Create      | POST        | 201                 |
| Update      | PUT         | 200                 |
| Patch       | PATCH       | 200                 |
| Delete      | DELETE      | 204                 |
| AsyncCreate | POST        | 202                 |
| AsyncUpdate | PUT         | 202                 |
| AsyncPatch  | PATCH       | 202                 |
| AsyncDelete | DELETE      | 202                 |

Every API method corresponds to a HTTP method and **ONE** Success status code. It's a convention. If an API functions returns with no error, Nirvana should return the status code.

```go
// Parameter describes a function parameter.
type Parameter struct {
	// Source is the parameter value generated from.
	Source Source
	// Name is the name to get value from a request.
	// ex. a query name, a header key, etc.
	Name string
	// Default value is used when a request does not provide a value
	// for the parameter.
	Default interface{}
	// Operators can modify and validate the target value.
	// Parameter value is passed to the first operator, then
	// previous operator's result is as next operator's parameter.
	// The result of last operator will be passed to target function.
	Operators []Operator
	// Description describes the parameter.
	Description string
}
```

`Parameter` describes the corresponding parameter which with same index in API function (If you use `nirvana.NewDefaultConfig()` to create server, All your API functions must use `context.Context` as the first parameter, and you don't need add a `Parameter` for `context.Context` in `Definition.Parameters`. For more details, see `Advanced Usage`).

`Source` is the value source of current parameter. `Name` is the key of `Source` (Not the name of API function parameter).

| Source | Description                                                                                                   |
| ------ | ------------------------------------------------------------------------------------------------------------- |
| Path   | Value from URL path                                                                                           |
| Query  | Value from URL query string                                                                                   |
| Header | Value from HTTP request header                                                                                |
| Form   | Value from HTTP body. But `Content-Type` must be "application/x-www-form-urlencoded" or "multipart/form-data" |
| File   | Value from HTTP body. But `Content-Type` must be "multipart/form-data"                                        |
| Body   | Value from HTTP body. Parameters of the type don't need a name                                                |
| Auto   | Data receiver must be a struct. Parameters of the type don't need a name. Explain later                       |
| Prefab | Value from internal method. See `Advanced Usage`                                                              |

**Source Auto**
Auto is for combining fields in a struct:

```go
// Here is an example for `Auto` struct.
// The struct has some fields. Every field has a tag with name `source`.
// The source should obey the format:
//     Source,Name[,default=value]
// `Source` and `Name` are same as before.
// `default` is optional. its value should be basic data type (bool, int*, uint*, float*, string).
type Example struct {
	ID     int    `source:"Path,id"`
	Start  int    `source:"Query,id,default=100"`
	Tenant string `source:"Header,X-Tenant,default=test"`
}
```

If you have lots of fields from a request, you could use `Auto` with a struct to get values from request. Don't use it when you only have several parameters. Separate parameters is more readable.

**Parameter Workflow**
All values from HTTP request is string. Nirvana has a mechanism to convert strings to specific types for API function. It's normal, but if there are some operators in `Parameter`, the target type of the parameter would be from the first operator.
Here is the data flow for a parameter:
![](https://user-images.githubusercontent.com/13895988/34516454-7215cda8-f03c-11e7-8fcf-e06147c9d98d.png)
If `Data` is empty and `Parameter.Default` is not nil, default value is used as `Typed Data` .

```go
// Result describes how to handle a result from function results.
type Result struct {
	// Destination is the target for the result. Different types make different behavior.
	Destination Destination
	// Operators can modify the result value.
	// Result value is passed to the first operator, then
	// previous operator's result is as next operator's parameter.
	// The result of last operator will be passed to destination handler.
	Operators []Operator
	// Description describes the result.
	Description string
}
```

`Result` is simpler than `Parameter`. Its `Destination` indicates the target to write data.

| Destination | Description                                                                                                                   |
| ----------- | ----------------------------------------------------------------------------------------------------------------------------- |
| Meta        | Indicates the value should be written to HTTP response header. Its type must be `map[string]string`                           |
| Data        | Indicates the value should be written to HTTP response body. The format is decided by HTTP `Accept` and `Definition.Produces` |
| Error       | If an error occurs, `Meta` and `Data` is ignored. Error message would be written to HTTP response body                        |

The workflow of `Result` is similar with `Parameter`. Only the input of operators is the returned value of API function and output would be written to HTTP response.

**Error**
Your error always indicate a 500 status code except your error implements the following the interface:

```go
// Error is a common interface for error.
// If an error implements the interface, type handlers can
// use Code() to get a specified HTTP status code.
type Error interface {
	// Code is a HTTP status code.
	Code() int
	// Message is an object which contains information of the error.
	Message() interface{}
}
```

An error should implement Error interface. It contains a status code and error message. Package `github.com/caicloud/nirvana/errors` provides many helper functions to create standard errors.
Here are two examples:

```go
// Example 1:
// Directly create an error.
// Fields (e.g. ${customer}) in format are corresponds to args (e.g. customer.Name) in order.
errors.NotFound.Error("${customer} is not found", customer.Name)

// Example 2:
// Create an error factory at first.
var CustomerNotFount = errors.NotFound.Build("Project:Customer:CustomerNotFount", "${customer} is not found")
// Then create error by factory.
CustomerNotFount.Error(customer.Name)
// In the solution, you can check if an error is derived by specified factory.
if CustomerNotFount.Derived(err) {
	// Do something.
}
```

Use interface `errors.Error` in function signature is strongly discouraged. You should always use `error` and create errors by the methods referred above.

### Nirvana Config Plugins

Nirvana provides a simple but useful plugin mechanism to create plugins. Nirvana server need a config to configure server options:

```go
// Config describes configuration of server.
type Config struct {
	...
	// Descriptors contains all APIs.
	Descriptors []definition.Descriptor
	...
	// configSet contains all configurations of plugins.
	configSet map[string]interface{}
}

type Configurer func(c *Config) error

// Configure configs by configurers. It panics if an error occurs.
func (c *Config) Configure(configurers ...Configurer) *Config {...}
```

An plugin can install its own config into Nirvana config by `Configurer`s. `Configurer` can modify its own config. For example:

```go
// Descriptor returns a configurer to add descriptors into config.
func Descriptor(descriptors ...definition.Descriptor) Configurer {
	return func(c *Config) error {
		c.Descriptors = append(c.Descriptors, descriptors...)
		return nil
	}
}
```

`nirvana.Descriptor` is an configurer to install API descriptors into Nirvana config (All descriptors should be installed by the configurer rather than add into Nirvana config directly). If your plugin has private config, you can set/get it into Nirvana config by:

```go
// Config gets external config by plugin name.
func (c *Config) Config(name string) interface{} {...}

// Set sets external config by plugin name.
// Set a nil config will delete it.
func (c *Config) Set(name string, config interface{}) {...}
```

Configurers only manipulate config. Nirvana server would install your plugins when server is starting. So all your plugins have their own installer:

```go
// ConfigInstaller is used to install config to service builder.
type ConfigInstaller interface {
	// Name is the external config name.
	Name() string
	// Install installs stuffs before server starting.
	Install(builder service.Builder, config *Config) error
	// Uninstall uninstalls stuffs after server terminating.
	Uninstall(builder service.Builder, config *Config) error
}
```

**Plugin Frame**

```go
func init() {
	nirvana.RegisterConfigInstaller(&pluginInstaller{})
}

// ExternalConfigName is the external config name of tracing.
const ExternalConfigName = "pluginName"

type pluginInstaller struct{}

// Name is the external config name.
func (i *pluginInstaller) Name() string {
	return ExternalConfigName
}

// Install installs config to builder.
func (i *pluginInstaller) Install(builder service.Builder, cfg *nirvana.Config) error {...}

// Uninstall uninstalls stuffs after server terminating.
func (i *pluginInstaller) Uninstall(builder service.Builder, cfg *nirvana.Config) error {...)

// ConfigA configures fieldA.
func ConfigA(fieldA FieldType) nirvana.Configurer {...}

// ConfigB configures fieldB.
func ConfigB() nirvana.Configurer {...}

// Disable returns a configurer to disable current plugin.
func Disable() nirvana.Configurer {
	return func(c *nirvana.Config) error {
		c.Set(ExternalConfigName, nil)
		return nil
	}
}
```

Then user can use the plugin by:

```go
import "/path/to/plugin"

func main() {
	config := nirvana.NewDefaultConfig("", 8080)
	config.Configure(plugin.ConfigA(fieldValue))
}
```
