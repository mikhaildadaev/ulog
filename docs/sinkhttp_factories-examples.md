# API / HttpSink / Factories

::: warning
This page is under development
:::

## NewDiscordSink
Sends error logs to a Discord channel via webhook.
```go
discordSink := ulog.NewDiscordSink("https://discord.com/api/webhooks/...", "MyBot", "")
defer discordSink.Close()
```
## NewKafkaSink
Sends logs to Apache Kafka via REST Proxy for stream processing.
```go
kafkaSink := ulog.NewKafkaSink("http://kafka-rest:8082", "logs")
defer kafkaSink.Close()
```
## NewLokiSink
Sends logs to Grafana Loki for storage and querying.
```go
lokiSink := ulog.NewLokiSink("http://loki:3100",map[string]string{"app": "myapp", "env": "production"})
defer lokiSink.Close()
```
## NewPrometheusSink
Sends metrics to Prometheus in exposition format.
```go
prometheusSink := ulog.NewPrometheusSink("http://prometheus:9091")
defer prometheusSink.Close()
```
## NewSlackSink
Sends error logs to a Slack channel via webhook.
```go
slackSink := ulog.NewSlackSink("https://hooks.slack.com/services/...", "ULog", ":robot:", "", "#alerts")
defer slackSink.Close()
```
## NewTelegramSink
Sends error logs to a Telegram chat via bot API.
```go
telegramSink := ulog.NewTelegramSink("botToken", "chatID")
defer telegramSink.Close()
```
## NewTempoSink
Sends traces to Grafana Tempo for distributed tracing.
```go
tempoSink := ulog.NewTempoSink("http://tempo:4317")
defer tempoSink.Close()
```
## NewWechatSink
Sends error logs to WeChat Work (企业微信) via webhook.
```go
wechatSink := ulog.NewWechatSink("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...")
defer wechatSink.Close()
```