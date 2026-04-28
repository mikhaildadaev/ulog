---
outline: deep
---

# 基准

::: tip Note
比较日志记录库的最佳方法是在**您自己的环境**与**您自己的工作负载**中运行基准测试。 每个项目都有独特的需求--延迟、吞吐量、内存、集成复杂性--没有一个基准可以全部捕获它们。

我建议您与其他库一起测试`ulog'，并选择最适合您需求的工具。
:::

## Core Performance

These benchmarks measure **pure formatting and context extraction overhead** by writing to `io.Discard`.

### 多线程

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

### 单读,单读

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

> **Note:** Benchmarks use `WithExtractor("node_id", "trace_id")` to automatically extract from context. All benchmarks write to `io.Discard`. Benchmarked on Intel Core i9-9880H (2.30 GHz).

---

## FileSink Write Performance

Real-world benchmark writing structured JSON logs to a **real file** with **atomic rotation** enabled.

### 多线程

| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     999.9K |        6,900 |          1962 |      6 |
|  Sync | **AllSupportLevels** |     152.7K |        7,800 |          1801 |      5 |

### 单读,单读

| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     969.7K |        6,000 |          1962 |      6 |
|  Sync | **AllSupportLevels** |     234.4K |        5,500 |          1798 |      5 |

> **Note:** Includes full overhead: JSON formatting, context extraction, file I/O, and non-blocking rotation checks. `Single Sync` is the recommended production configuration.

---

## HttpSink Write Performance

Benchmark measuring HTTP sink overhead using `httptest.Server` (no network latency).

### 多线程

| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     999.9K |       27,000 |         8,400 |     82 |
|  Sync | **AllSupportLevels** |      45.4K |       26,400 |         9,100 |     89 |

### 单读,单读

| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     555.2K |       42,100 |         9,100 |     82 |
|  Sync | **AllSupportLevels** |      13.6K |       82,500 |         9,400 |     85 |

> **Note:** Real-world latency will be dominated by network I/O (typically 10-100x higher). These numbers reflect `ulog` internal overhead only.
