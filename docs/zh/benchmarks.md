---
outline: deep
---

# 基准

::: info 资料
比较日志记录库的最佳方法是在**您自己的环境**与**您自己的工作负载**中运行基准测试。 每个项目都有独特的需求--延迟、吞吐量、内存、集成复杂性--没有一个基准可以全部捕获它们。

我建议您与其他库一起测试`ulog'，并选择最适合您需求的工具。
:::

## Core Performance

这些基准通过写入 `io.Discard`.

### MultiThread

| 模式  | 水平                  | 运作        | 时间 (ns/op)  | 记忆 (B/op)   | 分配    |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **DebugWithContext** |       5.8M |        180.7 |           536 |      3 |
| Async | **ErrorWithContext** |       2.0M |        578.3 |          1922 |      6 |
| Async | **InfoWithContext**  |       2.3M |        555.9 |          1922 |      6 |
| Async | **WarnWithContext**  |       2.4M |        470.7 |          1922 |      6 |
| Sync  | **DebugWithContext** |       6.3M |        203.3 |           536 |      3 |
| Sync  | **ErrorWithContext** |       3.2M |        372.1 |          1794 |      5 |
| Sync  | **InfoWithContext**  |       3.7M |        326.7 |          1794 |      5 |
| Sync  | **WarnWithContext**  |       4.0M |        299.9 |          1794 |      5 |

### SingleThread

| 模式  | 水平                  | 运作        | 时间  (ns/op) | 记忆 (B/op)   | 分配    |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **DebugWithContext** |       2.1M |        567.1 |           536 |      3 |
| Async | **ErrorWithContext** |       1.0M |         1045 |          1922 |      6 |
| Async | **InfoWithContext**  |       1.0M |         1006 |          1922 |      6 |
| Async | **WarnWithContext**  |       1.2M |        953.6 |          1922 |      6 |
| Sync  | **DebugWithContext** |       2.1M |        562.6 |           536 |      3 |
| Sync  | **ErrorWithContext** |       1.4M |        875.1 |          1794 |      5 |
| Sync  | **InfoWithContext**  |       1.5M |        810.0 |          1794 |      5 |
| Sync  | **WarnWithContext**  |       1.5M |        790.5 |          1794 |      5 |

::: tip 注
使用 `WithExtractor("node_id", "trace_id")` 自动从上下文中提取。 所有测试都写入 `io.Discard`。 以英特尔酷睿i9-9880h(2.30GHz)。
:::

## FileSink Performance

基准数据将结构化JSON日志写入启用 **原子旋转** 的 **真实文件**。

### MultiThread

| 模式  | 水平                  | 运作        | 时间 (ns/op) | 记忆 (B/op)    | 分配    |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     999.9K |        6,900 |          1962 |      6 |
| Sync  | **AllSupportLevels** |     152.7K |        7,800 |          1801 |      5 |

### SingleThread

| 模式  | 水平                  | 运作        | 时间 (ns/op) | 记忆 (B/op)    | 分配    |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     969.7K |        6,000 |          1962 |      6 |
| Sync  | **AllSupportLevels** |     234.4K |        5,500 |          1798 |      5 |

::: tip 注
`Single Sync` 是推荐的工作配置。 使用的基准测试包括所有附加功能：JSON格式，上下文提取，文件I/O和非阻塞旋转验证。
:::

## HttpSink Performance

使用 `httptest.Server` 测量 `ulog` HTTP接收器内部成本的基准数据。服务器'没有网络延迟。

### MultiThread

| 模式  | 水平                  | 运作        | 时间 (ns/op) | 记忆 (B/op)    | 分配    |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     999.9K |       27,000 |         8,400 |     82 |
| Sync  | **AllSupportLevels** |      45.4K |       26,400 |         9,100 |     89 |

### SingleThread

| 模式  | 水平                  | 运作        | 时间 (ns/op) | 记忆 (B/op)    | 分配    |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     555.2K |       42,100 |         9,100 |     82 |
| Sync  | **AllSupportLevels** |      13.6K |       82,500 |         9,400 |     85 |

::: tip 注
在真实环境中，延迟主要由网络I/O决定（通常高出10-100倍）。 这些数字只反映了`ulog`的内部成本。
:::
