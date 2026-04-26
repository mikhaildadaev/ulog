---
outline: deep
---

# API

## Installation
```go
go get github.com/mikhaildadaev/ulog
```

## Navigation
- **Core**
    - [Main](/core_main-examples) — Telemetry setup, configuration, and standard logger adapter.
    - [Options](/core_options-examples) — All configuration options: Extractor, Format, Level, Mode, Theme.
    - [Reference](/core_reference-examples) — All data types and 16 field constructors.
- **FileSink**
    - [Main](/sinkfile_main-examples) — Atomic file rotation with gzip compression.
    - [Params](/sinkfile_params-examples) — Configuration: MaxSize, MaxBackups, MaxAge.
- **HttpSink**
    - [Main](/sinkhttp_main-examples) — HTTP delivery.
    - [Factories](/sinkhttp_factories-examples) — Ready-to-use integrations: Telegram, Discord, Slack, Loki, Kafka, Prometheus, Tempo, WeChat.
    - [Params](/sinkhttp_params-examples) — Configuration: Batching, Circuit Breaker, Dedup, Sampling, Retry.

## Features
- **Unified API** — One API for logs, metrics, and traces.
- **Context Extraction** — Automatic `node_id`, `trace_id`, etc. from `context.Context`.
- **16 Field Types** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **File Sink** — Non-blocking atomic rotation with gzip.
- **HTTP Sink** — Batching, Circuit Breaker, Deduplication, Retry, Sampling.
- **8 Integrations** — Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.