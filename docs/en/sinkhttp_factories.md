---
outline: deep
---

# API / SinkHttp / Factories

::: info Info
Ready-to-use factories for `Discord`, `Kafka`, `Loki`, `Prometheus`, `Slack`, `Telegram`, `Tempo`, `WeChat`. Each factory is a pre-configured `SinkHttp` with the right formatter, headers, and filters.
:::

## NewSinkDiscord
Sends error logs to a Discord channel via webhook.
```go
sinkDiscord := ulog.NewSinkDiscord("https://discord.com/api/webhooks/...", "MyBot", "")
defer sinkDiscord.Close()
```
## NewSinkKafka
Sends logs to Apache Kafka via REST Proxy for stream processing.
```go
sinkKafka := ulog.NewSinkKafka("http://kafka-rest:8082", "logs")
defer sinkKafka.Close()
```
## NewSinkLoki
Sends logs to Grafana Loki for storage and querying.
```go
sinkLoki := ulog.NewSinkLoki("http://loki:3100",map[string]string{"app": "myapp", "env": "production"})
defer sinkLoki.Close()
```
## NewSinkPrometheus
Sends metrics to Prometheus in exposition format.
```go
sinkPrometheus := ulog.NewSinkPrometheus("http://pushgateway:9091")
defer sinkPrometheus.Close()
```
## NewSinkSlack
Sends error logs to a Slack channel via webhook.
```go
sinkSlack := ulog.NewSinkSlack("https://hooks.slack.com/services/...", "ULog", ":robot:", "", "#alerts")
defer sinkSlack.Close()
```
## NewSinkTelegram
Sends error logs to a Telegram chat via bot API.
```go
sinkTelegram := ulog.NewSinkTelegram("botToken", "chatID")
defer sinkTelegram.Close()
```
## NewSinkTempo
Sends traces to Grafana Tempo for distributed tracing.
```go
sinkTempo := ulog.NewSinkTempo("http://tempo:4318")
defer sinkTempo.Close()
```
## NewSinkWechat
Sends error logs to WeChat Work via webhook.
```go
sinkWechat := ulog.NewSinkWechat("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...")
defer sinkWechat.Close()
```