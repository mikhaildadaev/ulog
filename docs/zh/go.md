---
outline: deep
---

# Go
```bash
go get github.com/mikhaildadaev/ulog
```

::: info **关于**
`ulog` 的最新稳定版本是 **v1.26.11**.
:::

## Run Test 
```bash
go test ./...
go test -bench=. ./...
go test -cover ./...
go test -race ./...
```

## Key Features
- **统一 API** — 日志、指标和追踪的单一 API。
- **上下文提取** — 从 `context.Context` 自动提取 `node_id`、`trace_id` 等。
- **彩色输出** — `Dark` 和 `Light` 主题，TEXT 格式支持自动检测。
- **16 种字段类型** — `Bool`、`Bools`、`Duration`、`Durations`、`Error`、`Errors`、`Float64`、`Floats64`、`Int`、`Ints`、`Int64`、`Ints64`、`String`、`Strings`、`Time`、`Times`。
- **文件写入** — 非阻塞原子轮转，支持 `gzip` 压缩。
- **网络写入** — `Batching`、`Circuit Breaker`、`Deduplication`、`Retry`、`Sampling`。
- **8 种集成** — `Discord`、`Kafka`、`Loki`、`Prometheus`、`Slack`、`Telegram`、`Tempo`、`WeChat`。

## Limits
- **异步缓冲区**：缓冲区满时同步写入（无阻塞）
- **调用者信息**：仅 `LevelDebug` 级别可用（性能优化）
- **时间精度**：微秒（6 位数字）—— 满足 99% 的使用场景，减少内存分配
- **去重缓存**：仅存储在内存中，定期清理（重启后不持久化）
- **断路器**：应用重启后重置（无持久化状态）
- **文件轮转**：每次写入时检查大小；首次超过限制时触发轮转
- **HTTP 批处理**：应用在刷新前崩溃可能导致消息丢失
- **Kafka 接收器**：使用 REST Proxy API（非原生 Kafka 协议）—— 需要 Confluent REST Proxy
- **Loki 接收器**：使用 HTTP API (`/loki/api/v1/push`) —— 标签需预先配置
- **上下文提取**：仅适用于通过 `context.WithValue()` 存储的值
- **零依赖**：有意为之；像原生 Kafka 协议等功能不使用外部库
