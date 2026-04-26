---
outline: deep
---

# Core / Options

## Extractor
Automatic context extraction. Fields from `context.Context` are added to every log, metric, and trace automatically
```go
ctx := context.Background()
ctx = context.WithValue(ctx, "node_id", "123-abc")
ctx = context.WithValue(ctx, "trace_id", "abc-123")
telemetry := ulog.NewTelemetry(
    ulog.WithExtractor("node_id", "trace_id")
)
defer telemetry.Close()
telemetry.InfoWithContext(ctx, ulog.DataLog, 
    ulog.String("message", "user login"),
)
telemetry.InfoWithContext(ctx, ulog.DataMetric,
    ulog.String("name", "logins"),
    ulog.Float64("value", 1.0),
)
telemetry.InfoWithContext(ctx, ulog.DataTrace,
    ulog.String("span_id", "def"),
    ulog.String("name", "login"),
    ulog.Int64("duration", 150),
)
telemetry.Sync()
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "message":"user login",
    "node_id":"123-abc",
    "trace_id":"abc-123"
}
{
    "level":"info",
    "type":"metric",
    "name":"logins",
    "value":1,
    "node_id":"123-abc",
    "trace_id":"abc-123"
}
{
    "level":"info",
    "type":"trace",
    "span_id":"def",
    "name":"login",
    "duration":150,
    "node_id":"123-abc",
    "trace_id":"abc-123"
}
```

## Format
Switch between Text and JSON output on the fly
```go
telemetry := ulog.NewTelemetry(
    ulog.WithFormat(ulog.FormatJson),
)
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.String("message", "json message"),
)
telemetry.Sync()
telemetry.SetFormat(ulog.FormatText)
telemetry.Info(ulog.DataLog,
    ulog.String("message", "text message"),
)
telemetry.Sync()
```
Output:
```json
{"level":"info","type":"log","message":"json message"}
```
```text
[INFO] type="log" message="text message"
```

## Level
Filter logs by severity. Only messages at or above the configured level are written
```go
telemetry := ulog.NewTelemetry(
    ulog.WithLevel(ulog.LevelDebug),
)
defer telemetry.Close()
telemetry.Debug(ulog.DataLog,
    ulog.String("message", "debug message"),
)
telemetry.Error(ulog.DataLog,
    ulog.String("message", "error message"),
)
telemetry.Info(ulog.DataLog,
    ulog.String("message", "info message"),
)
telemetry.Warn(ulog.DataLog,
    ulog.String("message", "warn message"),
)
telemetry.Sync()
```
Output:
```json
{"level":"debug","type":"log","message":"debug message"}
{"level":"error","type":"log","message":"error message"}
{"level":"info","type":"log","message":"info message"}
{"level":"warn","type":"log","message":"warn message"}
```

## Mode
Switch between synchronous and asynchronous writing on the fly
```go
telemetry := ulog.NewTelemetry(
    ulog.WithMode(ulog.ModeAsync, ulog.DefaultWriterOut, 1000),
)
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.String("message", "async message"),
)
telemetry.Sync()
telemetry.SetMode(ulog.ModeSync, ulog.DefaultWriterOut)
telemetry.Info(ulog.DataLog,
    ulog.String("message", "sync message"),
)
telemetry.Sync()
```
Output:
```json
{"level":"info","type":"log","message":"async message"}
{"level":"info","type":"log","message":"sync message"}
```

## Theme
Switch between Dark and Light color themes for Text output. Themes only affect Text format, not JSON
```go
telemetry := ulog.NewTelemetry(
    ulog.WithFormat(ulog.FormatText),
    ulog.WithTheme(ulog.ThemeDark),
)
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.String("message", "dark message"),
)
telemetry.Sync()
telemetry.SetTheme(ulog.ThemeLight)
telemetry.Info(ulog.DataLog,
    ulog.String("message", "light message"),
)
telemetry.Sync()
```
Output:
```text
[INFO] type="log" message="dark message"
[INFO] type="log" message="light message"
```