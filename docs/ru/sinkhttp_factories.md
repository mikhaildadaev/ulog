---
outline: deep
---

# API / Запись по сети / Фабрики

::: info Информация
Готовые к использованию фабрики для `Discord`, `Kafka`, `Loki`, `Prometheus`, `Slack`, `Telegram`, `Tempo`, `WeChat`. Каждая фабрика — это предварительно настроенный `SinkHttp` с правильным форматером, заголовками и фильтрами.
:::

## NewSinkDiscord
Отправляет логи ошибок в канал Discord через вебхук.
```go
sinkDiscord := ulog.NewSinkDiscord("https://discord.com/api/webhooks/...", "MyBot", "")
defer sinkDiscord.Close()
```
## NewSinkKafka
Отправляет логи в Apache Kafka через REST Proxy для потоковой обработки.
```go
sinkKafka := ulog.NewSinkKafka("http://kafka-rest:8082", "logs")
defer sinkKafka.Close()
```
## NewSinkLoki
Отправляет логи в Grafana Loki для хранения и запросов.
```go
sinkLoki := ulog.NewSinkLoki("http://loki:3100",map[string]string{"app": "myapp", "env": "production"})
defer sinkLoki.Close()
```
## NewSinkPrometheus
Отправляет метрики в Prometheus в формате exposition.
```go
sinkPrometheus := ulog.NewSinkPrometheus("http://pushgateway:9091")
defer sinkPrometheus.Close()
```
## NewSinkSlack
Отправляет логи ошибок в канал Slack через вебхук.
```go
sinkSlack := ulog.NewSinkSlack("https://hooks.slack.com/services/...", "ULog", ":robot:", "", "#alerts")
defer sinkSlack.Close()
```
## NewSinkTelegram
Отправляет логи ошибок в чат Telegram через Bot API.
```go
sinkTelegram := ulog.NewSinkTelegram("botToken", "chatID")
defer sinkTelegram.Close()
```
## NewSinkTempo
Отправляет трейсы в Grafana Tempo для распределённой трассировки.
```go
sinkTempo := ulog.NewSinkTempo("http://tempo:4318")
defer sinkTempo.Close()
```
## NewSinkWechat
Отправляет логи ошибок в WeChat Work через вебхук.
```go
sinkWechat := ulog.NewSinkWechat("https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=...")
defer sinkWechat.Close()
```