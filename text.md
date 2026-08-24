Features

Observability 2.0 — One API for Logs, Metrics, and Traces
Circuit Breaker for HTTP delivery
Atomic file rotation with gzip (non-blocking)
8 ready integrations: Discord, Kafka, Loki, Prometheus, Slack, Telegram, Tempo, WeChat
Automatic context extraction (WithExtractor)
Zero-allocation time caching for JSON formatter
TeeSink for multi-output routing
Performance

Core formatting + context: ~370 ns/op
File write with rotation: 5,500 ns/op
HTTP delivery: 27,000 ns/op
Documentation

Full documentation site: https://mikhaildadaev.github.io/ulog/
3 languages: English, Русский, 简体中文
11 pages with working code examples
CI/CD

Tests on Go 1.22-1.26 × macOS, Ubuntu, Windows
Automatic site deployment via GitHub Actions