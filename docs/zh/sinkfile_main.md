---
outline: deep
---

# API / 文件接收器 / 主要

::: info **关于**
`SinkFile` 提供非阻塞的原子文件轮转，支持 `gzip` 压缩。您的服务在日志轮转或压缩期间永远不会阻塞。
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