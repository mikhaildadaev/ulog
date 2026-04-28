---
outline: deep
---

# Go

::: info Информация
Последняя стабильная версия `ulog` — **v1.26.12**.
:::

## Быстрый старт
```bash
go get github.com/mikhaildadaev/ulog
```

## Быстрая навигация

- [Бенчмарки](/ru/benchmarks) - Performance data for core, sinkFile and sinkHttp operations.
- **API**
    - **Core**
        - [Основное](/ru/core_main-examples) — Telemetry setup, configuration, and standard logger adapter.
        - [Опции](/ru/core_options-examples) — All configuration options: Extractor, Format, Level, Mode, Theme.
        - [Типы](/en/core_types-examples) — All data types and 16 field constructors.
    - **SinkFile**
        - [Основное](/ru/sinkfile_main-examples) — Atomic file rotation with gzip compression.
        - [Параметры](/ru/sinkfile_params-examples) — Configuration: MaxAge, MaxBackups, MaxSize.
    - **SinkHttp**
        - [Основное](/ru/sinkhttp_main-examples) — HTTP delivery.
        - [Фабрики](/ru/sinkhttp_factories-examples) — Ready-to-use integrations: Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.
        - [Параметры](/ru/sinkhttp_params-examples) — Configuration: Batching, Circuit Breaker, Dedup, Retry, Sampling.

## Ключевые функции

- **Unified API** — One API for logs, metrics, and traces.
- **Context Extraction** — Automatic `node_id`, `trace_id`, etc. from `context.Context`.
- **16 Field Types** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **SinkFile** — Non-blocking atomic rotation with gzip.
- **SinkHttp** — Batching, Circuit Breaker, Deduplication, Retry, Sampling.
- **8 Integrations** — Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat.
