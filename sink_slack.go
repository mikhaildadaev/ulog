// Этот файл (sink_slack.go) находится в стадии активной разработки.
// API может изменятьсяs
package ulog

import (
	"encoding/json"
)

// Публичные структуры
type SlackData struct {
	Channel   string `json:"channel,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	Text      string `json:"text"`
	UserName  string `json:"username,omitempty"`
}
type SlackSink = HttpSink

// Публичные конструкторы
func NewSlackSink(endPoint, userName, iconEmoji, iconURL, channel string, params ...HttpParams) *HttpSink {
	return NewHttpSink(endPoint, append([]HttpParams{
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
