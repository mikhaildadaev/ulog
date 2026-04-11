package ulog

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// Публичные структуры
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
type WriterDiscord struct {
	WebhookURL string
	Username   string
	AvatarURL  string
	Client     *http.Client
}
type WriterSlack struct {
	WebhookURL string
	Username   string
	IconEmoji  string
	IconURL    string
	Channel    string
	Client     *http.Client
}
type WriterTelegram struct {
	BotToken string
	ChatID   string
	Client   *http.Client
}

// Публичные методы
func (writerDiscord *WriterDiscord) Write(p []byte) (n int, err error) {
	if writerDiscord.Client == nil {
		writerDiscord.Client = &http.Client{}
	}
	webhook := WebhookDiscord{
		Content:   string(p),
		Username:  writerDiscord.Username,
		AvatarURL: writerDiscord.AvatarURL,
		TTS:       false,
	}
	jsonBody, err := json.Marshal(webhook)
	if err != nil {
		return 0, err
	}
	resp, err := writerDiscord.Client.Post(writerDiscord.WebhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("discord webhook returned status: %s", resp.Status)
	}
	return len(p), nil
}
func (writerSlack *WriterSlack) Write(p []byte) (n int, err error) {
	if writerSlack.Client == nil {
		writerSlack.Client = &http.Client{}
	}
	webhook := WebhookSlack{
		Text:      string(p),
		Username:  writerSlack.Username,
		IconEmoji: writerSlack.IconEmoji,
		IconURL:   writerSlack.IconURL,
		Channel:   writerSlack.Channel,
	}
	jsonBody, err := json.Marshal(webhook)
	if err != nil {
		return 0, err
	}
	resp, err := writerSlack.Client.Post(writerSlack.WebhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("slack webhook returned status: %s", resp.Status)
	}
	return len(p), nil
}
func (writerTelegram *WriterTelegram) Write(p []byte) (n int, err error) {
	if writerTelegram.Client == nil {
		writerTelegram.Client = &http.Client{}
	}
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", writerTelegram.BotToken)
	body := map[string]string{
		"chat_id": writerTelegram.ChatID,
		"text":    string(p),
	}
	jsonBody, _ := json.Marshal(body)
	resp, err := writerTelegram.Client.Post(url, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("telegram API returned status: %s", resp.Status)
	}
	return len(p), nil
}
