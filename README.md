# prometheus-chi-metric-middleware

Easy prometheus http server metrics for the chi router.

Based on [chi-prometheus](https://github.com/766b/chi-prometheus)

## Install

```bash
go get github.com/TomL-dev/prometheus-chi-metric-middleware
```

## Usage

```go
r := chi.NewRouter()

metricMiddleware := chiprometheus.NewMiddleware(metrics.Options{
	Namespace: "mock",
})

r.Use(metricMiddleware.Handler)
```

For other info, see the [example dir](./example/main.go)

## Result

On the `metric` endpoint of prometheus you'll see the timeseries created by the middleware. The following timeseries will be defined:

- [Counter] http_request_total{code, path}
- [Histogram] http_request_duration_nanoseconds{path}
- [Histogram] http_response_size_bytes{path}

## TODO

- custom bucket sizes
- now by default, it's matched on vars, instead of the actual values. (like /user/{id}). maybe make optional?
- exclusion patterns?