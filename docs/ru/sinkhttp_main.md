---
outline: deep
---

# API / Запись по сети / Основное

::: info **Информация**
`SinkHttp` — готовый к промышленной эксплуатации HTTP-синк с батчингом, circuit breaker, дедупликацией, повтором и выборкой. Ваш сервис никогда не блокируется при сетевой доставке.
:::

## NewSinkHttp
Создаёт HTTP-синк для отправки логов на удалённый сервер со встроенными `Batching`, `Circuit Breaker`, `Deduplication`, `Retry`, `Sampling`.
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

| Name                                                                                 | Description                                                                            | Default      |
|--------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|--------------|
| [`WithHttpBatch(size, flushInterval)`](/ru/sinkhttp_params#batch)                    | Batch messages: send up to `size` messages or every `flushInterval`                    | `100, 5s`    |
| [`WithHttpCircuitBreaker(maxFailures, timeout)`](/ru/sinkhttp_params#circuitbreaker) | Open circuit after `maxFailures` errors, wait `timeout` before recovery                | `10, 10s`    |
| [`WithHttpDedupWindow(window)`](/ru/sinkhttp_params#dedupwindow)                     | Ignore duplicate messages within `window` time                                         | `0`          |
| [`WithHttpDisabledBatch()`](/ru/sinkhttp_params#disabledbatch)                       | Disable message batching (send immediately)                                            | `false`      |
| [`WithHttpDisabledCircuit()`](/ru/sinkhttp_params#disabledcircuit)                   | Disable Circuit Breaker                                                                | `false`      |
| [`WithHttpDisableKeepAlive()`](/ru/sinkhttp_params#disablekeepalive)                 | Disable HTTP Keep-Alive connections                                                    | `false`      |
| [`WithHttpFilterData(type)`](/ru/sinkhttp_params#filterdata)                         | Filter by data type: `DataLog`, `DataMetric`, `DataTrace`                              | (all)        |
| [`WithHttpFilterLevel(level)`](/ru/sinkhttp_params#filterlevel)                      | Filter by minimum level: `LevelDebug`,`LevelError`,`LevelFatal`,`LevelInfo`,`LevelWarn`| `LevelError` |
| [`WithHttpFormatter(fn)`](/ru/sinkhttp_params#formatter)                             | Custom formatter function `func(attributes, fields) ([]byte, error)`                   |              |
| [`WithHttpHeader(key, value)`](/ru/sinkhttp_params#header)                           | Add custom HTTP header                                                                 |              |
| [`WithHttpMethod(method)`](/ru/sinkhttp_params#method)                               | HTTP method: `POST`, `PUT`, etc.                                                       | `POST`       |
| [`WithHttpRetry(maxRetries, backoff)`](/ru/sinkhttp_params#retry)                    | Retry failed requests up to `maxRetries` times with exponential `backoff`              | `0, 1s`      |
| [`WithHttpSampleRate(rate)`](/ru/sinkhttp_params#samplerate)                         | Sample 1 out of `rate` messages for non-error levels                                   | `0`          |
| [`WithHttpSampleWindow(window)`](/ru/sinkhttp_params#samplewindow)                   | Reset sample counter every `window`                                                    | `0`          |
| [`WithHttpTimeout(timeout)`](/ru/sinkhttp_params#timeout)                            | HTTP client timeout                                                                    | `10s`        |
