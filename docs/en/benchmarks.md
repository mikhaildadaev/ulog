---
outline: deep
---

# Benchmarks

::: info Information
The best way to compare libraries is to run benchmarks in **your own environment** with **your own workload**. Each project has unique requirements — latency, throughput, memory usage, and integration complexity — and no single test can cover them all.

I recommend that you test `ulog` alongside other libraries and choose the tool that best suits your needs.
:::

## Core Write Performance

These benchmarks measure the **cost of formatting and extracting context** by writing to `io.Discard`.

### Multithread

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

### Singlethread

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

::: tip Note
Uses `WithExtractor("node_id", "trace_id")` to automatically extract from the context. All tests write to `io.Discard`. Benchmarked on Intel Core i9-9880H (2.30GHz).
:::

## FileSink Write Performance

Benchmark data writes structured JSON logs to a **real file** with **atomic rotation** enabled.

### Multithread

| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     999.9K |        6,900 |          1962 |      6 |
|  Sync | **AllSupportLevels** |     152.7K |        7,800 |          1801 |      5 |

### Singlethread

| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     969.7K |        6,000 |          1962 |      6 |
|  Sync | **AllSupportLevels** |     234.4K |        5,500 |          1798 |      5 |

::: tip Note
`Single Sync` is the recommended working configuration. The benchmarks used include all additional features: JSON formatting, context extraction, file I/O, and non-blocking rotation verification.
:::

## HttpSink Write Performance

Benchmark data that measures the internal costs of the `ulog` HTTP receiver using `httptest.Server` without network latency.

### Multithread

| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     999.9K |       27,000 |         8,400 |     82 |
|  Sync | **AllSupportLevels** |      45.4K |       26,400 |         9,100 |     89 |

### Singlethread

| Mode  | Level                | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|----------------------|------------|--------------|---------------|--------|
| Async | **AllSupportLevels** |     555.2K |       42,100 |         9,100 |     82 |
|  Sync | **AllSupportLevels** |      13.6K |       82,500 |         9,400 |     85 |

::: tip Note
In a real environment, the delay is mainly determined by network I/O (usually 10-100 times higher). These numbers only reflect the internal costs of `ulog`.
:::.
