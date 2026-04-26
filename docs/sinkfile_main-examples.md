---
outline: deep
---

# API / FileSink / Main

## NewFileSink
Atomic file rotation with gzip compression. Non-blocking — your service won't stall during rotation
```go
fileSink, err := ulog.NewFileSink("app.log",
    ulog.WithFileMaxAge(30),
    ulog.WithFileMaxBackups(10),
    ulog.WithFileMaxSize(100),
)
if err != nil {
    panic(err)
}
defer fileSink.Close()
telemetry := ulog.NewTelemetry(
    ulog.WithFormat(ulog.FormatJson),
    ulog.WithMode(ulog.ModeAsync, fileSink, 10000),
)
defer telemetry.Close()
telemetry.Error(ulog.DataLog,
    ulog.String("message", "critical error"),
    ulog.String("service", "billing"),
)
telemetry.Sync()
```

## Params

| Name                         | Description                             | Default | 
|------------------------------|-----------------------------------------|---------|
| `WithFileMaxAge(number)`     | Maximum days to keep old log files      |      30 |
| `WithFileMaxBackups(number)` | Maximum number of old log files to keep |      10 |
| `WithFileMaxSize(number)`    | Maximum file size (MB) before rotation  |     100 |