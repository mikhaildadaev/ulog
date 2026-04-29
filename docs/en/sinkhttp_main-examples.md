---
outline: deep
---

# API / SinkHttp / Main

::: info Info
`SinkHttp` — production-ready HTTP sink with batching, circuit breaker, deduplication, retry, and sampling. Your service never blocks during network delivery.
:::

## NewSinkHttp
Creates an HTTP sink for sending logs to a remote endpoint with `Batching`, `Circuit Breaker`, `Deduplication`, `Retry`, `Sampling` built in
```go
sinkHttp := ulog.NewSinkHttp("http://localhost:8080/logs",
    ulog.WithHttpBatch(100, 5*time.Second),
    ulog.WithHttpCircuitBreaker(10, 10*time.Second),
    ulog.WithHttpDedupWindow(5*time.Second),
    ulog.WithHttpHeader("Authorization", "Bearer token"),
    ulog.WithHttpRetry(3, time.Second),
    ulog.WithHttpSampleRate(100),
    ulog.WithHttpTimeout(30*time.Second),
)
defer sinkHttp.Close()
telemetry := ulog.NewTelemetry(
    ulog.WithFormat(ulog.FormatJson),
    ulog.WithMode(ulog.ModeAsync, sinkHttp, 10000),
)
defer telemetry.Close()
telemetry.Error(ulog.DataLog,
    ulog.String("message", "payment failed"),
    ulog.String("service", "billing"),
)
telemetry.Sync()
```

| Name                                                                                                  | Description                                                                            | Default      |
|-------------------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|--------------|
| [`WithHttpBatch(size, flushInterval)`](/en/sinkhttp_params-examples#withhttpbatch)                    | Batch messages: send up to `size` messages or every `flushInterval`                    | `100, 5s`    |
| [`WithHttpCircuitBreaker(maxFailures, timeout)`](/en/sinkhttp_params-examples#withhttpcircuitbreaker) | Open circuit after `maxFailures` errors, wait `timeout` before recovery                | `10, 10s`    |
| [`WithHttpDedupWindow(window)`](/en/sinkhttp_params-examples#withhttpdedupwindow)                     | Ignore duplicate messages within `window` time                                         | `0`          |
| [`WithHttpDisabledBatch()`](/en/sinkhttp_params-examples#withhttpdisabledbatch)                       | Disable message batching (send immediately)                                            | `false`      |
| [`WithHttpDisabledCircuit()`](/en/sinkhttp_params-examples#withhttpdisabledcircuit)                   | Disable Circuit Breaker                                                                | `false`      |
| [`WithHttpDisableKeepAlive()`](/en/sinkhttp_params-examples#withhttpdisablekeepalive)                 | Disable HTTP Keep-Alive connections                                                    | `false`      |
| [`WithHttpFilterData(type)`](/en/sinkhttp_params-examples#withhttpfilterdata)                         | Filter by data type: `DataLog`, `DataMetric`, `DataTrace`                              | (all)        |
| [`WithHttpFilterLevel(level)`](/en/sinkhttp_params-examples#withhttpfilterlevel)                      | Filter by minimum level: `LevelDebug`,`LevelError`,`LevelFatal`,`LevelInfo`,`LevelWarn`| `LevelError` |
| [`WithHttpFormatter(fn)`](/en/sinkhttp_params-examples#withhttpformatter)                             | Custom formatter function `func(attributes, fields) ([]byte, error)`                   |              |
| [`WithHttpHeader(key, value)`](/en/sinkhttp_params-examples#withhttpheader)                           | Add custom HTTP header                                                                 |              |
| [`WithHttpMethod(method)`](/en/sinkhttp_params-examples#withhttpmethod)                               | HTTP method: `POST`, `PUT`, etc.                                                       | `POST`       |
| [`WithHttpRetry(maxRetries, backoff)`](/en/sinkhttp_params-examples#withhttpretry)                    | Retry failed requests up to `maxRetries` times with exponential `backoff`              | `0, 1s`      |
| [`WithHttpSampleRate(rate)`](/en/sinkhttp_params-examples#withhttpsamplerate)                         | Sample 1 out of `rate` messages for non-error levels                                   | `0`          |
| [`WithHttpSampleWindow(window)`](/en/sinkhttp_params-examples#withhttpsamplewindow)                   | Reset sample counter every `window`                                                    | `0`          |
| [`WithHttpTimeout(timeout)`](/en/sinkhttp_params-examples#withhttptimeout)                            | HTTP client timeout                                                                    | `10s`        |
