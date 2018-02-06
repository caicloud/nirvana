# Nirvana

<img align="right" width="225px" src="https://user-images.githubusercontent.com/2191361/35839723-e9e5cdfa-0b2c-11e8-853a-8d3870f9e7ac.png">

[![Build Status](https://travis-ci.org/caicloud/nirvana.svg?branch=master)](https://travis-ci.org/caicloud/nirvana)
[![Coverage Status](https://coveralls.io/repos/github/caicloud/nirvana/badge.svg?branch=master)](https://coveralls.io/github/caicloud/nirvana?branch=master)
[![GoDoc](http://godoc.org/github.com/caicloud/nirvana?status.svg)](http://godoc.org/github.com/caicloud/nirvana)
[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/nirvana)](https://goreportcard.com/report/github.com/caicloud/nirvana)
<!-- [![Code Health](https://landscape.io/github/caicloud/nirvana/master/landscape.svg?style=flat)](https://landscape.io/github/caicloud/nirvana/master) -->

Nirvana is a golang API framework designed for productivity and usability. It aims to be the building block for
all golang services in Caicloud. The high-level goals and features include:

- consistent API behavior, structure and layout across all golang projects
- improve engineering productivity with openAPI and client generation, etc
- validation can be added by declaring validation method as part of API definition
- out-of-box instrumentation support, e.g. metrics, profiling, tracing, etc
- easy and standard configuration management, as well as standard cli interface

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

TODO(ddysher): there's quite a bit manual setup to generate openAPI docs, liubo

### Make it configurable

TODO

## User Guide

TODO
