# Workflow

1. Tracing collector starts as infra service
2. **API server starts with configable tracing reporter**
3. External client calls API
4. **API server processes: tracing middleware extracts or starts span and store in context**
5. Actual handler extracts span and add log events
6. Actual handler spawns child spans for sub-operations, including
  - Do something, call span(tracing) API directly
  - Use SDK client which supports tracing to call other service
  - **Use generic client which supports tracing to call other service, e.g. through http**

NIRVANA should support **2. 4. 6.3** .

# API

https://github.com/opentracing/opentracing-go seems to be only choice
except starting from scratch, which is not recommended.
[See data model and spec](https://github.com/opentracing/specification/blob/master/specification.md)

# Implementation

- Zipkin
  - Collector server: https://github.com/openzipkin/zipkin
  - Go client: https://github.com/openzipkin/zipkin-go-opentracing
- Jaeger
  - Collect server: https://github.com/jaegertracing/jaeger
  - Go client: https://github.com/uber/jaeger-client-go
- Others, maybe ...

**Personal Notes (@xiaoq17):**

- All clients implement opentracing API, so they work fine with others' collector server.
- Product or infra or devops chooses server:
  - Server performance is not benchmarked, no recommendation.
  - Jaeger web UI looks better.
- NIRVANA just picks one client:
  - Neutral before discussion.
  - Examples:
    - https://github.com/openzipkin/zipkin-go-opentracing/tree/master/examples
    - https://github.com/uber/jaeger-client-go/tree/master/transport
