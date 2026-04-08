package ulog

import (
	"bytes"
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
func TestColorScheme(t *testing.T) {
	if lightScheme.prefixError != colorLightRed {
		t.Error("Light scheme has wrong color")
	}
	t.Setenv("TERM_THEME", "light")
	scheme := getLoggerScheme()
	if scheme.prefixError != colorLightRed {
		t.Error("Should detect light theme from TERM_THEME")
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

//	func TestFormatting(t *testing.T) {
//		var buf bytes.Buffer
//		logger := &LoggerStandard{
//			Logger: log.New(&buf, "", 0),
//			level:  LevelDebug,
//			scheme: darkScheme,
//		}
//		logger.Info("User %s has %d points", "alice", 100)
//		if !strings.Contains(buf.String(), "User alice has 100 points") {
//			t.Errorf("Formatted message not correct: %s", buf.String())
//		}
//	}
func TestLoggerError(t *testing.T) {
	var buf bytes.Buffer
	logger := &UniversalLogger{
		scheme: darkScheme,
	}
	logger.level.Store(int32(LevelDebug))
	stdLogger := NewLoggerError(logger)
	stdLogger.Println("")
	if !strings.Contains(buf.String(), colorDarkRed+"[ERROR]") {
		t.Errorf("Expected ERROR level, got %s", buf.String())
	}
}
func TestWithWriter(t *testing.T) {
	var buf bytes.Buffer
	logger := &UniversalLogger{
		scheme: darkScheme,
	}
	logger.level.Store(int32(LevelDebug))
	writer := NewWithWriter(LevelWarn, logger)
	writer.Write([]byte("test message"))
	output := buf.String()
	checks := []string{"[WARN]", "test message"}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("Expected %q not found in %q", check, output)
		}
	}
}
func TestLevel(t *testing.T) {
	t.Run("Fatal level", func(t *testing.T) {
		var exited bool
		oldExit := osExit
		defer func() { osExit = oldExit }()
		osExit = func(int) { exited = true }
		var buf bytes.Buffer
		logger := &UniversalLogger{
			scheme: darkScheme,
		}
		logger.level.Store(int32(LevelError))
		logger.Fatal("")
		if !exited {
			t.Error("Fatal should call os.Exit")
		}
		if !strings.Contains(buf.String(), colorDarkPurple+"[FATAL]") {
			t.Errorf("Fatal message not logged: %s", buf.String())
		}
	})
	t.Run("Standard level", func(t *testing.T) {
		var buf bytes.Buffer
		logger := &UniversalLogger{
			scheme: darkScheme,
		}
		logger.level.Store(int32(LevelDebug))
		tests := []struct {
			name     string
			logFunc  func()
			expected string
		}{
			{"Debug", func() { logger.Debug("") }, colorDarkCyan + "[DEBUG]"},
			{"Error", func() { logger.Error("") }, colorDarkRed + "[ERROR]"},
			{"Info", func() { logger.Info("") }, colorDarkGreen + "[INFO]"},
			{"Warn", func() { logger.Warn("") }, colorDarkYellow + "[WARN]"},
		}
		for _, tt := range tests {
			buf.Reset()
			tt.logFunc()
			if !strings.Contains(buf.String(), tt.expected) {
				t.Errorf("%s: expected %q, got %q", tt.name, tt.expected, buf.String())
			}
		}
	})
}
func TestLevelFiltering(t *testing.T) {
	levels := []TypeLevel{LevelDebug, LevelInfo, LevelWarn, LevelError}
	for _, level := range levels {
		var buf bytes.Buffer
		logger := &UniversalLogger{
			scheme: darkScheme,
		}
		logger.level.Store(int32(level))
		logger.Debug("debug")
		logger.Info("info")
		logger.Warn("warn")
		logger.Error("error")
		output := buf.String()
		switch level {
		case LevelInfo:
			if strings.Contains(output, "[DEBUG]") {
				t.Error("Debug message should not appear at Info level")
			}
		case LevelWarn:
			if strings.Contains(output, "[DEBUG]") || strings.Contains(output, "[INFO]") {
				t.Error("Debug/Info messages should not appear at Warn level")
			}
		case LevelError:
			if strings.Contains(output, "[DEBUG]") || strings.Contains(output, "[INFO]") || strings.Contains(output, "[WARN]") {
				t.Error("Only Error messages should appear at Error level")
			}
		}
	}
}
func TestSetLevel(t *testing.T) {
	logger := NewLogger().(*UniversalLogger)
	logger.SetLevel(LevelError)
	if logger.getLevel() != LevelError {
		t.Errorf("Expected level %d, got %d", LevelError, logger.getLevel())
	}
}
func TestSetTheme(t *testing.T) {
	logger := NewLogger().(*UniversalLogger)
	logger.SetTheme(ThemeDark)
	if logger.getScheme() != darkScheme {
		t.Error("Theme not changed to dark")
	}
	logger.SetTheme(ThemeLight)
	if logger.getScheme() != lightScheme {
		t.Error("Theme not changed to light")
	}
}
