[![MIT License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/mikhaildadaev/ulog/blob/main/LICENSE.md)
[![Go Reference](https://pkg.go.dev/badge/github.com/mikhaildadaev/ulog.svg)](https://pkg.go.dev/github.com/mikhaildadaev/ulog)
[![Go Report Card](https://goreportcard.com/badge/github.com/mikhaildadaev/ulog)](https://goreportcard.com/report/github.com/mikhaildadaev/ulog)
[![Go Version](https://img.shields.io/github/go-mod/go-version/mikhaildadaev/ulog)](https://github.com/mikhaildadaev/ulog)
[![CI](https://github.com/mikhaildadaev/ulog/actions/workflows/ci.yml/badge.svg)](https://github.com/mikhaildadaev/ulog/actions/workflows/ci.yml)

# ULOG Logger

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
| **Debug** | Async | Simple |       1.8B |	        0.56 |	           0 |      0 |
| **Debug** | Async | Format |       8.7M |    	  114.40 |           576 |      1 |
| **Debug** |  Sync | Simple |       1.8B |    	    0.57 |	           0 |      0 |
| **Debug** |  Sync | Format |       8.2M |    	  121.70 |	         576 |      1 |
|  **Info** | Async | Simple |       4.1M |       245.60 |	         104 |      2 |
|  **Info** | Async | Format |       3.0M |       328.20 |	         734 |      4 |
|  **Info** |  Sync | Simple |       9.0M |    	  111.40 | 	       	  24 |      1 |
|  **Info** |  Sync | Format |       4.0M |       249.40 |	         606 |      3 |
|  **Warn** | Async | Simple |       5.1M |    	  196.60 |        	 104 |      2 |
|  **Warn** | Async | Format |       2.8M |       354.10 |        	 734 |      4 |
|  **Warn** |  Sync | Simple |       9.3M |       107.10 |        	  24 |      1 |
|  **Warn** |  Sync | Format |       3.7M |    	  270.30 |        	 606 |      3 |
| **Error** | Async | Simple |       4.7M |       213.60 |        	 120 |      2 |
| **Error** | Async | Format |       2.4M |       409.80 |        	 734 |      4 |
| **Error** |  Sync | Simple |       9.0M |       111.80 |        	  24 |      1 |
| **Error** |  Sync | Format |       3.8M |       265.70 |        	 606 |      3 |

### Single Thread

|   Level   |  Mode | Format | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-----------|-------|--------|------------|--------------|---------------|--------|
| **Debug** | Async | Simple |     317.8M |         3.78 |             0 |      0 |
| **Debug**	| Async | Format |       8.6M |       136.80 |	         576 |      1 |
| **Debug**	|  Sync | Simple |     315.0M |         3.87 |	           0 |      0 |
| **Debug**	|  Sync | Format |       8.4M |       143.10 |	         576 |      1 |
|  **Info**	| Async | Simple |       1.8M |       641.30 |	         104 |      2 |
|  **Info**	| Async | Format |       1.3M |       940.60 |	         734 |      4 |
|  **Info**	|  Sync | Simple |       2.8M |       443.80 |	          24 |      1 |
|  **Info**	|  Sync | Format |       1.6M |       711.80 |	         606 |      3 |
|  **Warn**	| Async | Simple |       1.8M |       644.80 |	         104 |      2 |
|  **Warn**	| Async | Format |       1.3M |       938.10 |	         734 |      4 |
|  **Warn**	|  Sync | Simple |       2.8M |       444.30 |	          24 |      1 |
|  **Warn**	|  Sync | Format |       1.6M |       709.90 |	         606 |      3 |
| **Error**	| Async | Simple |       1.8M |       664.30 |	         120 |      2 |
| **Error**	| Async | Format |       1.3M |       925.90 |	         734 |      4 |
| **Error**	|  Sync | Simple |       2.8M |       424.30 |	          24 |      1 |
| **Error**	|  Sync | Format |       1.6M |       767.00 |	         606 |	    3 |

*Benchmarked on Intel Core i9-9880H (2.30 GHz)*

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
