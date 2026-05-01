---
outline: deep
---

# 基准

::: info **关于**
比较库的最佳方式是在**您自己的环境**中使用**您自己的工作负载**运行基准测试。每个项目都有独特的需求——延迟、吞吐量、内存使用和集成复杂性——没有任何单一的测试能够覆盖所有情况。

我建议您将 `ulog` 与其他库一起测试，并选择最适合您需求的工具。
:::

## Core Performance
这些基准测试通过写入 `io.Discard` 来测量**格式化和上下文提取的成本**。

### MultiThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
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
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **DebugWithContext** |       2.1M |        567.1 |           536 |      3 |
| Async | **ErrorWithContext** |       1.0M |         1045 |          1922 |      6 |
| Async | **InfoWithContext**  |       1.0M |         1006 |          1922 |      6 |
| Async | **WarnWithContext**  |       1.2M |        953.6 |          1922 |      6 |
| Sync  | **DebugWithContext** |       2.1M |        562.6 |           536 |      3 |
| Sync  | **ErrorWithContext** |       1.4M |        875.1 |          1794 |      5 |
| Sync  | **InfoWithContext**  |       1.5M |        810.0 |          1794 |      5 |
| Sync  | **WarnWithContext**  |       1.5M |        790.5 |          1794 |      5 |

::: tip **注**
使用 `WithExtractor("node_id", "trace_id")` 从上下文中自动提取。所有测试均写入 `io.Discard`（相当于 Unix 上的 `/dev/null` 或 Windows 上的 `NUL`）。这仅测量日志记录的开销（字段格式化、JSON 编码、上下文提取），不包括磁盘或网络 I/O。实际性能将取决于您的输出目标（文件、网络等）。

*在 Intel Core i9-9880H (2.30 GHz) 上测试。*
:::

## SinkFile Performance
基准数据将结构化JSON日志写入启用 **原子旋转** 的 **真实文件**。

### MultiThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     999.9K |        6,900 |          1962 |      6 |
| Sync  | **AllSupportLevels** |     152.7K |        7,800 |          1801 |      5 |

### SingleThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     969.7K |        6,000 |          1962 |      6 |
| Sync  | **AllSupportLevels** |     234.4K |        5,500 |          1798 |      5 |

::: tip **注**
使用 `WithExtractor("node_id", "trace_id")` 从上下文中自动提取。将结构化 JSON 日志写入**真实文件**，并启用**原子轮转**（`WithFileMaxSize(15)`）。包含完整开销：JSON 格式化、上下文提取、文件 I/O 和非阻塞轮转检查。

*在 Intel Core i9-9880H (2.30 GHz) 上测试。*
:::

## SinkHttp Performance
基准测试数据，使用 `httptest.Server` 测量 `ulog` HTTP 接收器的内部成本，不含网络延迟。

### MultiThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     999.9K |       27,000 |         8,400 |     82 |
| Sync  | **AllSupportLevels** |      45.4K |       26,400 |         9,100 |     89 |

### SingleThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     555.2K |       42,100 |         9,100 |     82 |
| Sync  | **AllSupportLevels** |      13.6K |       82,500 |         9,400 |     85 |

::: tip **注**
使用 `httptest.Server` 模拟 HTTP 端点。测量完整开销：JSON 格式化、上下文提取、HTTP 请求/响应。在真实环境中，延迟主要由网络 I/O 决定（通常高出 10–100 倍）。这些数字仅反映 `ulog` 的内部成本。*Multi* 基准测试使用 `b.RunParallel` 模拟真实的并发负载。

*在 Intel Core i9-9880H (2.30 GHz) 上测试。*
:::
