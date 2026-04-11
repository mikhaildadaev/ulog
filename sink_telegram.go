package ulog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// Публичные структуры
type TelegramSink struct {
	botToken              string
	chatID                string
	parseMode             string
	disableWebPagePreview bool
	disableNotification   bool
	client                *http.Client
}
type TelegramOption func(*TelegramSink)

// Публичные конструкторы
func NewTelegramSink(botToken, chatID string, options ...TelegramOption) *TelegramSink {
	sink := &TelegramSink{
		botToken:  botToken,
		chatID:    chatID,
		parseMode: "HTML",
		client:    &http.Client{Timeout: 10 * time.Second},
	}
	for _, option := range options {
		option(sink)
	}
	return sink
}

// Публичные функции
func WithTelegramDisableNotification(disable bool) TelegramOption {
	return func(t *TelegramSink) {
		t.disableNotification = disable
	}
}
func WithTelegramDisableWebPagePreview(disable bool) TelegramOption {
	return func(t *TelegramSink) {
		t.disableWebPagePreview = disable
	}
}
func WithTelegramParseMode(mode string) TelegramOption {
	return func(t *TelegramSink) {
		t.parseMode = mode
	}
}
func WithTelegramTimeout(timeout time.Duration) TelegramOption {
	return func(t *TelegramSink) {
		t.client.Timeout = timeout
	}
}

// Публичные методы
func (telegramSink *TelegramSink) Close() error {
	telegramSink.client.CloseIdleConnections()
	return nil
}
func (telegramSink *TelegramSink) Write(p []byte) (n int, err error) {
	msg := string(p)
	if len(msg) > maxTelegramMessageLen {
		msg = msg[:maxTelegramMessageLen-3] + "..."
	}
	if telegramSink.parseMode == "MarkdownV2" {
		msg = escapeTelegramMarkdownV2(msg)
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramSink.botToken)
	body := map[string]interface{}{
		"chat_id": telegramSink.chatID,
		"text":    msg,
	}
	if telegramSink.parseMode != "" {
		body["parse_mode"] = telegramSink.parseMode
	}
	if telegramSink.disableWebPagePreview {
		body["disable_web_page_preview"] = true
	}
	if telegramSink.disableNotification {
		body["disable_notification"] = true
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return 0, fmt.Errorf("telegram marshal: %w", err)
	}
	resp, err := telegramSink.client.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("telegram post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("telegram: %s", resp.Status)
	}
	return len(p), nil
}

// Приватные функции
func escapeTelegramMarkdownV2(text string) string {
	specialChars := []string{"_", "*", "[", "]", "(", ")", "~", "`", ">", "#", "+", "-", "=", "|", "{", "}", ".", "!"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	return result
}
