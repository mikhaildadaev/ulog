---
outline: deep
---

# API / Запись в файл / Основное

::: info Информация
`SinkFile` обеспечивает неблокирующую атомарную ротацию файлов со сжатием gzip. Ваш сервис никогда не блокируется во время ротации или сжатия логов.
:::

## NewSinkFile
Атомарная ротация файлов со сжатием `gzip`. Неблокирующая — ваш сервис не зависнет во время ротации.
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

| Name                                                                               | Description                             | Default      | 
|------------------------------------------------------------------------------------|-----------------------------------------|--------------|
| [`WithFileMaxAge(dayCount)`](/ru/sinkfile_params-examples#withfilemaxage)          | Maximum days to keep old log files      |           30 |
| [`WithFileMaxBackups(fileCount)`](/ru/sinkfile_params-examples#withfilemaxbackups) | Maximum number of old log files to keep |           10 |
| [`WithFileMaxSize(fileSize)`](/ru/sinkfile_params-examples#withfilemaxsize)        | Maximum file size (MB) before rotation  |          100 |