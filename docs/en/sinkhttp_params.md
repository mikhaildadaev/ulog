---
outline: deep
---

# API / SinkHttp / Params

::: info **Info**
This page covers all configuration options for `SinkHttp`: batching, circuit breaker, deduplication, retry, sampling, and more. Each option is shown with its default value and a brief description.
:::

## WithHttpBatch
Batch messages: send up to `size` messages or every `flushInterval`.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpBatch(100, 5*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpCircuitBreaker
Open circuit after `maxFailures` errors, wait `timeout` before recovery.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpCircuitBreaker(10, 10*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpDedupWindow
Ignore duplicate messages within `window` time.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDedupWindow(5*time.Second),
)
defer sinkHttp.Close()
```

## WithHttpDisabledBatch
Disable message batching (send immediately).
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisabledBatch(),
)
defer sinkHttp.Close()
```

## WithHttpDisabledCircuit
Disable Circuit Breaker.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisabledCircuit(),
)
defer sinkHttp.Close()
```

## WithHttpDisableKeepAlive
Disable HTTP Keep-Alive connections.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpDisableKeepAlive(),
)
defer sinkHttp.Close()
```

## WithHttpFilterData
Filter by data type: `DataLog`, `DataMetric`, `DataTrace`.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpFilterData(ulog.DataLog),
)
defer sinkHttp.Close()
```

## WithHttpFilterLevel
Filter by minimum level: `LevelDebug`, `LevelError`, `LevelFatal`, `LevelInfo`, `LevelWarn`.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpFilterLevel(ulog.LevelError),
)
defer sinkHttp.Close()
```

## WithHttpFormatter
Custom formatter function `func(attributes, fields) ([]byte, error)`.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
        // Custom formatting logic
        return json.Marshal(fields)
    }),
)
defer sinkHttp.Close()
```

## WithHttpHeader
Add custom HTTP header.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpHeader("Authorization", "Bearer token"),
    ulog.WithHttpHeader("X-Custom", "value"),
)
defer sinkHttp.Close()
```

## WithHttpMethod
HTTP method: `POST`, `PUT`, etc.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpMethod("POST"),
)
defer sinkHttp.Close()
```

## WithHttpRetry
Retry failed requests up to `maxRetries` times with exponential `backoff`.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpRetry(3, time.Second),
)
defer sinkHttp.Close()
```

## WithHttpSampleRate
Sample 1 out of `rate` messages for non-error levels.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpSampleRate(100),
)
defer sinkHttp.Close()
```

## WithHttpSampleWindow
Reset sample counter every `window`.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpSampleWindow(1*time.Minute),
)
defer sinkHttp.Close()
```

## WithHttpTimeout
HTTP client timeout.
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpTimeout(30*time.Second),
)
defer sinkHttp.Close()
```
