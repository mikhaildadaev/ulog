package ulog

import (
	"io"
	"testing"
)

// Бенчмарки компонентов
func BenchmarkDebug(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		logger := New()
		logger.Debug("test debug message")
		logger.SetOutput(io.Discard)
		defer logger.Sync()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("test debug message")
		}
	})
	b.Run("Format", func(b *testing.B) {
		logger := New()
		logger.Debug("test debug message",
			String("code", "ERR_001"),
			Int("user_id", 12345),
			String("path", "/api/v1/test"),
		)
		logger.SetOutput(io.Discard)
		defer logger.Sync()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("test debug message",
				String("code", "ERR_001"),
				Int("user_id", 12345),
				String("path", "/api/v1/test"),
			)
		}
	})
}
func BenchmarkError(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		logger := New()
		logger.Error("test error message")
		logger.SetOutput(io.Discard)
		defer logger.Sync()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("test error message")
		}
	})
	b.Run("Format", func(b *testing.B) {
		logger := New()
		logger.Error("test error message",
			String("code", "ERR_001"),
			Int("user_id", 12345),
			String("path", "/api/v1/test"),
		)
		logger.SetOutput(io.Discard)
		defer logger.Sync()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("test error message",
				String("code", "ERR_001"),
				Int("user_id", 12345),
				String("path", "/api/v1/test"),
			)
		}
	})
}
func BenchmarkInfo(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		logger := New()
		logger.Info("test info message")
		logger.SetOutput(io.Discard)
		defer logger.Sync()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("test info message")
		}
	})
	b.Run("Format", func(b *testing.B) {
		logger := New()
		logger.Info("test info message",
			String("code", "ERR_001"),
			Int("user_id", 12345),
			String("path", "/api/v1/test"),
		)
		logger.SetOutput(io.Discard)
		defer logger.Sync()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("test info message",
				String("code", "ERR_001"),
				Int("user_id", 12345),
				String("path", "/api/v1/test"),
			)
		}
	})
}
func BenchmarkWarn(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		logger := New()
		logger.Warn("test warn message")
		logger.SetOutput(io.Discard)
		defer logger.Sync()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Warn("test warn message")
		}
	})
	b.Run("Format", func(b *testing.B) {
		logger := New()
		logger.Warn("test warn message",
			String("code", "ERR_001"),
			Int("user_id", 12345),
			String("path", "/api/v1/test"),
		)
		logger.SetOutput(io.Discard)
		defer logger.Sync()
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Warn("test warn message",
				String("code", "ERR_001"),
				Int("user_id", 12345),
				String("path", "/api/v1/test"),
			)
		}
	})
}
func BenchmarkLoggerWriter(b *testing.B) {
	logger := New()
	logger.SetOutput(io.Discard)
	defer logger.Sync()
	writer := NewWithWriter(logger, LevelInfo)
	data := []byte("test message")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(data)
	}
}
func BenchmarkWithDisabledLevel(b *testing.B) {
	logger := New()
	//logger.SetOutput(io.Discard)
	defer logger.Sync()
	logger.SetLevel(LevelError)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("this should be skipped")
	}
}
