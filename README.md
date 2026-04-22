[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/mikhaildadaev/ulog/blob/main/LICENSE.md)
[![Go Reference](https://pkg.go.dev/badge/github.com/mikhaildadaev/ulog.svg)](https://pkg.go.dev/github.com/mikhaildadaev/ulog)
[![Go Report Card](https://goreportcard.com/badge/github.com/mikhaildadaev/ulog)](https://goreportcard.com/report/github.com/mikhaildadaev/ulog)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mikhaildadaev/ulog)](https://github.com/mikhaildadaev/ulog)
[![CI](https://github.com/mikhaildadaev/ulog/actions/workflows/ci.yml/badge.svg)](https://github.com/mikhaildadaev/ulog/actions/workflows/ci.yml)

# ULOG Toolkit

A high-performance, zero-dependency **Observability 2.0 platform** for Go.  
One API for **Logs**, **Metrics**, and **Traces** with production-ready integrations out of the box. 
Structured, colored, async, context-aware.

## Features

- **Observability 2.0** – One API for Logs, Metrics, and Traces
- **Blazing fast** – 180-580 ns/op, 5.8 µs file write with rotation
- **Atomic file rotation** – Non-blocking, gzip compression, auto-cleanup
- **Circuit Breaker** – Production-ready HTTP sink with retry, dedup, sampling
- **8 ready integrations** – Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat
- **Context-aware** – Automatic `trace_id` extraction
- **Colored output** – Dark/Light themes with auto-detection
- **Zero dependencies** – Only standard library

## Installation

```bash
go get github.com/mikhaildadaev/ulog
```

## Quick API

### Сonstructors

#### API
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

#### API
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

#### API
- ulog.WithExtractor(keys ...string)
- ulog.WithFormat(format TypeFormat)
- ulog.WithLevel(level TypeLevel)
- ulog.WithMode(mode TypeMode, writer io.Writer, bufferSize ...int)
- ulog.WithTheme(theme TypeTheme)

## Performance

### Multi Thread

|         Level        |  Mode | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|----------------------|-------|------------|--------------|---------------|--------|
| **DebugWithContext** | Async |       5.8M |       180.70 |           536 |      3 |
| **DebugWithContext** |  Sync |       6.3M |       203.30 |	       536 |      3 |
| **ErrorWithContext** | Async |       2.0M |       578.30 |          1922 |      6 |
| **ErrorWithContext** |  Sync |       3.2M |       372.10 |          1794 |      5 |
|  **InfoWithContext** | Async |       2.3M |       555.90 |	      1922 |      6 |
|  **InfoWithContext** |  Sync |       3.7M |       326.70 |	      1794 |      5 |
|  **WarnWithContext** | Async |       2.4M |       470.70 |          1922 |      6 |
|  **WarnWithContext** |  Sync |       4.0M |       299.90 |          1794 |      5 |

### Single Thread

|         Level        |  Mode | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|----------------------|-------|------------|--------------|---------------|--------|
| **DebugWithContext** | Async |       2.1M |       567.10 |	       536 |      3 |
| **DebugWithContext** |  Sync |       2.1M |       562.60 |	       536 |      3 |
| **ErrorWithContext** | Async |       1.0M |      1045.00 |	      1922 |      6 |
| **ErrorWithContext** |  Sync |       1.4M |       875.10 |	      1794 |	  5 |
|  **InfoWithContext** | Async |       1.0M |      1006.00 |          1922 |      6 |
|  **InfoWithContext** |  Sync |       1.5M |       810.00 |	      1794 |      5 |
|  **WarnWithContext** | Async |       1.2M |       953.60 |          1922 |      6 |
|  **WarnWithContext** |  Sync |       1.5M |       790.50 |	      1794 |      5 |

> **Note:**
> - Benchmarks use `WithExtractor("node_id", "trace_id")` to automatically extract from context.
> - All benchmarks write to `io.Discard` (equivalent to `/dev/null` on Unix or `NUL` on Windows).
> - This measures only the logging overhead (field formatting, JSON encoding, context extraction) without disk or network I/O.
> - Real-world performance will depend on your output destination (file, network, etc.).
> - *Benchmarked on Intel Core i9-9880H (2.30 GHz)*

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
    // Universal telemetry async mode with JSON output
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
    // Universal telemetry sync mode with TEXT output
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
    // Standard logger adapter (writes only errors)
    telemetry := ulog.NewTelemetry(
        ulog.WithFormat(ulog.FormatJson),
        ulog.WithMode(ulog.ModeSync, os.Stdout),
    )
    telemetryLog := ulog.NewTelemetryLog(ulog.LevelError, telemetry)
    telemetryLog.Print("error from standard logger")
}
```

## Limits

- **Async buffer**: if full, log is written synchronously (no blocking)
- **Caller information**: only for `LevelDebug` (performance optimization)
- **Field keys**: any string, will be JSON-escaped
- **Time precision**: microseconds (6 digits) – sufficient for 99% of use cases, reduces allocations
- **Deduplication cache**: in-memory only, cleared periodically (no persistence across restarts)
- **Circuit Breaker**: resets on application restart (no persistent state)
- **File rotation**: checks size on each write; rotation triggered by first write exceeding limit
- **HTTP batching**: messages may be lost if application crashes before flush
- **Kafka sink**: uses REST Proxy API (not native Kafka protocol) – requires Confluent REST Proxy
- **Loki sink**: uses HTTP API (`/loki/api/v1/push`) – labels must be pre-configured
- **Context extraction**: only works with values stored via `context.WithValue()`
- **Zero dependencies**: by design; no external libraries for features like Kafka native protocol

## Tests and Benchmarks

Run:

```bash
go test ./...
go test -bench=. ./...
go test -cover ./...
go test -race ./...
```

## Roadmap

- **More `io.Writer` implementations** – OpenTelemetry