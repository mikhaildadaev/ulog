---
outline: deep
---

# API / 核心 / 类型

::: info 关于
本页记录了所有数据类型 `DataLog`、`DataMetric`、`DataTrace` 以及全部 16 种字段类型。每种字段都附有可运行的代码示例和预期的 JSON 输出。
:::


## Data
一个 API 支持三种信号类型：日志、指标和追踪
### Log
人类可读的日志消息
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog, 
    ulog.String("message", "user login"),
)
telemetry.Sync()
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
机器指标
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataMetric,
    ulog.String("name", "logins"),
    ulog.Float64("value", 1.0),
)
telemetry.Sync()
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
分布式追踪
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataTrace,
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
    "type":"trace",
    "span_id":"def",
    "name":"login",
    "duration":150
}
```

## Field
16 个类型安全的字段构造函数。
### Bool
Boolean 字段
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Bool("bool", true),
)
telemetry.Sync()
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
Boolean 切片
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Bools("bools", []bool{true, false}),
)
telemetry.Sync()
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
Duration 字段
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Duration("duration", 5*time.Second),
)
telemetry.Sync()
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
Duration 切片
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Durations("durations", []time.Duration{5*time.Second, 10*time.Second}),
)
telemetry.Sync()
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
Error 字段
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Error(fmt.Errorf("err")),
)
telemetry.Sync()
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
Errors 切片
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Errors([]error{fmt.Errorf("err1"), fmt.Errorf("err2")}),
)
telemetry.Sync()
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
Float64 字段
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Float64("float64", 3.14159),
)
telemetry.Sync()
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
Float64 切片
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Floats64("floats64", []float64{1.5, 2.5}),
)
telemetry.Sync()
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
Int 字段
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Int("int", 42),
)
telemetry.Sync()
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
Int 切片
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Ints("ints", []int{10, 20, 30}),
)
telemetry.Sync()
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
Int64 字段
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Int64("int64", 1234567890),
)
telemetry.Sync()
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
Int64 切片
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Ints64("ints64", []int64{1234567890, 9876543210}),
)
telemetry.Sync()
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
String 字段
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.String("string", "str"),
)
telemetry.Sync()
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
String 切片
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Strings("strings", []string{"str1", "str2", "str3"})
)
telemetry.Sync()
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
Time 字段
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Time("time", time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)),
)
telemetry.Sync()
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
Time 切片
```go
telemetry := ulog.NewTelemetry()
defer telemetry.Close()
telemetry.Info(ulog.DataLog,
    ulog.Times("times", []time.Time{time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC),time.Date(2025, 4, 22, 12, 0, 0, 0, time.UTC)}),
)
telemetry.Sync()
```
Output:
```json
{
    "level":"info",
    "type":"log",
    "times":["2026-04-22T12:00:00.000000+00:00","2025-04-22T12:00:00.000000+00:00"]
}
```
