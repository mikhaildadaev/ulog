package ulog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Публичные структуры
type SinkSlack struct {
	WebhookURL string
	Username   string
	IconEmoji  string
	IconURL    string
	Channel    string
	Client     *http.Client
}
type SinkTelegram struct {
	BotToken string
	ChatID   string
	Client   *http.Client
}
type WebhookSlack struct {
	Text      string `json:"text"`
	Username  string `json:"username,omitempty"`
	IconEmoji string `json:"icon_emoji,omitempty"`
	IconURL   string `json:"icon_url,omitempty"`
	Channel   string `json:"channel,omitempty"`
}
type WebhookDiscord struct {
	Content   string `json:"content"`
	Username  string `json:"username,omitempty"`
	AvatarURL string `json:"avatar_url,omitempty"`
	TTS       bool   `json:"tts,omitempty"`
}
type SinkDiscord struct {
	WebhookURL string
	Username   string
	AvatarURL  string
	Client     *http.Client
}

// Публичные методы
func (sinkDiscord *SinkDiscord) Write(p []byte) (n int, err error) {
	if sinkDiscord.Client == nil {
		sinkDiscord.Client = &http.Client{Timeout: maxTimeout}
	}
	msg := string(p)
	if len(msg) > maxDiscordMessageLen {
		msg = msg[:maxDiscordMessageLen-3] + "..."
	}
	webhook := WebhookDiscord{
		Content:   msg,
		Username:  sinkDiscord.Username,
		AvatarURL: sinkDiscord.AvatarURL,
		TTS:       false,
	}
	jsonBody, err := json.Marshal(webhook)
	if err != nil {
		return 0, fmt.Errorf("telegram marshal: %w", err)
	}
	resp, err := sinkDiscord.Client.Post(sinkDiscord.WebhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("discord webhook returned status: %s", resp.Status)
	}
	return len(p), nil
}
func (sinkSlack *SinkSlack) Write(p []byte) (n int, err error) {
	if sinkSlack.Client == nil {
		sinkSlack.Client = &http.Client{Timeout: maxTimeout}
	}
	msg := string(p)
	if len(msg) > maxSlackMessageLen {
		msg = msg[:maxSlackMessageLen-3] + "..."
	}
	webhook := WebhookSlack{
		Text:      msg,
		Username:  sinkSlack.Username,
		IconEmoji: sinkSlack.IconEmoji,
		IconURL:   sinkSlack.IconURL,
		Channel:   sinkSlack.Channel,
	}
	jsonBody, err := json.Marshal(webhook)
	if err != nil {
		return 0, err
	}
	resp, err := sinkSlack.Client.Post(sinkSlack.WebhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("slack webhook returned status: %s", resp.Status)
	}
	return len(p), nil
}
func (sinkTelegram *SinkTelegram) Write(p []byte) (n int, err error) {
	if sinkTelegram.Client == nil {
		sinkTelegram.Client = &http.Client{Timeout: maxTimeout}
	}
	msg := string(p)
	if len(msg) > maxTelegramMessageLen {
		msg = msg[:maxTelegramMessageLen-3] + "..."
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", sinkTelegram.BotToken)
	body := map[string]string{
		"chat_id": sinkTelegram.ChatID,
		"text":    msg,
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := sinkTelegram.Client.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("telegram API returned status: %s", resp.Status)
	}
	return len(p), nil
}

// Приватные константы
const (
	maxDiscordMessageLen  = 2000
	maxSlackMessageLen    = 4000
	maxTelegramMessageLen = 4096
	maxTimeout            = 10 * time.Second
)

// Данный файл находится в стадии разработки
// Batch [пакетировнаия]
// Circuit Breaker [защита от перегрузки]
// Retry [повторных попыток]
