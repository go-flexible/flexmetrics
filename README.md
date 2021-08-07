# Metrics Server

The package `go-flexible/flexmetrics` provides a default set of configuration for hosting prometheus and `pprof` metrics.

## Configuration

The metric server can be configured through the environment to match setup in the infrastructure.

- `PROMETHEUS_ADDR` default: `:2112`
- `PROMETHEUS_PATH` default: `/metrics`

## Examples

### Starting server and exposing metrics

```go
srv := metricsrv.New(&metricsrv.Config{
    Server: &http.Server{
        Addr: "0.0.0.0:5117",
    },
    Path: "/metrics",
})
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


