// Этот файл (sink_discord.go) находится в стадии активной разработки.
// API может изменяться
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
type DiscordSink struct {
	avatarURL  string
	client     *http.Client
	tts        bool
	username   string
	webhookURL string
}
type DiscordOption func(*DiscordSink)

// Публичные конструкторы
func NewDiscordSink(webhookURL string, options ...DiscordOption) *DiscordSink {
	sink := &DiscordSink{
		client:     &http.Client{Timeout: 10 * time.Second},
		webhookURL: webhookURL,
	}
	for _, option := range options {
		option(sink)
	}
	return sink
}

// Публичные функции
func WithDiscordAvatarURL(avatarURL string) DiscordOption {
	return func(discordSink *DiscordSink) {
		discordSink.avatarURL = avatarURL
	}
}
func WithDiscordTimeout(timeout time.Duration) DiscordOption {
	return func(discordSink *DiscordSink) {
		discordSink.client.Timeout = timeout
	}
}
func WithDiscordTTS(tts bool) DiscordOption {
	return func(discordSink *DiscordSink) {
		discordSink.tts = tts
	}
}
func WithDiscordUsername(username string) DiscordOption {
	return func(discordSink *DiscordSink) {
		discordSink.username = username
	}
}

// Публичные методы
func (discordSink *DiscordSink) Close() error {
	discordSink.client.CloseIdleConnections()
	return nil
}
func (discordSink *DiscordSink) Write(p []byte) (n int, err error) {
	msg := string(p)
	if len(msg) > maxDiscordMessageLen {
		msg = msg[:maxDiscordMessageLen-3] + "..."
	}
	msg = escapeDiscordMarkdown(msg)
	webhook := struct {
		Content   string `json:"content"`
		Username  string `json:"username,omitempty"`
		AvatarURL string `json:"avatar_url,omitempty"`
		TTS       bool   `json:"tts,omitempty"`
	}{
		Content:   msg,
		Username:  discordSink.username,
		AvatarURL: discordSink.avatarURL,
		TTS:       discordSink.tts,
	}
	jsonBody, err := json.Marshal(webhook)
	if err != nil {
		return 0, fmt.Errorf("discord marshal: %w", err)
	}
	resp, err := discordSink.client.Post(discordSink.webhookURL, "application/json", bytes.NewReader(jsonBody))
	if err != nil {
		return 0, fmt.Errorf("discord post: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return 0, fmt.Errorf("discord: %s", resp.Status)
	}
	return len(p), nil
}

// Приватные функции
func escapeDiscordMarkdown(text string) string {
	specialChars := []string{"*", "_", "`", "~", "|"}
	result := text
	for _, char := range specialChars {
		result = strings.ReplaceAll(result, char, "\\"+char)
	}
	result = strings.ReplaceAll(result, "@everyone", "@\u200beveryone")
	result = strings.ReplaceAll(result, "@here", "@\u200bhere")
	return result
}
