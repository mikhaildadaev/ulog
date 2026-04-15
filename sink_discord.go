// Этот файл (sink_discord.go) находится в стадии активной разработки.
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
