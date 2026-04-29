---
outline: deep
---

# Go

::: info Инфо
Последняя стабильная версия `ulog` — **v1.26.12**.
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
- **Унифицированный API** — Единый API для логов, метрик и трейсов.
- **Извлечение контекста** — Автоматическое извлечение `node_id`, `trace_id` и т.д. из `context.Context`.
- **16 типов полей** — `Bool`, `Bools`, `Duration`, `Durations`, `Error`, `Errors`, `Float64`, `Floats64`, `Int`, `Ints`, `Int64`, `Ints64`, `String`, `Strings`, `Time`, `Times`.
- **Запись в файл** — Неблокирующая атомарная ротация со сжатием `gzip`.
- **Запись по сети** — `Batching`, `Circuit Breaker`, `Deduplication`, `Retry`, `Sampling`.
- **8 интеграций** — `Discord`, `Kafka`, `Loki`, `Prometheus`, `Slack`, `Telegram`, `Tempo`, `WeChat`.

## Quick Navigation
- [Бенчмарки](/ru/benchmarks) - Данные о производительности ядра, записи в файл и записи по сети.
- **API**
    - **Ядро**
        - [Основное](/ru/core_main-examples) — Настройка телеметрии, конфигурирование и стандартный адаптер регистратора.
        - [Опции](/ru/core_options-examples) — Все параметры конфигурации: Экстрактор, Форматы, Уровни, Режимы, Темы
        - [Типы](/en/core_types-examples) — Все типы данных и 16 конструкторов полей.
    - **Запись в файл**
        - [Основное](/ru/sinkfile_main-examples) — Создание файлового синка и базовая настройка.
        - [Параметры](/ru/sinkfile_params-examples) — Конфигурация ротации и сжатия: максимальный размер, возраст, количество бэкапов.
    - **Запись по сети**
        - [Основное](/ru/sinkhttp_main-examples) — Создание сетевого синка и базовая настройка.
        - [Фабрики](/ru/sinkhttp_factories-examples) — Готовые 8 фабрик интеграций из коробки.
        - [Параметры](/ru/sinkhttp_params-examples) — Конфигурация отправки: батчинг, дедупликация, повтор, выборка, автоматический выключатель.
