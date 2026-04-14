// Этот файл (sink_discord.go) находится в стадии активной разработки.
// API может изменяться
package ulog

import (
	"encoding/json"
)

// Публичные структуры
type DiscordWebhook struct {
	AvatarURL string `json:"avatar_url,omitempty"`
	Content   string `json:"content,omitempty"`
	TTS       bool   `json:"tts,omitempty"`
	Username  string `json:"username,omitempty"`
}
type DiscordSink = HttpSink

// Публичные конструкторы
func NewDiscordSink(webhookURL string, username, avatarURL string, options ...HttpOption) *HttpSink {
	return NewHttpSink(webhookURL, append([]HttpOption{
		WithHttpFormatter(func(level TypeLevel, p []byte) ([]byte, error) {
			webhook := DiscordWebhook{
				AvatarURL: avatarURL,
				Content:   string(p),
				TTS:       false,
				Username:  username,
			}
			return json.Marshal(webhook)
		}),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpLevelMin(LevelError),
		WithHttpMethod("POST"),
	}, options...)...)
}
