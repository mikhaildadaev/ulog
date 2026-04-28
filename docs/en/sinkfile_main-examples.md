---
outline: deep
---

# API / SinkFile / Main

::: warning
This page is under development
:::

## NewSinkFile
Atomic file rotation with gzip compression. Non-blocking — your service won't stall during rotation
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

| Name                                                                    | Description                             | Default | 
|-------------------------------------------------------------------------|-----------------------------------------|---------|
| [`WithFileMaxAge(number)`](/en/sinkfile_params-examples#maxage)         | Maximum days to keep old log files      |      30 |
| [`WithFileMaxBackups(number)`](/en/sinkfile_params-examples#maxbackups) | Maximum number of old log files to keep |      10 |
| [`WithFileMaxSize(number)`](/en/sinkfile_params-examples#maxsize)       | Maximum file size (MB) before rotation  |     100 |