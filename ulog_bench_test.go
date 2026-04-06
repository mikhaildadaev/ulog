package ulog

import (
	"io"
	"testing"
)

// Бенчмарки компонентов
func BenchmarkDebug(b *testing.B) {
	logger := New()
	logger.SetOutput(io.Discard)
	defer logger.Sync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("test debug message")
	}
}
func BenchmarkError(b *testing.B) {
	logger := New()
	logger.SetOutput(io.Discard)
	defer logger.Sync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Error("test error message")
	}
}
func BenchmarkInfo(b *testing.B) {
	logger := New()
	logger.SetOutput(io.Discard)
	defer logger.Sync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Info("test info message")
	}
}
func BenchmarkWarn(b *testing.B) {
	logger := New()
	logger.SetOutput(io.Discard)
	defer logger.Sync()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Warn("test warn message")
	}
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

	defer logger.Sync()
	logger.SetLevel(LevelError)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("this should be skipped")
	}
}
