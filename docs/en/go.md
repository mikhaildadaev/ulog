---
outline: deep
---

# Go

::: info Info
The latest stable version of `ulog` is **v1.26.12**.
:::

## Get Started
```bash
go get github.com/mikhaildadaev/ulog
```

## Get Test 
```bash
go test ./...
go test -bench=. ./...
go test -cover ./...
go test -race ./...
```

## Key Features
- **Unified API** — One API for logs, metrics, and traces.
- **Context Extraction** — Automatic extraction `node_id`, `trace_id`, etc. from `context.Context`.
- **16 Field Types** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **SinkFile** — Non-blocking atomic rotation with `gzip`.
- **SinkHttp** — `Batching`, `Circuit Breaker`, `Deduplication`, `Retry`, `Sampling`.
- **8 Integrations** — `Discord`, `Kafka`, `Loki`, `Prometheus`, `Slack`, `Telegram`, `Tempo`, `WeChat`.

## Quick Navigation
- [Benchmarks](/en/benchmarks) - Core, file, and network performance data.
- **API**
    - **Core**
        - [Main](/en/core_main-examples) — Telemetry setup, configuration, and standard logger adapter.
        - [Options](/en/core_options-examples) — All configuration options: Extractor, Formats, Levels, Modes, Themes.
        - [Types](/en/core_types-examples) — All data types and 16 field constructors.
    - **SinkFile**
        - [Main](/en/sinkfile_main-examples) — Creating a file sink and basic setup.
        - [Params](/en/sinkfile_params-examples) — Rotation and compression config: max size, age, backup count.
    - **SinkHttp**
        - [Main](/en/sinkhttp_main-examples) — Creating an http sink and basic setup.
        - [Factories](/en/sinkhttp_factories-examples) — 8 ready-to-use integration factories.
        - [Params](/en/sinkhttp_params-examples) — Delivery config: batching, deduplication, retry, sampling, circuit breaker.

