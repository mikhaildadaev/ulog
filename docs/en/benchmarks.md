---
outline: deep
---

# Benchmarks
::: info **Info**
The best way to compare libraries is to run benchmarks in **your own environment** with **your own workload**. Each project has unique requirements — latency, throughput, memory usage, and integration complexity — and no single test can cover them all.

I recommend that you test `ulog` alongside other libraries and choose the tool that best suits your needs.
:::

## Core Performance
These benchmarks measure the **cost of formatting and extracting context** by writing to `io.Discard`.

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

::: tip **Note**
Uses `WithExtractor("node_id", "trace_id")` to automatically extract from context. All tests write to `io.Discard` (equivalent to `/dev/null` on Unix or `NUL` on Windows). This measures only the logging overhead (field formatting, JSON encoding, context extraction) without disk or network I/O. Real-world performance will depend on your output destination (file, network, etc.). 

*Benchmarked on Intel Core i9-9880H (2.30 GHz).*
:::

## SinkFile Performance
Benchmark data writes structured JSON logs to a **real file** with **atomic rotation** enabled.

### MultiThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     212.2K |        7,784 |          1963 |      6 |
| Sync  | **AllSupportLevels** |     140.6K |        8,360 |          1800 |      5 |

### SingleThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     244.5K |        6,471 |          1965 |      6 |
| Sync  | **AllSupportLevels** |     200.8K |        6,390 |          1802 |      5 |

::: tip **Note**
Uses `WithExtractor("node_id", "trace_id")` to automatically extract from context. Writes structured JSON logs to a **real file** with **atomic rotation** enabled (`WithFileMaxSize(15)`). Includes full overhead: JSON formatting, context extraction, file I/O, and non-blocking rotation checks. 

*Benchmarked on Intel Core i9-9880H (2.30 GHz).*
:::

## SinkHttp Performance
Benchmark data that measures the internal costs of the `ulog` HTTP receiver using `httptest.Server` without network latency.

### MultiThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |      18.1K |       55,427 |         8,828 |     82 |
| Sync  | **AllSupportLevels** |      19.5K |       64,989 |        31,419 |     91 |

### SingleThread
| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |      13.1K |       76,392 |         9,179 |     82 |
| Sync  | **AllSupportLevels** |      14.4K |       81,589 |         8,940 |     85 |

::: tip **Note**
Uses `httptest.Server` to simulate HTTP endpoint. Measures full overhead: JSON formatting, context extraction, HTTP request/response. In a real environment, the delay is mainly determined by network I/O (usually 10-100 times higher). These numbers only reflect the internal costs of `ulog`. *Multi* benchmarks use `b.RunParallel` to simulate real-world concurrent load. 

*Benchmarked on Intel Core i9-9880H (2.30 GHz).*
:::
