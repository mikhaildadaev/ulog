---
outline: deep
---

# API / HTTP 接收器 / 构造函数

::: info **关于**
本页记录了 `SinkHttp`，一个生产就绪的 HTTP 接收器，内置批处理、断路器、去重、重试和采样功能。您的服务在网络传输期间永远不会被阻塞。
:::

## NewSinkHttp
创建一个 HTTP 接收器，用于将日志发送到远程端点，内置 `Batching`、`Circuit Breaker`、`Deduplication`、`Retry`、`Sampling`。
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

| Name                                                                                         | Description                                                                            | Default      |
|----------------------------------------------------------------------------------------------|----------------------------------------------------------------------------------------|--------------|
| [`WithHttpBatch(size, flushInterval)`](/en/sinkhttp_params#withhttpbatch)                    | Batch messages: send up to `size` messages or every `flushInterval`                    | `100, 5s`    |
| [`WithHttpCircuitBreaker(maxFailures, timeout)`](/en/sinkhttp_params#withhttpcircuitbreaker) | Open circuit after `maxFailures` errors, wait `timeout` before recovery                | `10, 10s`    |
| [`WithHttpDedupWindow(window)`](/en/sinkhttp_params#withhttpdedupwindow)                     | Ignore duplicate messages within `window` time                                         | `0`          |
| [`WithHttpDisabledBatch()`](/en/sinkhttp_params#withhttpdisabledbatch)                       | Disable message batching (send immediately)                                            | `false`      |
| [`WithHttpDisabledCircuit()`](/en/sinkhttp_params#withhttpdisabledcircuit)                   | Disable Circuit Breaker                                                                | `false`      |
| [`WithHttpDisableKeepAlive()`](/en/sinkhttp_params#withhttpdisablekeepalive)                 | Disable HTTP Keep-Alive connections                                                    | `false`      |
| [`WithHttpFilterData(type)`](/en/sinkhttp_params#withhttpfilterdata)                         | Filter by data type: `DataLog`, `DataMetric`, `DataTrace`                              | (all)        |
| [`WithHttpFilterLevel(level)`](/en/sinkhttp_params#withhttpfilterlevel)                      | Filter by minimum level: `LevelDebug`,`LevelError`,`LevelFatal`,`LevelInfo`,`LevelWarn`| `LevelError` |
| [`WithHttpFormatter(fn)`](/en/sinkhttp_params#withhttpformatter)                             | Custom formatter function `func(attributes, fields) ([]byte, error)`                   |              |
| [`WithHttpHeader(key, value)`](/en/sinkhttp_params#withhttpheader)                           | Add custom HTTP header                                                                 |              |
| [`WithHttpMethod(method)`](/en/sinkhttp_params#withhttpmethod)                               | HTTP method: `POST`, `PUT`, etc.                                                       | `POST`       |
| [`WithHttpRetry(maxRetries, backoff)`](/en/sinkhttp_params#withhttpretry)                    | Retry failed requests up to `maxRetries` times with exponential `backoff`              | `0, 1s`      |
| [`WithHttpSampleRate(rate)`](/en/sinkhttp_params#withhttpsamplerate)                         | Sample 1 out of `rate` messages for non-error levels                                   | `0`          |
| [`WithHttpSampleWindow(window)`](/en/sinkhttp_params#withhttpsamplewindow)                   | Reset sample counter every `window`                                                    | `0`          |
| [`WithHttpTimeout(timeout)`](/en/sinkhttp_params#withhttptimeout)                            | HTTP client timeout                                                                    | `10s`        |
