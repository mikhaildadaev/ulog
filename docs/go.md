---
outline: deep
---

# Go

## Get Started
```go
go get github.com/mikhaildadaev/ulog
```

## Quick Navigation

- [Benchmarks](/benchmarks) - Performance data for core, fileSink and httpSink operations.
- **API**
    - **Core**
        - [Main](/core_main-examples) — Telemetry setup, configuration, and standard logger adapter.
        - [Options](/core_options-examples) — All configuration options: Extractor, Format, Level, Mode, Theme.
        - [Reference](/core_reference-examples) — All data types and 16 field constructors.
    - **FileSink**
        - [Main](/sinkfile_main-examples) — Atomic file rotation with gzip compression.
        - [Params](/sinkfile_params-examples) — Configuration: MaxAge, MaxBackups, MaxSize.
    - **HttpSink**
        - [Main](/sinkhttp_main-examples) — HTTP delivery.
        - [Factories](/sinkhttp_factories-examples) — Ready-to-use integrations: Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.
        - [Params](/sinkhttp_params-examples) — Configuration: Batching, Circuit Breaker, Dedup, Retry, Sampling.

## Key Features

- **Unified API** — One API for logs, metrics, and traces.
- **Context Extraction** — Automatic `node_id`, `trace_id`, etc. from `context.Context`.
- **16 Field Types** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **File Sink** — Non-blocking atomic rotation with gzip.
- **HTTP Sink** — Batching, Circuit Breaker, Deduplication, Retry, Sampling.
- **8 Integrations** — Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.
