# nirvana

[![Build Status](https://travis-ci.org/caicloud/nirvana.svg?branch=master)](https://travis-ci.org/caicloud/nirvana)
[![Coverage Status](https://coveralls.io/repos/github/caicloud/nirvana/badge.svg?branch=master)](https://coveralls.io/github/caicloud/nirvana?branch=master)
[![GoDoc](http://godoc.org/github.com/caicloud/nirvana?status.svg)](http://godoc.org/github.com/caicloud/nirvana)
[![Go Report Card](https://goreportcard.com/badge/github.com/caicloud/nirvana)](https://goreportcard.com/report/github.com/caicloud/nirvana)
[![Code Health](https://landscape.io/github/caicloud/nirvana/master/landscape.svg?style=flat)](https://landscape.io/github/caicloud/nirvana/master)

## Introduction

nirvana is a golang http framework designed with productivity and usability in mind. It aims to be
the building block for all golang services in Caicloud. The high-level goals are:

- Reduce api level errors and inconsistency
- Improve engineering productivity via removing repeated work, adding code generation, etc
- Adding new resource type should only require defining struct definition
- Adding validation should only require declaring validation method as part of struct definition
- Consistent behavior, structure and layout across all golang server projects

## Features

Following is a list of requirements we've seen from writing golang services. Some of these features
are general ones while some are specific to Caicloud. Note this is not an exhaustive list.

**Routing, Request & Response**

- Routes mapping from request to function
- Routes grouping
- Request/Response API object marshal/unmarshal
- General middleware support with sane default (logging, recovery, tracing)
- Contextual process chain and parameter injection
- Enforcing API error convention

**Instrumentation**

- Provide default metrics at well-known endpoints for prometheus to scrape
- Tracing is provided by default to allow better troubleshooting
- Profiling can be enabled in debug mode for troubleshootting

**Validation**

- Provide default validation on api types with struct tags
- Support custom validations defined by developers on api types
- Support validation on all parameters (path, query, etc)

**Usability**

- A working project should be brought up with few lines using the framework
- Framework must automatically follow engineering conventions to help developers focus on business logic
- OpenAPI (swagger 2.0) specification can be generated automatically with no extra work
- Provides a well-established layout conforming to golang project layout
- Easy and standard configuration management
- A reasonable support for websocket

## Get Started
In Nirvana, All APIs are described by `definition.Descriptor`. Before explaining, An example would let you have a straightforward sense.
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

// API function.
func Echo(ctx context.Context, msg string) (string, error) {
	return msg, nil
}
```
It's an echo API descriptor. It may be a bit complex at your first sight, but it should be extremely obvious after you understanding the structure.

First, look at the API function `Echo`. It receives two parameters and returns two results. As the default config, `context.Context` always is the first parameter in API function. Apart form the first parameter, the usage of other parts is normal. As you see, it's an `Echo` function.

As the name of `Descriptor`, it describes what an API is. The API only receives requests like:
```
HTTP Path: /echo[?msg=]
HTTP Method: Get
HTTP Headers:
    Content-Type: Any Type
    Accept: text/plain or */*
```
When a request is coming, Nirvana would parse it and generate function parameters for API function by `Definition.Parameters`. Parameters are converted to the type defined in API function from request. After the execution of API function, Nirvana collects the results and writes to request. `Definition.Produces` decides the format of the result.

After defining API descriptors, the remaining work is creating an server to serve requests:
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
Nirvana server has a plugin machanism. All plugins are registered into config. The server will install them when the server is starting.
`nirvana.Descriptor()` is a plugin method for installing API descriptors into config. After server starting, you could test echo API by `http://localhost:8080/echo?msg=test`.

