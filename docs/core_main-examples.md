---
outline: deep
---

# Core / Main

::: warning
This page is under development
:::

## NewTelemetry
Telemetry instance with all configuration options
```go
ctx := context.Background()
ctx = context.WithValue(ctx, "node_id", "123-abc")
ctx = context.WithValue(ctx, "trace_id", "abc-123")
telemetry := ulog.NewTelemetry(
    ulog.WithExtractor("node_id", "trace_id"),
    ulog.WithFormat(ulog.FormatJson),
    ulog.WithLevel(ulog.LevelDebug),
    ulog.WithMode(ulog.ModeAsync, ulog.DefaultWriterOut, 1000),
    ulog.WithTheme(ulog.ThemeLight),
)
defer telemetry.Close()
telemetry.InfoWithContext(ctx, ulog.DataLog, 
    ulog.String("message", "text"),
)
telemetry.InfoWithContext(ctx, ulog.DataMetric, 
    ulog.String("name", "payments"),
    ulog.Float64("value", 99.99),
)
telemetry.InfoWithContext(ctx, ulog.DataTrace,
    ulog.String("name", "payment_processing"),
    ulog.Int64("duration", 150),
    ulog.String("span_id", "span-456"),
)
telemetry.Sync()
telemetry.SetExtractor()
telemetry.SetFormat(ulog.FormatText)
telemetry.SetLevel(ulog.LevelDebug)
telemetry.SetMode(ulog.ModeSync, ulog.DefaultWriterOut)
telemetry.SetTheme(ulog.ThemeDark),
telemetry.Info(ulog.DataLog,
	ulog.String("message", "text"),
)
telemetry.Info(ulog.DataMetric,
	ulog.String("name", "payments"),
	ulog.Float64("value", 99.99),
)
telemetry.Info(ulog.DataTrace,
	ulog.String("name", "payment_processing"),
	ulog.Int64("duration", 150),
	ulog.String("span_id", "span-456"),
)
telemetry.Sync()
```
Output:
```json
{"level":"info","type":"log","message":"text","node_id":"123-abc","trace_id":"abc-123"}
{"level":"info","type":"metric","name":"payments","value":99.99,"node_id":"123-abc","trace_id":"abc-123"}
{"level":"info","type":"trace","name":"payment_processing","duration":150,"span_id":"span-456","node_id":"123-abc","trace_id":"abc-123"}
```
```text
[INFO] type="log" message="text"
[INFO] type="metric" name="payments" value=99.99
[INFO] type="trace" name="payment_processing" duration=150 span_id="span-456"
```

## NewTelemetryLog
Adapter for standard `log.Logger`
```go
telemetry := ulog.NewTelemetry(
    ulog.WithFormat(ulog.FormatText),
    ulog.WithMode(ulog.ModeSync, ulog.DefaultWriterOut),
)
stdLogger := ulog.NewTelemetryLog(ulog.LevelError, telemetry)
stdLogger.Print("this will be logged as ERROR")
stdLogger.Printf("user %s failed to login", "john")
stdLogger.Println("another error message")
```
Output:
```text
[ERROR] type="log" message="this will be logged as ERROR"
[ERROR] type="log" message="user john failed to login"
[ERROR] type="log" message="another error message"
```

## Options

| Name                                                 | Description                                                                     | Values                                                             | Default      |
|------------------------------------------------------|---------------------------------------------------------------------------------|--------------------------------------------------------------------|--------------|
| [`WithExtractor()`](/core_options-examples#extractor)| Auto-extract fields from `context.Context` by key names                         | `keys ...string`                                                   |              |
| [`WithFormat()`](/core_options-examples#format)      | Output format: structured JSON or human-readable TEXT with optional ANSI colors | `FormatJson`, `FormatText`                                         | `FormatJson` |
| [`WithLevel()`](/core_options-examples#level)        | Minimum log severity. Only messages at or above this level are written          | `LevelDebug`, `LevelError`, `LevelFatal`, `LevelInfo`, `LevelWarn` | `LevelInfo`  |
| [`WithMode()`](/core_options-examples#mode)          | Write mode: non-blocking `ModeAsync` with buffer or blocking `ModeSync`         | `ModeAsync`, `ModeSync`                                            | `ModeSync`   |
| [`WithTheme()`](/core_options-examples#theme)        | ANSI color theme for TEXT output: optimized for dark or light terminals         | `ThemeDark`, `ThemeLight`                                          | `ThemeDark`  |

## Reference

| Name                                          | Description                                    | Values                                                                                                                                                     |
|-----------------------------------------------|------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------|
| [`TypeData`](/core_reference-examples#data)   | Log messages, Prometheus metrics, Tempo traces | `DataLog`, `DataMetric`, `DataTrace`                                                                                                                       |
| [`TypeField`](/core_reference-examples#field) | 16 type-safe field constructors                | `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times` |