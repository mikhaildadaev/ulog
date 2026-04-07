package ulog

import (
	"io"
	"testing"
)

// Бенчмарки компонентов
func BenchmarkDebug(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		logger := NewLogger()
		logger.Debug("test debug simple message")
		logger.SetOutput(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("test debug simple message")
		}
		logger.Sync()
	})
	b.Run("Format", func(b *testing.B) {
		logger := NewLogger()
		logger.Debug("test debug format message",
			Int("user_id", 12345),
			String("path", "/api/v1/test"),
		)
		logger.SetOutput(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Debug("test debug format message",
				Int("user_id", 12345),
				String("path", "/api/v1/test"),
			)
		}
		logger.Sync()
	})
}
func BenchmarkError(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		logger := NewLogger()
		logger.Error("test error simple message")
		logger.SetOutput(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("test error simple message")
		}
		logger.Sync()
	})
	b.Run("Format", func(b *testing.B) {
		logger := NewLogger()
		logger.Error("test error format message",
			Int("user_id", 12345),
			String("path", "/api/v1/test"),
		)
		logger.SetOutput(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Error("test error format message",
				Int("user_id", 12345),
				String("path", "/api/v1/test"),
			)
		}
		logger.Sync()
	})
}
func BenchmarkInfo(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		logger := NewLogger()
		logger.Info("test info simple message")
		logger.SetOutput(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("test info simple message")
		}
		logger.Sync()
	})
	b.Run("Format", func(b *testing.B) {
		logger := NewLogger()
		logger.Info("test info format message",
			Int("user_id", 12345),
			String("path", "/api/v1/test"),
		)
		logger.SetOutput(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Info("test info format message",
				Int("user_id", 12345),
				String("path", "/api/v1/test"),
			)
		}
		logger.Sync()
	})
}
func BenchmarkWarn(b *testing.B) {
	b.Run("Simple", func(b *testing.B) {
		logger := NewLogger()
		logger.Warn("test warn simple message")
		logger.SetOutput(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Warn("test warn simple message")
		}
		logger.Sync()
	})
	b.Run("Format", func(b *testing.B) {
		logger := NewLogger()
		logger.Warn("test warn format message",
			Int("user_id", 12345),
			String("path", "/api/v1/test"),
		)
		logger.SetOutput(io.Discard)
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			logger.Warn("test warn format message",
				Int("user_id", 12345),
				String("path", "/api/v1/test"),
			)
		}
		logger.Sync()
	})
}
func BenchmarkLoggerWriter(b *testing.B) {
	logger := NewLogger()
	logger.SetOutput(io.Discard)
	writer := NewWithWriter(logger, LevelInfo)
	data := []byte("test message")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		writer.Write(data)
	}
	logger.Sync()
}
