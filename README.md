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
