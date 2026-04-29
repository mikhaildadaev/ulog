[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/mikhaildadaev/ulog/blob/main/LICENSE.md)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mikhaildadaev/ulog)](https://github.com/mikhaildadaev/ulog)
[![Go Reference](https://pkg.go.dev/badge/github.com/mikhaildadaev/ulog.svg)](https://pkg.go.dev/github.com/mikhaildadaev/ulog)
[![Go Report Card](https://goreportcard.com/badge/github.com/mikhaildadaev/ulog)](https://goreportcard.com/report/github.com/mikhaildadaev/ulog)
[![CI](https://github.com/mikhaildadaev/ulog/actions/workflows/ci.yml/badge.svg)](https://github.com/mikhaildadaev/ulog/actions/workflows/ci.yml)

# ULOG

A high-performance, zero-dependency platform for logs, metrics, and traces.  

## Go
> **Information:**
> The latest stable version of ulog is v1.26.12.

### Get Started
```bash
go get github.com/mikhaildadaev/ulog
```

### Run Test 
```bash
go test ./...
go test -bench=. ./...
go test -cover ./...
go test -race ./...
```

## Key Features
- **Unified API** — One API for logs, metrics, and traces.
- **Context Extraction** — Automatic extraction `node_id`, `trace_id`, etc. from `context.Context`.
- **Colored output** – `Dark` and `Light` themes with auto-detection for TEXT format.
- **16 Field Types** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **SinkFile** — Non-blocking atomic rotation with `gzip`.
- **SinkHttp** — `Batching`, `Circuit Breaker`, `Deduplication`, `Retry`, `Sampling`.
- **8 Integrations** — `Discord`, `Kafka`, `Loki`, `Prometheus`, `Slack`, `Telegram`, `Tempo`, `WeChat`.

## Limits
- **Async buffer**: if full, log is written synchronously (no blocking)
- **Caller information**: only for `LevelDebug` (performance optimization)
- **Time precision**: microseconds (6 digits) — sufficient for 99% of use cases, reduces allocations
- **Deduplication cache**: in-memory only, cleared periodically (no persistence across restarts)
- **Circuit Breaker**: resets on application restart (no persistent state)
- **File rotation**: checks size on each write; rotation triggered by first write exceeding limit
- **HTTP batching**: messages may be lost if application crashes before flush
- **Kafka sink**: uses REST Proxy API (not native Kafka protocol) — requires Confluent REST Proxy
- **Loki sink**: uses HTTP API (`/loki/api/v1/push`) — labels must be pre-configured
- **Context extraction**: only works with values stored via `context.WithValue()`
- **Zero dependencies**: by design; no external libraries for features like Kafka native protocol

## Benchmarks
> **Information:**
> The best way to compare libraries is to run benchmarks in **your own environment** with **your own workload**. Each project has unique requirements — latency, throughput, memory usage, and integration complexity — and no single test can cover them all.
> I recommend that you test `ulog` alongside other libraries and choose the tool that best suits your needs.

### Core Performance
These benchmarks measure the cost of formatting and extracting context by writing to io.Discard.

#### MultiThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **DebugWithContext** |       5.8M |      180.700 |           536 |      3 |
| Async | **ErrorWithContext** |       2.0M |      578.300 |          1922 |      6 |
| Async | **InfoWithContext**  |       2.3M |      555.900 |	      1922 |      6 |
| Async | **WarnWithContext**  |       2.4M |      470.700 |          1922 |      6 |
| Sync  | **DebugWithContext** |       6.3M |      203.300 |	       536 |      3 |
| Sync  | **ErrorWithContext** |       3.2M |      372.100 |          1794 |      5 |
| Sync  | **InfoWithContext**  |       3.7M |      326.700 |	      1794 |      5 |
| Sync  | **WarnWithContext**  |       4.0M |      299.900 |          1794 |      5 |

#### SingleThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **DebugWithContext** |       2.1M |      567.100 |	       536 |      3 |
| Async | **ErrorWithContext** |       1.0M |     1045.000 |	      1922 |      6 |
| Async | **InfoWithContext**  |       1.0M |     1006.000 |          1922 |      6 |
| Async | **WarnWithContext**  |       1.2M |      953.600 |          1922 |      6 |
| Sync  | **DebugWithContext** |       2.1M |      562.600 |	       536 |      3 |
| Sync  | **ErrorWithContext** |       1.4M |      875.100 |	      1794 |	  5 |
| Sync  | **InfoWithContext**  |       1.5M |      810.000 |	      1794 |      5 |
| Sync  | **WarnWithContext**  |       1.5M |      790.500 |	      1794 |      5 |

> **Note:**
> - Benchmarks use `WithExtractor("node_id", "trace_id")` to automatically extract from context.
> - All benchmarks write to `io.Discard` (equivalent to `/dev/null` on Unix or `NUL` on Windows).
> - This measures only the logging overhead (field formatting, JSON encoding, context extraction) without disk or network I/O.
> - Real-world performance will depend on your output destination (file, network, etc.).
> - *Benchmarked on Intel Core i9-9880H (2.30 GHz)*

### SinkFile Performance
Benchmark data writes structured JSON logs to a real file with atomic rotation enabled.

#### MultiThread

| Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|------------|--------------|---------------|--------|
| Async |     999.9K |        6,900 |          1962 |      6 |
| Sync  |     152,7K |        7,800 |          1801 |      5 |

#### SingleThread

| Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|------------|--------------|---------------|--------|
| Async |     969,7K |        6,000 |          1962 |      6 |
| Sync  |     234,4K |        5,500 |          1798 |      5 |

> **Note:**
> - Benchmarks use `WithExtractor("node_id", "trace_id")` to automatically extract from context.
> - Writes structured JSON logs to a **real file** with **atomic rotation** enabled (`WithFileMaxSize(15)`).
> - Includes full overhead: JSON formatting, context extraction, file I/O, and non-blocking rotation checks.
> - *Benchmarked on Intel Core i9-9880H (2.30 GHz)*

### SinkHttp Performance
Benchmark data that measures the internal costs of the ulog HTTP receiver using httptest.Server without network latency.

#### MultiThread

| Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|------------|--------------|---------------|--------|
| Async |     999,9M |       27,000 |         8,400 |     82 |
| Sync  |      45,4K |       26,400 |         9,100 |     89 |

#### SingleThread

| Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|------------|--------------|---------------|--------|
| Async |     555,2K |       42,100 |         9,100 |     82 |
| Sync  |      13,6K |       82,500 |         9,400 |     85 |

> **Note:**
> - Benchmarks use `httptest.Server` to simulate HTTP endpoint.
> - Measures full overhead: JSON formatting, context extraction, HTTP request/response.
> - *Multi* benchmarks use `b.RunParallel` to simulate real-world concurrent load.
> - *Benchmarked on Intel Core i9-9880H (2.30 GHz)*

## Quick navigation

### Сonstructors
- ulog.Bool(key string, value bool) Field
- ulog.Bools(key string, value []bool) Field
- ulog.Duration(key string, value time.Duration) Field
- ulog.Durations(key string, value []time.Duration) Field
- ulog.Error(err error) Field
- ulog.Errors(errs []error) Field
- ulog.Float64(key string, value float64) Field
- ulog.Floats64(key string, value []float64) Field
- ulog.Int(key string, value int) Field
- ulog.Ints(key string, value []int) Field
- ulog.Int64(key string, value int64) Field
- ulog.Ints64(key string, value []int64) Field
- ulog.String(key string, value string) Field
- ulog.Strings(key string, value []string) Field
- ulog.Time(key string, value time.Time) Field
- ulog.Times(key string, value []time.Time) Field

### Functions
- ulog.Close() error
- ulog.Debug(typeData TypeData, fields ...Field)
- ulog.DebugWithContext(ctx context.Context, typeData TypeData, fields ...Field)
- ulog.Error(typeData TypeData, fields ...Field)
- ulog.ErrorWithContext(ctx context.Context, typeData TypeData, fields ...Field)
- ulog.Fatal(typeData TypeData, fields ...Field)
- ulog.FatalWithContext(ctx context.Context, typeData TypeData, fields ...Field)
- ulog.Info(typeData TypeData, fields ...Field)
- ulog.InfoWithContext(ctx context.Context, typeData TypeData, fields ...Field)
- ulog.SetExtractor(keys ...string)
- ulog.SetFormat(format TypeFormat)
- ulog.SetLevel(level TypeLevel)
- ulog.SetMode(mode TypeMode, writer io.Writer, bufferSize ...int)
- ulog.SetTheme(theme TypeTheme)
- ulog.Sync() error
- ulog.Warn(typeData TypeData, fields ...Field)
- ulog.WarnWithContext(ctx context.Context, typeData TypeData, fields ...Field)

### Methods
- ulog.WithExtractor(keys ...string)
- ulog.WithFormat(format TypeFormat)
- ulog.WithLevel(level TypeLevel)
- ulog.WithMode(mode TypeMode, writer io.Writer, bufferSize ...int)
- ulog.WithTheme(theme TypeTheme)

## Usage
```go
import (
    "fmt"
    "log"
    "github.com/mikhaildadaev/ulog"
)

func main() {
    ctx := context.Background()
    ctx = context.WithValue(ctx, "node_id", "123-abc")
    ctx = context.WithValue(ctx, "trace_id", "abc-123")
    telemetryAsync := ulog.NewTelemetry(
        ulog.WithFormat(ulog.FormatJson),
        ulog.WithMode(ulog.ModeAsync, os.Stdout, 10000),
    )
    defer telemetryAsync.Close()
    telemetryAsync.Debug(DataLog, "debugging request", ulog.String("path", "/api/user"))
    telemetryAsync.DebugWithContext(ctx, DataLog, "debugging request", ulog.String("path", "/api/user"))
    telemetryAsync.Info(DataLog, "server started", ulog.Int("port", 8080))
    telemetryAsync.InfoWithContext(ctx, DataLog, "server started", ulog.Int("port", 8080))
    telemetryAsync.Warn(DataLog, "high latency", ulog.Duration("latency", 150*time.Millisecond))
    telemetryAsync.WarnWithContext(ctx, DataLog, "high latency", ulog.Duration("latency", 150*time.Millisecond))
    telemetryAsync.Error(DataLog, "database error", ulog.Error(nil))
    telemetryAsync.ErrorWithContext(ctx, DataLog, "database error", ulog.Error(nil))
    telemetryAsync.Sync()
    telemetrySync := ulog.NewTelemetry(
        ulog.WithFormat(ulog.FormatText),
        ulog.WithMode(ulog.ModeSync, os.Stdout),
        ulog.WithTheme(ulog.ThemeDark),
    )
    telemetrySync.Debug(DataLog, "debugging request", ulog.String("path", "/api/user"))
    telemetrySync.DebugWithContext(ctx, DataLog, "debugging request", ulog.String("path", "/api/user"))
    telemetrySync.Info(DataLog, "server started", ulog.Int("port", 8080))
    telemetrySync.InfoWithContext(ctx, DataLog, "server started", ulog.Int("port", 8080))
    telemetrySync.Warn(DataLog, "high latency", ulog.Duration("latency", 150*time.Millisecond))
    telemetrySync.WarnWithContext(ctx, DataLog, "high latency", ulog.Duration("latency", 150*time.Millisecond))
    telemetrySync.Error(DataLog, "database error", ulog.Error(nil))
    telemetrySync.ErrorWithContext(ctx, DataLog, "database error", ulog.Error(nil))
    telemetry := ulog.NewTelemetry(
        ulog.WithFormat(ulog.FormatJson),
        ulog.WithMode(ulog.ModeSync, os.Stdout),
    )
    telemetryLog := ulog.NewTelemetryLog(ulog.LevelError, telemetry)
    telemetryLog.Print("error from standard logger")
}
```

## Roadmap

- **More `io.Writer` implementations** – OpenTelemetry