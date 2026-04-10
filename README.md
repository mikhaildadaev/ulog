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
- **Structured fields** – type‑safe fields: `String()`, `Int()`, `Time()`, `Err()`, etc.;
- **Fully configurable** – functional options (`WithLevel`, `WithTheme`, `WithMode`, …);
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
- ulog.Err(err error) Field
- ulog.Errs(errs []error) Field
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

...

*Benchmarked on Intel Core i9-9880H (2.30 GHz)*

## Usage

```go
import (
    "fmt"
    "log"
    "github.com/mikhaildadaev/ulog"
)

func main() {
    // Basic logger
    logger := ulog.NewLogger(
        ulog.WithMode(ulog.ModeSync, os.Stdout),
        ulog.WithFormat(ulog.FormatText),
        ulog.WithTheme(ulog.ThemeDark),
    )
    logger.Debug("debugging connection")
    logger.Info("server started", ulog.Int("port", 8080))
    logger.Warn("high latency", ulog.Duration("latency", 150*time.Millisecond))
    logger.Error("database error", ulog.Err(nil))
    // Logger with context (auto‑extract trace_id)
    ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
    logger.InfoWithContext(ctx, "request processed", ulog.String("path", "/api/user"))
    // Async logger
    async := ulog.NewLogger(
        ulog.WithMode(ulog.ModeAsync, os.Stdout, 10000),
        ulog.WithFormat(ulog.FormatJson),
    )
    defer async.Close()
    async.Info("async message")
    async.Sync() // wait for flush
    // Standard log.Logger adapter
    std := ulog.NewLoggerLog(ulog.LevelError, logger)
    std.Print("error from standard logger")
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
