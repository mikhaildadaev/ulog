---
outline: deep
---

# Go

::: info Information
The latest stable version of `ulog` is **v1.26.12**.
:::

## Get Started
```bash
go get github.com/mikhaildadaev/ulog
```

## Quick Navigation

- [Benchmarks](/en/benchmarks) - Performance data for core, fileSink and httpSink operations.
- **API**
    - **Core**
        - [Main](/en/core_main-examples) — Telemetry setup, configuration, and standard logger adapter.
        - [Formats](/en/core_formats-examples) — All data types and 16 field constructors.
        - [Options](/en/core_options-examples) — All configuration options: Extractor, Format, Level, Mode, Theme.
    - **SinkFile**
        - [Main](/en/sinkfile_main-examples) — Atomic file rotation with gzip compression.
        - [Params](/en/sinkfile_params-examples) — Configuration: MaxAge, MaxBackups, MaxSize.
    - **SinkHttp**
        - [Main](/en/sinkhttp_main-examples) — HTTP delivery.
        - [Factories](/en/sinkhttp_factories-examples) — Ready-to-use integrations: Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.
        - [Params](/en/sinkhttp_params-examples) — Configuration: Batching, Circuit Breaker, Dedup, Retry, Sampling.

## Key Features

- **Unified API** — One API for logs, metrics, and traces.
- **Context Extraction** — Automatic `node_id`, `trace_id`, etc. from `context.Context`.
- **16 Field Types** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **SinkFile** — Non-blocking atomic rotation with gzip.
- **SinkHttp** — Batching, Circuit Breaker, Deduplication, Retry, Sampling.
- **8 Integrations** — Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.
