---
outline: deep
---

# API / HTTP 接收器 / 工厂

::: info 关于
开箱即用的 `Discord`、`Kafka`、`Loki`、`Prometheus`、`Slack`、`Telegram`、`Tempo`、`WeChat` 工厂。每个工厂都是一个预配置的 `SinkHttp`，具有正确的格式化器、请求头和过滤器。
:::

## NewSinkDiscord
通过 Webhook 将错误日志发送到 Discord 频道。
```go
sinkDiscord := ulog.NewSinkDiscord("https://discord.com/api/webhooks/...", "MyBot", "")
defer sinkDiscord.Close()
```
## NewSinkKafka
通过 REST Proxy 将日志发送到 Apache Kafka 进行流处理。
```go
sinkKafka := ulog.NewSinkKafka("http://kafka-rest:8082", "logs")
defer sinkKafka.Close()
```
## NewSinkLoki
将日志发送到 Grafana Loki 进行存储和查询。
```go
sinkLoki := ulog.NewSinkLoki("http://loki:3100",map[string]string{"app": "myapp", "env": "production"})
defer sinkLoki.Close()
```
## NewSinkPrometheus
以 Exposition 格式将指标发送到 Prometheus。
```go
sinkPrometheus := ulog.NewSinkPrometheus("http://pushgateway:9091")
defer sinkPrometheus.Close()
```
## NewSinkSlack
通过 Webhook 将错误日志发送到 Slack 频道。
```go
sinkSlack := ulog.NewSinkSlack("https://hooks.slack.com/services/...", "ULog", ":robot:", "", "#alerts")
defer sinkSlack.Close()
```
## NewSinkTelegram
通过 Bot API 将错误日志发送到 Telegram 聊天。
```go
sinkTelegram := ulog.NewSinkTelegram("botToken", "chatID")
defer sinkTelegram.Close()
```
## NewSinkTempo
将追踪数据发送到 Grafana Tempo 进行分布式追踪。
```go
sinkTempo := ulog.NewSinkTempo("http://tempo:4318")
defer sinkTempo.Close()
```
## NewSinkWechat
通过 Webhook 将错误日志发送到企业微信。
```go
sinkWechat := ulog.NewSinkWechat("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...")
defer sinkWechat.Close()
```