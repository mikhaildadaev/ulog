---
outline: deep
---

# Go

::: info 资料
`ulog` 的最新稳定版本是 **v1.26.12**.
:::

## Get Started
```bash
go get github.com/mikhaildadaev/ulog
```

## Quick Navigation

- [基准](/en/benchmarks) - Performance data for core, fileSink and httpSink operations.
- **API**
    - **核心**
        - [主要](/en/core_main-examples) — Telemetry setup, configuration, and standard logger adapter.
        - [选项](/en/core_options-examples) — All configuration options: Extractor, Format, Level, Mode, Theme.
        - [类别](/en/core_types-examples) — All data types and 16 field constructors.
    - **写入文件**
        - [主要](/en/sinkfile_main-examples) — Atomic file rotation with gzip compression.
        - [帕拉姆斯](/en/sinkfile_params-examples) — Configuration: MaxAge, MaxBackups, MaxSize.
    - **通过网络录制**
        - [主要](/en/sinkhttp_main-examples) — HTTP delivery.
        - [工厂](/en/sinkhttp_factories-examples) — Ready-to-use integrations: Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.
        - [帕拉姆斯](/en/sinkhttp_params-examples) — Configuration: Batching, Circuit Breaker, Dedup, Retry, Sampling.

## Key Features

- **Unified API** — One API for logs, metrics, and traces.
- **Context Extraction** — Automatic extract `node_id`, `trace_id`, etc. from `context.Context`.
- **16 Field Types** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **写入文件** — Non-blocking atomic rotation with gzip.
- **通过网络录制** — Batching, Circuit Breaker, Deduplication, Retry, Sampling.
- **8 Integrations** — Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.
