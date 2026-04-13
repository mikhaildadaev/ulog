// Этот файл (sink_slack.go) находится в стадии активной разработки.
// API может изменяться
//
// Планируется добавить:
// - Batch [пакетирование]
// - Circuit Breaker [защита от перегрузки]
// - Retry [повторные отправки]
package ulog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Публичные структуры
type SlackSink struct {
	channel    string
	client     *http.Client
	iconEmoji  string
	iconURL    string
	minLevel   TypeLevel
	username   string
	webhookURL string
}
type SlackOption func(*SlackSink)

// Публичные конструкторы
func NewSlackSink(minLevel TypeLevel, webhookURL string, options ...SlackOption) *SlackSink {
	sink := &SlackSink{
		client:     &http.Client{Timeout: 10 * time.Second},
		minLevel:   minLevel,
		webhookURL: webhookURL,
	}
	for _, option := range options {
		option(sink)
	}
	return sink
}

// Публичные функции
func WithSlackChannel(channel string) SlackOption {
	return func(slackSink *SlackSink) {
		slackSink.channel = channel
	}
}
func WithSlackIconEmoji(emoji string) SlackOption {
	return func(slackSink *SlackSink) {
		slackSink.iconEmoji = emoji
	}
}
func WithSlackIconURL(url string) SlackOption {
	return func(slackSink *SlackSink) {
		slackSink.iconURL = url
	}
}
func WithSlackTimeout(timeout time.Duration) SlackOption {
	return func(slackSink *SlackSink) {
		slackSink.client.Timeout = timeout
	}
}
func WithSlackUsername(username string) SlackOption {
	return func(slackSink *SlackSink) {
		slackSink.username = username
	}
}

// Публичные методы
func (slackSink *SlackSink) Close() error {
	slackSink.client.CloseIdleConnections()
	return nil
}
func (slackSink *SlackSink) Write(p []byte) (n int, err error) {
	msg := string(p)
	if len(msg) > maxSlackMessageLen {
		msg = msg[:maxSlackMessageLen-3] + "..."
	}
	webhook := struct {
		Text      string `json:"text"`
		Username  string `json:"username,omitempty"`
		IconEmoji string `json:"icon_emoji,omitempty"`
		IconURL   string `json:"icon_url,omitempty"`
		Channel   string `json:"channel,omitempty"`
	}{
		Text:      msg,
		Username:  slackSink.username,
		IconEmoji: slackSink.iconEmoji,
		IconURL:   slackSink.iconURL,
		Channel:   slackSink.channel,
	}
	jsonBody, err := json.Marshal(webhook)
	if err != nil {
		return 0, fmt.Errorf("slack marshal: %w", err)
	}
	resp, err := slackSink.client.Post(slackSink.webhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("slack post: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("slack: %s", resp.Status)
	}
	return len(p), nil
}
func (slackSink *SlackSink) WriteWithLevel(level TypeLevel, p []byte) (n int, err error) {
	if level < slackSink.minLevel {
		return len(p), nil
	}
	return slackSink.Write(p)
}
