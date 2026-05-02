---
outline: deep
---

# API / 文件接收器 / 构造函数

::: info **关于**
本页记录了 `SinkHttp`，一个生产就绪的 HTTP 接收器，支持批处理、断路器、去重、重试和采样。您的服务在网络传输期间永远不会被阻塞。
:::

## NewSinkFile
原子文件轮转，支持 `gzip` 压缩。完全非阻塞 — 您的服务在轮转期间不会停顿。
```go
var writer io.Writer = ulog.DefaultWriterOut
sinkFile, err := ulog.NewSinkFile("app.log",
    ulog.WithFileMaxAge(30),
    ulog.WithFileMaxBackups(10),
    ulog.WithFileMaxSize(100),
)
if err != nil {
    fmt.Fprintf(ulog.DefaultWriterErr, "ulog: %v — using stdout instead\n", err)
} else {
    defer sinkFile.Close()
    writer = sinkFile
}
telemetry := ulog.NewTelemetry(
    ulog.WithFormat(ulog.FormatJson),
    ulog.WithMode(ulog.ModeAsync, writer, 10000),
)
defer telemetry.Close()
telemetry.Error(ulog.DataLog,
    ulog.String("message", "critical error"),
    ulog.String("service", "billing"),
)
telemetry.Sync()
```

| Name                                                                      | Description                             | Default | 
|---------------------------------------------------------------------------|-----------------------------------------|---------|
| [`WithFileMaxAge(dayCount)`](/en/sinkfile_params#withfilemaxage)          | Maximum days to keep old log files      |      30 |
| [`WithFileMaxBackups(fileCount)`](/en/sinkfile_params#withfilemaxbackups) | Maximum number of old log files to keep |      10 |
| [`WithFileMaxSize(fileSize)`](/en/sinkfile_params#withfilemaxsize)        | Maximum file size (MB) before rotation  |     100 |