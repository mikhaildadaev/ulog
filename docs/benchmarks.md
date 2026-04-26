---
outline: deep
---

# Benchmarks

::: warning
This page is under development
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

## File Write with Rotation

Real-world benchmark writing structured JSON logs to a **real file** with **atomic rotation** enabled.

| Thread     | Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|------------|-------|------------|--------------|---------------|--------|
| **Multi**  | Async |       1.0M |        6,900 |          1962 |      6 |
| **Multi**  |  Sync |     152.7K |        7,800 |          1801 |      5 |
| **Single** | Async |     969.7K |        6,000 |          1962 |      6 |
| **Single** |  Sync |     234.4K |        5,500 |          1798 |      5 |

> **Note:** Includes full overhead: JSON formatting, context extraction, file I/O, and non-blocking rotation checks. `Single Sync` is the recommended production configuration.

---

## HTTP Write Overhead

Benchmark measuring HTTP sink overhead using `httptest.Server` (no network latency).

| Thread     | Mode  | Operations | Time (ns/op) | Memory (B/op) | Allocs |
|------------|-------|------------|--------------|---------------|--------|
| **Multi**  | Async |       1.0M |       27,000 |         8,400 |     82 |
| **Multi**  |  Sync |      45.4K |       26,400 |         9,100 |     89 |
| **Single** | Async |     555.2K |       42,100 |         9,100 |     82 |
| **Single** |  Sync |      13.6K |       82,500 |         9,400 |     85 |

> **Note:** Real-world latency will be dominated by network I/O (typically 10-100x higher). These numbers reflect `ulog` internal overhead only.

---

## Comparison: File Write with Rotation

| Library              | Time (ns/op)  | Notes                     |
|----------------------|---------------|---------------------------|
| **ulog**             | **5,500**     | Built-in atomic rotation  |
| Zap + lumberjack     | ~7,000-10,000 | External library required |
| Zerolog + lumberjack | ~6,000-9,000  | External library required |

> **ulog is faster because rotation is built into the core, eliminating data copying between libraries.**