package ulog

import (
	"io"
	"testing"
)

// Бенчмарки компонентов
func BenchmarkDebug(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, 10000},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.Debug("test debug simple message")
			}
			logger.SetOutput(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Debug("test debug simple message")
			}
			logger.Sync()
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.Debug("test debug format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetOutput(format.mode, format.writer, format.bufferSize)
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
}
func BenchmarkError(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, 10000},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.Error("test error simple message")
			}
			logger.SetOutput(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Error("test error simple message")
			}
			logger.Sync()
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.Error("test error format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetOutput(format.mode, format.writer, format.bufferSize)
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
}
func BenchmarkInfo(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, 10000},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.Info("test info simple message")
			}
			logger.SetOutput(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("test info simple message")
			}
			logger.Sync()
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.Info("test info format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetOutput(format.mode, format.writer, format.bufferSize)
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
}
func BenchmarkWarn(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, 10000},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.Warn("test warn simple message")
			}
			logger.SetOutput(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Warn("test warn simple message")
			}
			logger.Sync()
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.Warn("test warn format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetOutput(format.mode, format.writer, format.bufferSize)
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
}
func BenchmarkLoggerWriter(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, 10000},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		logger := NewLogger()
		logger.SetOutput(format.mode, format.writer, format.bufferSize)
		writer := NewWithWriter(LevelInfo, logger)
		data := []byte("test message")
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			writer.Write(data)
		}
		logger.Sync()
	}
}
