---
outline: deep
---

# API / SinkFile / Constructors

::: info **Info**
`SinkFile` provides non-blocking atomic file rotation with `gzip` compression. Your service never blocks during log rotation or compression.
:::

## NewSinkFile
Atomic file rotation with `gzip` compression. Non-blocking — your service won't stall during rotation
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