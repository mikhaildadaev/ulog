// Этот файл (sink_factory.go) находится в стадии активной разработки.
// API может изменяться
package ulog

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Публичные структуры
type DiscordData struct {
	AvatarURL string `json:"avatar_url,omitempty"`
	Content   string `json:"content,omitempty"`
	TTS       bool   `json:"tts,omitempty"`
	UserName  string `json:"username,omitempty"`
}
type DiscordSink = HttpSink
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

// Публичные конструкторы
func NewDiscordSink(endPoint, userName, avatarURL string, params ...httpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			discordData := DiscordData{
				AvatarURL: avatarURL,
				Content:   getMessage(fields),
				TTS:       false,
				UserName:  userName,
			}
			return json.Marshal(discordData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewPrometheusSink(endPoint string, params ...httpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFilterData(DataMetric),
		WithHttpFormatter(func(attrs writeAttributes, fields []Field) ([]byte, error) {
			var builder strings.Builder
			name, value, labels := getMetric(fields)
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
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			slackData := SlackData{
				Channel:   channel,
				IconEmoji: iconEmoji,
				IconURL:   iconURL,
				Text:      getMessage(fields),
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
		WithHttpFilterLevel(LevelError),
		WithHttpFormatter(func(attributes writeAttributes, fields []Field) ([]byte, error) {
			telegramData := TelegramData{
				ChatID:    chatID,
				Text:      getMessage(fields),
				ParseMode: "HTML",
			}
			return json.Marshal(telegramData)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	}, params...)...)
}
