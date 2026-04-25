---
outline: deep
---

# API

ULOG provides a unified API for **Logs**, **Metrics**, and **Traces**.

## Quick Navigation

- **Core**
    - [Main](/core_main-examples) — Telemetry setup, configuration, and standard logger adapter.
    - [Options](/core_options-examples) — All formats, levels, modes and switching.
    - [Reference](/core_reference-examples) — All formats, levels, modes and switching.
- **Sink**
    - [File](/sink_file-examples) — Atomic file rotation with gzip compression.
    - [HTTP](/sink_http-examples) — HTTP delivery with Circuit Breaker, retry, and batching.

## Key Features

- **Unified API** — One API for logs, metrics, and traces.
- **Context Extraction** — Automatic `node_id`, `trace_id`, etc. from `context.Context`.
- **16 Field Types** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **File Sink** — Non-blocking atomic rotation with gzip.
- **HTTP Sink** — Batching, Circuit Breaker, Deduplication, Retry, Sampling.
- **8 Integrations** — Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.