[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/mikhaildadaev/ulog/blob/main/LICENSE.md)
[![Go Reference](https://pkg.go.dev/badge/github.com/mikhaildadaev/ulog.svg)](https://pkg.go.dev/github.com/mikhaildadaev/ulog)
[![Go Report Card](https://goreportcard.com/badge/github.com/mikhaildadaev/ulog)](https://goreportcard.com/report/github.com/mikhaildadaev/ulog)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mikhaildadaev/ulog)](https://github.com/mikhaildadaev/ulog)
[![CI](https://github.com/mikhaildadaev/ulog/actions/workflows/ci.yml/badge.svg)](https://github.com/mikhaildadaev/ulog/actions/workflows/ci.yml)

# ULOG Toolkit

A high-performance, zero-dependency structured logger for Go with JSON and Text outputs, colored themes, async writer, and full context support.

## Features

- **Blazing fast** – nanoseconds per log, minimal allocations;
- **Two output formats** – JSON (machine) and TEXT (human) with colors;
- **Async writer** – non‑blocking logging with configurable buffer;
- **Context‑aware** – automatic field extraction (trace_id, user_id, etc.);
- **Structured fields** – type‑safe fields: `String()`, `Int()`, `Time()`, `Error()`, etc.;
- **Fully configurable** – functional options: `WithLevel`, `WithTheme`, `WithMode`, etc.;
- **Zero dependencies** – only standard library;
- **Tested** – race‑free, high coverage, benchmarks included.

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
- ulog.Debug(message string, fields ...Field)
- ulog.DebugWithContext(ctx context.Context, msg string, fields ...Field)
- ulog.Error(message string, fields ...Field)
- ulog.ErrorWithContext(ctx context.Context, msg string, fields ...Field)
- ulog.Fatal(message string, fields ...Field)
- ulog.FatalWithContext(ctx context.Context, msg string, fields ...Field)
- ulog.Info(message string, fields ...Field)
- ulog.InfoWithContext(ctx context.Context, msg string, fields ...Field)
- ulog.Warn(message string, fields ...Field)
- ulog.WarnWithContext(ctx context.Context, msg string, fields ...Field)
- ulog.SetExtractor(keys ...string)
- ulog.SetFormat(format TypeFormat)
- ulog.SetLevel(level TypeLevel)
- ulog.SetMode(mode TypeMode, writer io.Writer, bufferSize ...int)
- ulog.SetTheme(theme TypeTheme)
- ulog.Sync() error

### Methods

#### API
- ulog.WithExtractor(keys ...string)
- ulog.WithFormat(format TypeFormat)
- ulog.WithLevel(level TypeLevel)
- ulog.WithMode(mode TypeMode, writer io.Writer, bufferSize ...int)
- ulog.WithTheme(theme TypeTheme)

## Performance

### Multi Thread

|   Level   |  Mode | Format | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-----------|-------|--------|------------|--------------|---------------|--------|
| **Debug** | Async | Simple |       1.0B |	        0.58 |	           0 |      0 |
| **Debug** | Async | Format |       1.0B |    	    0.57 |             0 |      0 |
| **Debug** |  Sync | Simple |       1.0B |    	    0.57 |	           0 |      0 |
| **Debug** |  Sync | Format |       1.0B |    	    0.57 |	           0 |      0 |
|  **Info** | Async | Simple |       6.8M |       165.80 |	         104 |      2 |
|  **Info** | Async | Format |       4.1M |       273.70 |	         728 |      5 |
|  **Info** |  Sync | Simple |      11.4M |    	  103.80 | 	       	  24 |      1 |
|  **Info** |  Sync | Format |       7.0M |       172.50 |	         616 |      4 |
|  **Warn** | Async | Simple |       7.3M |    	  161.40 |        	 104 |      2 |
|  **Warn** | Async | Format |       4.4M |       242.50 |        	 728 |      5 |
|  **Warn** |  Sync | Simple |      11.5M |       102.00 |        	  24 |      1 |
|  **Warn** |  Sync | Format |       6.9M |    	  164.20 |        	 616 |      4 |
| **Error** | Async | Simple |       6.2M |       203.90 |        	 120 |      2 |
| **Error** | Async | Format |       4.0M |       275.70 |        	 728 |      5 |
| **Error** |  Sync | Simple |      11.4M |       104.90 |        	  24 |      1 |
| **Error** |  Sync | Format |       6.0M |       200.20 |        	 616 |      4 |

### Single Thread

|   Level   |  Mode | Format | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-----------|-------|--------|------------|--------------|---------------|--------|
| **Debug** | Async | Simple |     304.6M |         4.21 |             0 |      0 |
| **Debug**	| Async | Format |     288.9M |         5.29 |	           0 |      0 |
| **Debug**	|  Sync | Simple |     283.9M |         3.92 |	           0 |      0 |
| **Debug**	|  Sync | Format |     307.7M |         3.85 |	           0 |      0 |
|  **Info**	| Async | Simple |       2.1M |       551.90 |	         104 |      2 |
|  **Info**	| Async | Format |       1.4M |       826.50 |	         728 |      5 |
|  **Info**	|  Sync | Simple |       3.0M |       388.70 |	          24 |      1 |
|  **Info**	|  Sync | Format |       1.8M |       634.70 |	         616 |      4 |
|  **Warn**	| Async | Simple |       2.0M |       586.40 |	         104 |      2 |
|  **Warn**	| Async | Format |       1.5M |       818.80 |	         728 |      5 |
|  **Warn**	|  Sync | Simple |       3.0M |       394.30 |	          24 |      1 |
|  **Warn**	|  Sync | Format |       1.9M |       627.70 |	         616 |      3 |
| **Error**	| Async | Simple |       2.1M |       587.30 |	         120 |      2 |
| **Error**	| Async | Format |       1.4M |       824.20 |	         728 |      5 |
| **Error**	|  Sync | Simple |       3.0M |       395.90 |	          24 |      1 |
| **Error**	|  Sync | Format |       1.8M |       647.10 |	         616 |	    4 |

> **Note:**
> - `Format` benchmarks use `WithExtractor("trace_id")` to automatically extract from context.
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
    ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
    // Universal logger [Async] with JSON output
    loggerAsync := ulog.NewLogger(
        ulog.WithMode(ulog.ModeAsync, os.Stdout, 10000),
        ulog.WithFormat(ulog.FormatJson),
    )
    defer loggerAsync.Close()
    loggerAsync.Debug("debugging request", ulog.String("path", "/api/user"))
    loggerAsync.DebugWithContext(ctx, "debugging request", ulog.String("path", "/api/user"))
    loggerAsync.Info("server started", ulog.Int("port", 8080))
    loggerAsync.InfoWithContext(ctx, "server started", ulog.Int("port", 8080))
    loggerAsync.Warn("high latency", ulog.Duration("latency", 150*time.Millisecond))
    loggerAsync.WarnWithContext(ctx, "high latency", ulog.Duration("latency", 150*time.Millisecond))
    loggerAsync.Error("database error", ulog.Error(nil))
    loggerAsync.ErrorWithContext(ctx, "database error", ulog.Error(nil))
    loggerAsync.Sync()
    // Universal logger [Sync] with colored text output
    loggerSync := ulog.NewLogger(
        ulog.WithMode(ulog.ModeSync, os.Stdout),
        ulog.WithFormat(ulog.FormatText),
        ulog.WithTheme(ulog.ThemeDark),
    )
    loggerSync.Debug("debugging request", ulog.String("path", "/api/user"))
    loggerSync.DebugWithContext(ctx, "debugging request", ulog.String("path", "/api/user"))
    loggerSync.Info("server started", ulog.Int("port", 8080))
    loggerSync.InfoWithContext(ctx, "server started", ulog.Int("port", 8080))
    loggerSync.Warn("high latency", ulog.Duration("latency", 150*time.Millisecond))
    loggerSync.WarnWithContext(ctx, "high latency", ulog.Duration("latency", 150*time.Millisecond))
    loggerSync.Error("database error", ulog.Error(nil))
    loggerSync.ErrorWithContext(ctx, "database error", ulog.Error(nil))
    // Standard logger adapter (writes only errors)
    logger := ulog.NewLogger(
        ulog.WithMode(ulog.ModeSync, os.Stdout),
        ulog.WithFormat(ulog.FormatJson),
    )
    loggerLog := ulog.NewLoggerLog(ulog.LevelError, logger)
    loggerLog.Print("error from standard logger")
}
```

## Limits

- Async buffer: if full, log is written synchronously (no blocking)
- Caller information: only for LevelDebug (performance)
- Field keys: any string, will be JSON-escaped

## Tests and Benchmarks

Run:

```bash
go test ./...
go test -bench=. ./...
go test -cover ./...
go test -race ./...
```

## Roadmap

- [ ] **More `io.Writer` implementations** – Discord, File, Telegram, Slack, Loki, Elasticsearch, OpenTelemetry