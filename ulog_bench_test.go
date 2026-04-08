package ulog

import (
	"io"
	"testing"
)

// Бенчмарки компонентов
func Benchmark_Debug_Multi(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Debug("test debug simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Debug("test debug simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Debug("test debug format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Debug("test debug format message",
						Int("user_id", 12345),
						String("path", "/api/v1/test"),
					)
				}
			})
		})
	}
}
func Benchmark_Debug_Single(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Debug("test debug simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Debug("test debug simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Debug("test debug format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Debug("test debug format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
		})
	}
}
func Benchmark_Error_Multi(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Error("test error simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Error("test error simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Error("test error format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Error("test error format message",
						Int("user_id", 12345),
						String("path", "/api/v1/test"),
					)
				}
			})
		})
	}
}
func Benchmark_Error_Single(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Error("test error simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Error("test error simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Error("test error format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Error("test error format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
		})
	}
}
func Benchmark_Info_Multi(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Info("test info simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info("test info simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Info("test info format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Info("test info format message",
						Int("user_id", 12345),
						String("path", "/api/v1/test"),
					)
				}
			})
		})
	}
}
func Benchmark_Info_Single(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Info("test info simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("test info simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Info("test info format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Info("test info format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
		})
	}
}
func Benchmark_Warn_Multi(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Warn("test warn simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Warn("test warn simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Warn("test warn format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.Warn("test warn format message",
						Int("user_id", 12345),
						String("path", "/api/v1/test"),
					)
				}
			})
		})
	}
}
func Benchmark_Warn_Single(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Warn("test warn simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Warn("test warn simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			if b.N == 1 {
				logger.Warn("test warn format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.Warn("test warn format message",
					Int("user_id", 12345),
					String("path", "/api/v1/test"),
				)
			}
		})
	}
}
func Benchmark_LoggerWriter_Multi(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run(format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			writer := NewWithWriter(LevelInfo, logger)
			data := []byte("test message")
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					writer.Write(data)
				}
			})
		})
	}
}
func Benchmark_LoggerWriter_Single(b *testing.B) {
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, io.Discard, defaultBufferSize},
		{"Sync", ModeSync, io.Discard, 0},
	}
	for _, format := range formats {
		b.Run(format.name, func(b *testing.B) {
			logger := NewLogger()
			defer logger.Close()
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			writer := NewWithWriter(LevelInfo, logger)
			data := []byte("test message")
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				writer.Write(data)
			}
		})
	}
}
