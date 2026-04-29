---
outline: deep
---

# Go

::: info 关于
`ulog` 的最新稳定版本是 **v1.26.12**.
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
- **统一 API** — 日志、指标和追踪的单一 API。
- **上下文提取** — 从 `context.Context` 自动提取 `node_id`、`trace_id` 等。
- **16 种字段类型** — `Bool`、`Bools`、`Duration`、`Durations`、`Error`、`Errors`、`Float64`、`Floats64`、`Int`、`Ints`、`Int64`、`Ints64`、`String`、`Strings`、`Time`、`Times`。
- **文件写入** — 非阻塞原子轮转，支持 `gzip` 压缩。
- **网络写入** — `Batching`、`Circuit Breaker`、`Deduplication`、`Retry`、`Sampling`。
- **8 种集成** — `Discord`、`Kafka`、`Loki`、`Prometheus`、`Slack`、`Telegram`、`Tempo`、`WeChat`。

## Quick Navigation
- [基准测试](/en/benchmarks) - 核心、文件和网络的性能数据。
- **API**
    - **核心**
        - [主要](/en/core_main-examples) — 遥测设置、配置和标准日志适配器。
        - [选项](/en/core_options-examples) — 所有配置参数：提取器、格式、级别、模式、主题。
        - [类型](/en/core_types-examples) — 所有数据类型和 16 个字段构造函数。
    - **文件接收器**
        - [主要](/en/sinkfile_main-examples) — 创建文件接收器和基本设置。
        - [参数](/en/sinkfile_params-examples) — 轮转和压缩配置：最大大小、保留天数、备份数量。
    - **HTTP 接收器**
        - [主要](/en/sinkhttp_main-examples) — 创建 HTTP 接收器和基本设置。
        - [工厂](/en/sinkhttp_factories-examples) — 8 个开箱即用的集成工厂。
        - [参数](/en/sinkhttp_params-examples) — 发送配置：批处理、去重、重试、采样、断路器。
