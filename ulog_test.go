package ulog

import (
	"strings"
	"sync"
	"testing"
)

// Публичные функции
func TestIsIgnoredError(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"EOF", []byte("read: EOF"), true},
		{"TLS handshake", []byte("TLS handshake error"), true},
		{"connection refused", []byte("dial: connection refused"), true},
		{"timeout", []byte("i/o timeout"), true},
		{"normal message", []byte("user logged in"), false},
		{"empty", []byte{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIgnoredError(tt.data); got != tt.expected {
				t.Errorf("isIgnoredError(%q) = %v, want %v", tt.data, got, tt.expected)
			}
		})
	}
}
func TestConcurrency(t *testing.T) {
	logger := NewLogger()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.SetLevel(LevelDebug)
			logger.SetTheme(ThemeDark)
		}()
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("test")
			logger.Debug("debug")
		}()
	}
	wg.Wait()
}
func TestEnvLogLevel(t *testing.T) {
	tests := []struct {
		env      string
		expected TypeLevel
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"fatal", LevelFatal},
		{"", LevelInfo},
	}
	for _, tt := range tests {
		t.Setenv("LOG_LEVEL", tt.env)
		if got := getLoggerLevel(); got != tt.expected {
			t.Errorf("LOG_LEVEL=%s: got %d, want %d", tt.env, got, tt.expected)
		}
	}
}
func TestGetLoggerTheme(t *testing.T) {
	t.Run("Dark theme", func(t *testing.T) {
		t.Setenv("TERM_THEME", "dark")
		theme := getLoggerTheme()
		if !strings.HasPrefix(theme.prefixError, colorDarkRed) {
			t.Error("Light theme prefix should start with light red color")
		}
		if !strings.HasPrefix(theme.prefixDebug, colorDarkCyan) {
			t.Error("Light theme debug should start with cyan")
		}
		if !strings.HasPrefix(theme.prefixFatal, colorDarkPurple) {
			t.Error("Light theme fatal should start with purple")
		}
		if !strings.HasPrefix(theme.prefixInfo, colorDarkGreen) {
			t.Error("Light theme info should start with green")
		}
		if !strings.HasPrefix(theme.prefixWarn, colorDarkYellow) {
			t.Error("Light theme warn should start with yellow")
		}
	})
	t.Run("Light theme", func(t *testing.T) {
		t.Setenv("TERM_THEME", "light")
		theme := getLoggerTheme()
		if !strings.HasPrefix(theme.prefixError, colorLightRed) {
			t.Error("Light theme prefix should start with light red color")
		}
		if !strings.HasPrefix(theme.prefixDebug, colorLightCyan) {
			t.Error("Light theme debug should start with cyan")
		}
		if !strings.HasPrefix(theme.prefixFatal, colorLightPurple) {
			t.Error("Light theme fatal should start with purple")
		}
		if !strings.HasPrefix(theme.prefixInfo, colorLightGreen) {
			t.Error("Light theme info should start with green")
		}
		if !strings.HasPrefix(theme.prefixWarn, colorLightYellow) {
			t.Error("Light theme warn should start with yellow")
		}
	})
}

//func TestSetLevel(t *testing.T) {
//	logger := NewLogger()
//	logger.SetLevel(LevelError)
//	if logger.getLevel() != LevelError {
//		t.Errorf("Expected level %d, got %d", LevelError, logger.getLevel())
//	}
//}

//func TestSetTheme(t *testing.T) {
//	logger := NewLogger(WithTheme(ThemeDark))
//	buf := &bytes.Buffer{}
//	logger.SetMode(ModeSync, buf, 0)
//	logger.Info("test message")
//	output := buf.String()
//	if !strings.Contains(output, "\033[92m") {
//		t.Error("Dark theme colors not found in output")
//	}
//	buf.Reset()
//	logger.SetTheme(ThemeLight)
//	logger.Info("test message")
//	if !strings.Contains(output, "\033[92m") {
//		t.Error("Light theme colors not found")
//	}
//}
