# Metrics Server

[![Go](https://github.com/go-flexible/flexmetrics/actions/workflows/go.yml/badge.svg?branch=develop)](https://github.com/go-flexible/flexmetrics/actions/workflows/go.yml)

The package `go-flexible/flexmetrics` provides a default set of configuration for hosting prometheus and `pprof` metrics.

## Configuration

The metric server can be configured through the environment to match setup in the infrastructure.

- `PROMETHEUS_ADDR` default: `:2112`
- `PROMETHEUS_PATH` default: `/metrics`

## Examples

### Starting server and exposing metrics

```go
# Rely on the package defaults
srv := flexmetrics.New()
srv.Run(ctx)

# Or bring your own
httpServer := &http.Server{
  Addr: ":8081",
}
srv := flexmetrics.New(
  flexmetrics.WithServer(httpServer),
  flexmetrics.WithPath("/__/metrics"),
)
srv.Run(ctx)
```

`pprof` metrics will be exposed on:

```text
/debug/pprof/
/debug/pprof/cmdline
/debug/pprof/profile
/debug/pprof/symbol
/debug/pprof/trace
```
