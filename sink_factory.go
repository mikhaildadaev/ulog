package ulog

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Публичные структуры
type DiscordData struct {
	AvatarURL string `json:"avatar_url,omitempty"`
	Content   string `json:"content,omitempty"`
	TTS       bool   `json:"tts,omitempty"`
	UserName  string `json:"username,omitempty"`
}
type DiscordSink = HttpSink
type LokiData struct {
	Streams []struct {
		Stream map[string]string `json:"stream"`
		Values [][]string        `json:"values"`
	} `json:"streams"`
}
type LokiSink = HttpSink
type PrometheusData struct {
	Labels map[string]string `json:"labels,omitempty"`
	Name   string            `json:"name"`
	Value  float64           `json:"value"`
}
type PrometheusSink = HttpSink
type SlackData struct {
	Channel   string `json:"channel,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	Text      string `json:"text"`
	UserName  string `json:"username,omitempty"`
}
type SlackSink = HttpSink
type TelegramData struct {
	ChatID              string `json:"chat_id"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
	Text                string `json:"text"`
	ParseMode           string `json:"parse_mode,omitempty"`
}
type TelegramSink = HttpSink
type TempoData struct {
	Attributes map[string]any `json:"attributes,omitempty"`
	Duration   int64          `json:"duration_ms"`
	Name       string         `json:"name"`
	SpanID     string         `json:"span_id"`
	Timestamp  time.Time      `json:"timestamp"`
	TraceID    string         `json:"trace_id"`
}
type TempoSink = HttpSink
type WechatData struct {
	Content             string   `json:"content"`
	MsgType             string   `json:"msgtype"`
	MentionedList       []string `json:"mentioned_list,omitempty"`
	MentionedMobileList []string `json:"mentioned_mobile_list,omitempty"`
}
type WechatSink = HttpSink

// Публичные конструкторы
func NewDiscordSink(endPoint, userName, avatarURL string, params ...httpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			discordData := DiscordData{
				AvatarURL: avatarURL,
				Content:   getLogMessage(fields),
				TTS:       false,
				UserName:  userName,
			}
			return json.Marshal(discordData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewLokiSink(endPoint string, labels map[string]string, params ...httpParams) *HttpSink {
	lokiFormatter := func(attributes writeAttributes, fields []Field) ([]byte, error) {
		message := getLogMessage(fields)
		level := getLevel(attributes.typeLevel)
		streamLabels := make(map[string]string)
		for k, v := range labels {
			streamLabels[k] = v
		}
		streamLabels["level"] = level
		if traceID := getTraceID(fields); traceID != "" {
			streamLabels["trace_id"] = traceID
		}
		extraFields := make(map[string]interface{})
		for _, f := range fields {
			if f.nameKey == "message" {
				continue
			}
			extraFields[f.nameKey] = getLogField(f)
		}
		var logLine string
		if len(extraFields) > 0 {
			extraJSON, _ := json.Marshal(extraFields)
			logLine = fmt.Sprintf("%s %s", message, string(extraJSON))
		} else {
			logLine = message
		}
		timestamp := strconv.FormatInt(time.Now().UnixNano(), 10)
		lokiData := LokiData{
			Streams: []struct {
				Stream map[string]string `json:"stream"`
				Values [][]string        `json:"values"`
			}{
				{
					Stream: streamLabels,
					Values: [][]string{{timestamp, logLine}},
				},
			},
		}
		return json.Marshal(lokiData)
	}
	fullURL := strings.TrimSuffix(endPoint, "/") + "/loki/api/v1/push"
	return NewHttpSink(fullURL, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFormatter(lokiFormatter),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewPrometheusSink(endPoint string, params ...httpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFilterData(DataMetric),
		WithHttpFormatter(func(attrs writeAttributes, fields []Field) ([]byte, error) {
			var builder strings.Builder
			name, value, labels := getMetricData(fields)
			builder.WriteString(name)
			for k, v := range labels {
				builder.WriteByte(',')
				builder.WriteString(k)
				builder.WriteString("=\"")
				builder.WriteString(v)
				builder.WriteByte('"')
			}
			builder.WriteByte(' ')
			builder.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
			builder.WriteByte('\n')
			return []byte(builder.String()), nil
		}),
		WithHttpHeader("Content-Type", "text/plain"),
	}, params...)...)
}
func NewSlackSink(endPoint, userName, iconEmoji, iconURL, channel string, params ...httpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			slackData := SlackData{
				Channel:   channel,
				IconEmoji: iconEmoji,
				IconURL:   iconURL,
				Text:      getLogMessage(fields),
				UserName:  userName,
			}
			return json.Marshal(slackData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewTelegramSink(botToken, chatID string, params ...httpParams) *HttpSink {
	endPoint := "https://api.telegram.org/bot" + botToken + "/sendMessage"
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			telegramData := TelegramData{
				ChatID:    chatID,
				Text:      getLogMessage(fields),
				ParseMode: "HTML",
			}
			return json.Marshal(telegramData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewTempoSink(endPoint string, params ...httpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFilterData(DataTrace),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			tempoData := TempoData{
				Duration:  getTraceDuration(fields),
				Name:      getTraceName(fields),
				SpanID:    getTraceSpanID(fields),
				Timestamp: time.Now(),
				TraceID:   getTraceID(fields),
			}
			return json.Marshal(tempoData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
	}, params...)...)
}
func NewWechatSink(endPoint string, params ...httpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFilterData(DataLog),
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			wechatData := WechatData{
				Content: getLogMessage(fields),
				MsgType: "markdown",
			}
			return json.Marshal(wechatData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
