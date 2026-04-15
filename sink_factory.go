// Этот файл (sink_factory.go) находится в стадии активной разработки.
// API может изменяться
package ulog

import (
	"encoding/json"
)

// Публичные структуры
type DiscordData struct {
	AvatarURL string `json:"avatar_url,omitempty"`
	Content   string `json:"content,omitempty"`
	TTS       bool   `json:"tts,omitempty"`
	UserName  string `json:"username,omitempty"`
}
type DiscordSink = HttpSink
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
		WithHttpFormatter(func(attributes writeAttributes, p []byte) ([]byte, error) {
			webhook := DiscordData{
				AvatarURL: avatarURL,
				Content:   string(p),
				TTS:       false,
				UserName:  userName,
			}
			return json.Marshal(webhook)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpLevelMin(LevelError),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewSlackSink(endPoint, userName, iconEmoji, iconURL, channel string, params ...httpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFormatter(func(attributes writeAttributes, p []byte) ([]byte, error) {
			data := SlackData{
				Channel:   channel,
				IconEmoji: iconEmoji,
				IconURL:   iconURL,
				Text:      string(p),
				UserName:  userName,
			}
			return json.Marshal(data)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpLevelMin(LevelError),
		WithHttpMethod("POST"),
	}, params...)...)
}
func NewTelegramSink(botToken, chatID string, params ...httpParams) *HttpSink {
	endPoint := "https://api.telegram.org/bot" + botToken + "/sendMessage"
	return NewHttpSink(endPoint, append([]httpParams{
		WithHttpFormatter(func(attributes writeAttributes, p []byte) ([]byte, error) {
			data := TelegramData{
				ChatID:    chatID,
				Text:      string(p),
				ParseMode: "HTML",
			}
			return json.Marshal(data)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpLevelMin(LevelError),
		WithHttpMethod("POST"),
	}, params...)...)
}
