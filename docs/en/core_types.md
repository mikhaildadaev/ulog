---
outline: deep
---

# API / Core / Types

::: info **Info**
This page documents all data types `DataLog`, `DataMetric`, `DataTrace` and all 16 field types. Each field is shown with a working code example and expected JSON output.
:::

## Data
One API for three signal types: logs, metrics, and traces
### Log
Human-readable log messages
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog, 
    ulog.String("message", "user login"),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "message":"user login"
}
```
### Metric
Machine metrics
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataMetric,
    ulog.String("name", "logins"),
    ulog.Float64("value", 1.0),
)
```
Output:
```json
{
    "level":"info",
    "type":"metric",
    "name":"logins",
    "value":1.0
}
```
### Trace
Distributed traces
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataTrace,
    ulog.String("span_id", "def"),
    ulog.String("name", "login"),
    ulog.Int64("duration", 150),
)
```
Output:
```json
{
    "level":"info",
    "type":"trace",
    "span_id":"def",
    "name":"login",
    "duration":150
}
```

## Field
16 type-safe field constructors.
### Bool
Boolean field
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Bool("bool", true),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "bool":true
}
```

### Bools
Boolean slice
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Bools("bools", []bool{true, false}),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "bools":[true,false]
}
```

### Duration
Duration field
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Duration("duration", 5*time.Second),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "duration":"5s"
}
```

### Durations
Duration slice
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Durations("durations", []time.Duration{5*time.Second, 10*time.Second}),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "durations":["5s","10s"]
}
```

### Error
Error field
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Error(fmt.Errorf("err")),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "error":"err"
}
```

### Errors
Error slice
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Errors([]error{fmt.Errorf("err1"), fmt.Errorf("err2")}),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "errors":["err1","err2"]
}
```

### Float64
Float64 field
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Float64("float64", 3.14159),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "float64":3.14159
}
```

### Floats64
Float64 slice
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Floats64("floats64", []float64{1.5, 2.5}),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "floats64":[1.5,2.5]
}
```

### Int
Int field
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Int("int", 42),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "int":42
}
```

### Ints
Int slice
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Ints("ints", []int{10, 20, 30}),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "ints":[10,20,30]
}
```

### Int64
Int64 field
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Int64("int64", 1234567890),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "int64":1234567890
}
```

### Ints64
Int64 slice
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Ints64("ints64", []int64{1234567890, 9876543210}),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "ints64":[1234567890,9876543210]
}
```

### String
String field
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.String("string", "str"),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "string":"str"
}
```

### Strings
String slice
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Strings("strings", []string{"str1", "str2", "str3"})
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "strings":["str1","str2","str3"]
}
```

### Time
Time field
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Time("time", time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "time":"2026-04-22T12:00:00.000000+00:00"
}
```

### Times
Time slice
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Times("times", []time.Time{time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),time.Date(2025, 4, 22, 12, 0, 0, 0, time.UTC)}),
)
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "times":["2026-04-22T12:00:00.000000+00:00","2025-04-22T12:00:00.000000+00:00"]
}
```
