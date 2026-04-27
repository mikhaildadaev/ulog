---
outline: deep
---

# Benchmarks

::: tip Note
The best way to compare logging libraries is to run benchmarks in **your own environment** with **your own workload**. Every project has unique requirements — latency, throughput, memory, integration complexity — and no single benchmark can capture them all.

I recommend that you to test `ulog` alongside other libraries and choose the tool that best fits your needs.
:::

## Core Write Performance

These benchmarks measure **pure formatting and context extraction overhead** by writing to `io.Discard`.

### Multi Thread

| Level                | Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|----------------------|-------|------------|--------------|---------------|--------|
| **DebugWithContext** | Async |       5.8M |        180.7 |           536 |      3 |
| **DebugWithContext** | Sync  |       6.3M |        203.3 |           536 |      3 |
| **ErrorWithContext** | Async |       2.0M |        578.3 |          1922 |      6 |
| **ErrorWithContext** | Sync  |       3.2M |        372.1 |          1794 |      5 |
| **InfoWithContext**  | Async |       2.3M |        555.9 |          1922 |      6 |
| **InfoWithContext**  | Sync  |       3.7M |        326.7 |          1794 |      5 |
| **WarnWithContext**  | Async |       2.4M |        470.7 |          1922 |      6 |
| **WarnWithContext**  | Sync  |       4.0M |        299.9 |          1794 |      5 |

### Single Thread

| Level                | Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|----------------------|-------|------------|--------------|---------------|--------|
| **DebugWithContext** | Async |       2.1M |        567.1 |           536 |      3 |
| **DebugWithContext** |  Sync |       2.1M |        562.6 |           536 |      3 |
| **ErrorWithContext** | Async |       1.0M |         1045 |          1922 |      6 |
| **ErrorWithContext** |  Sync |       1.4M |        875.1 |          1794 |      5 |
| **InfoWithContext**  | Async |       1.0M |         1006 |          1922 |      6 |
| **InfoWithContext**  |  Sync |       1.5M |        810.0 |          1794 |      5 |
| **WarnWithContext**  | Async |       1.2M |        953.6 |          1922 |      6 |
| **WarnWithContext**  |  Sync |       1.5M |        790.5 |          1794 |      5 |

> **Note:** Benchmarks use `WithExtractor("node_id", "trace_id")` to automatically extract from context. All benchmarks write to `io.Discard`. Benchmarked on Intel Core i9-9880H (2.30 GHz).

---

## File Write Performance

Real-world benchmark writing structured JSON logs to a **real file** with **atomic rotation** enabled.

### Multi Thread

| Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|------------|--------------|---------------|--------|
| Async |     999.9K |        6,900 |          1962 |      6 |
|  Sync |     152.7K |        7,800 |          1801 |      5 |

### Single Thread

| Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|------------|--------------|---------------|--------|
| Async |     969.7K |        6,000 |          1962 |      6 |
|  Sync |     234.4K |        5,500 |          1798 |      5 |

> **Note:** Includes full overhead: JSON formatting, context extraction, file I/O, and non-blocking rotation checks. `Single Sync` is the recommended production configuration.

---

## Http Write Performance

Benchmark measuring HTTP sink overhead using `httptest.Server` (no network latency).

### Multi Thread

| Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|------------|--------------|---------------|--------|
| Async |     999.9K |       27,000 |         8,400 |     82 |
|  Sync |      45.4K |       26,400 |         9,100 |     89 |

### Single Thread

| Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|-------|------------|--------------|---------------|--------|
| Async |     555.2K |       42,100 |         9,100 |     82 |
|  Sync |      13.6K |       82,500 |         9,400 |     85 |

> **Note:** Real-world latency will be dominated by network I/O (typically 10-100x higher). These numbers reflect `ulog` internal overhead only.
