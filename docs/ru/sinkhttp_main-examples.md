---
outline: deep
---

# API / Запись по сети / Основное

::: warning
This page is under development
:::

## NewSinkHttp
Creates an HTTP sink for sending logs to a remote endpoint with **Batching**, **Circuit Breaker**, **Deduplication**, **Retry**, and **Sampling** built in
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

## Параметры

| Наименование                                                                                       | Описание                                                                            | По умолчанию      |
|--------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|--------------|
| [`WithHttpBatch(size, flushInterval)`](/ru/sinkhttp_params-examples#batch)                    | Batch messages: send up to `size` messages or every `flushInterval`                    | `100, 5s`    |
| [`WithHttpCircuitBreaker(maxFailures, timeout)`](/ru/sinkhttp_params-examples#circuitbreaker) | Open circuit after `maxFailures` errors, wait `timeout` before recovery                | `10, 10s`    |
| [`WithHttpDedupWindow(window)`](/ru/sinkhttp_params-examples#dedupwindow)                     | Ignore duplicate messages within `window` time                                         | `0`          |
| [`WithHttpDisabledBatch()`](/ru/sinkhttp_params-examples#disabledbatch)                       | Disable message batching (send immediately)                                            | `false`      |
| [`WithHttpDisabledCircuit()`](/ru/sinkhttp_params-examples#disabledcircuit)                   | Disable Circuit Breaker                                                                | `false`      |
| [`WithHttpDisableKeepAlive()`](/ru/sinkhttp_params-examples#disablekeepalive)                 | Disable HTTP Keep-Alive connections                                                    | `false`      |
| [`WithHttpFilterData(type)`](/ru/sinkhttp_params-examples#filterdata)                         | Filter by data type: `DataLog`, `DataMetric`, `DataTrace`                              | (all)        |
| [`WithHttpFilterLevel(level)`](/ru/sinkhttp_params-examples#filterlevel)                      | Filter by minimum level: `LevelDebug`,`LevelError`,`LevelFatal`,`LevelInfo`,`LevelWarn`| `LevelError` |
| [`WithHttpFormatter(fn)`](/ru/sinkhttp_params-examples#formatter)                             | Custom formatter function `func(attributes, fields) ([]byte, error)`                   |              |
| [`WithHttpHeader(key, value)`](/ru/sinkhttp_params-examples#header)                           | Add custom HTTP header                                                                 |              |
| [`WithHttpMethod(method)`](/ru/sinkhttp_params-examples#method)                               | HTTP method: `POST`, `PUT`, etc.                                                       | `POST`       |
| [`WithHttpRetry(maxRetries, backoff)`](/ru/sinkhttp_params-examples#retry)                    | Retry failed requests up to `maxRetries` times with exponential `backoff`              | `0, 1s`      |
| [`WithHttpSampleRate(rate)`](/ru/sinkhttp_params-examples#samplerate)                         | Sample 1 out of `rate` messages for non-error levels                                   | `0`          |
| [`WithHttpSampleWindow(window)`](/ru/sinkhttp_params-examples#samplewindow)                   | Reset sample counter every `window`                                                    | `0`          |
| [`WithHttpTimeout(timeout)`](/ru/sinkhttp_params-examples#timeout)                            | HTTP client timeout                                                                    | `10s`        |
