// Этот файл (sink_telegram.go) находится в стадии активной разработки.
// API может изменяться
package ulog

import "encoding/json"

// Публичные структуры
type TelegramData struct {
	ChatID              string `json:"chat_id"`
	DisableNotification bool   `json:"disable_notification,omitempty"`
	Text                string `json:"text"`
	ParseMode           string `json:"parse_mode,omitempty"`
}
type TelegramSink = HttpSink

// Публичные конструкторы
func NewTelegramSink(botToken, chatID string, options ...HttpOption) *HttpSink {
	apiURL := "https://api.telegram.org/bot" + botToken + "/sendMessage"
	return NewHttpSink(apiURL, append([]HttpOption{
		WithHttpMethod("POST"),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpLevelMin(LevelError),
		WithHttpFormatter(func(level TypeLevel, p []byte) ([]byte, error) {
			data := TelegramData{
				ChatID:    chatID,
				Text:      string(p),
				ParseMode: "HTML",
			}
			return json.Marshal(data)
		}),
	}, options...)...)
}
