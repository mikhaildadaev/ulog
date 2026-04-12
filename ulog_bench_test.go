package ulog

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Бенчмарки компонентов
func Benchmark_Logger_Debug_Multi(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
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
			if b.N == 1 {
				logger.DebugWithContext(ctx, "test debug simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.DebugWithContext(ctx, "test debug simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.DebugWithContext(ctx, "test debug format message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.DebugWithContext(ctx, "test debug format message")
				}
			})
		})
	}
}
func Benchmark_Logger_Debug_Single(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
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
			if b.N == 1 {
				logger.DebugWithContext(ctx, "test debug simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.DebugWithContext(ctx, "test debug simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.DebugWithContext(ctx, "test debug format message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.DebugWithContext(ctx, "test debug format message")
			}
		})
	}
}
func Benchmark_Logger_Error_Multi(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
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
			if b.N == 1 {
				logger.ErrorWithContext(ctx, "test error simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.ErrorWithContext(ctx, "test error simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.ErrorWithContext(ctx, "test error format message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.ErrorWithContext(ctx, "test error format message")
				}
			})
		})
	}
}
func Benchmark_Logger_Error_Single(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
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
			if b.N == 1 {
				logger.ErrorWithContext(ctx, "test error simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.ErrorWithContext(ctx, "test error simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.ErrorWithContext(ctx, "test error format message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.ErrorWithContext(ctx, "test error format message")
			}
		})
	}
}
func Benchmark_Logger_Info_Multi(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
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
			if b.N == 1 {
				logger.InfoWithContext(ctx, "test info simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.InfoWithContext(ctx, "test info simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.InfoWithContext(ctx, "test info format message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.InfoWithContext(ctx, "test info format message")
				}
			})
		})
	}
}
func Benchmark_Logger_Info_Single(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
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
			if b.N == 1 {
				logger.InfoWithContext(ctx, "test info simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.InfoWithContext(ctx, "test info simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.InfoWithContext(ctx, "test info format message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.InfoWithContext(ctx, "test info format message")
			}
		})
	}
}
func Benchmark_Logger_Warn_Multi(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
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
			if b.N == 1 {
				logger.WarnWithContext(ctx, "test warn simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.WarnWithContext(ctx, "test warn simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.WarnWithContext(ctx, "test warn format message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.WarnWithContext(ctx, "test warn format message")
				}
			})
		})
	}
}
func Benchmark_Logger_Warn_Single(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
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
			if b.N == 1 {
				logger.WarnWithContext(ctx, "test warn simple message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.WarnWithContext(ctx, "test warn simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.WarnWithContext(ctx, "test warn format message")
			}
			logger.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.WarnWithContext(ctx, "test warn format message")
			}
		})
	}
}
func Benchmark_LoggerLog_Error_Multi(b *testing.B) {
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
			logger := NewLogger(
				WithMode(format.mode, format.writer, format.bufferSize),
			)
			loggerLog := NewLoggerLog(LevelError, logger)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					loggerLog.Print("test error message")
				}
			})
		})
	}
}
func Benchmark_LoggerLog_Error_Single(b *testing.B) {
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
			logger := NewLogger(
				WithMode(format.mode, format.writer, format.bufferSize),
			)
			loggerLog := NewLoggerLog(LevelError, logger)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				loggerLog.Print("test error message")
			}
		})
	}
}
func BenchmarkTeeSink_Multi(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	tmpDir := filepath.Join(cwd, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		b.Fatal(err)
	}
	logFile := filepath.Join(tmpDir, "ulog_file.log")
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, nil, defaultBufferSize},
		{"Sync", ModeSync, nil, 0},
	}
	for _, format := range formats {
		sinkFile, err := NewFileSink(logFile,
			WithFileMaxSize(10),
		)
		if err != nil {
			b.Fatal(err)
		}
		defer sinkFile.Close()
		teeSink := NewTeeSink(sinkFile)
		defer teeSink.Close()
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.ErrorWithContext(ctx, "test error simple message")
			}
			logger.SetMode(format.mode, teeSink, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.ErrorWithContext(ctx, "test error simple message")
				}
			})
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.ErrorWithContext(ctx, "test error format message")
			}
			logger.SetMode(format.mode, teeSink, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					logger.ErrorWithContext(ctx, "test error format message")
				}
			})
		})
	}
}
func BenchmarkTeeSink_Single(b *testing.B) {
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	tmpDir := filepath.Join(cwd, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		b.Fatal(err)
	}
	logFile := filepath.Join(tmpDir, "ulog_file.log")
	formats := []struct {
		name       string
		mode       TypeMode
		writer     io.Writer
		bufferSize int
	}{
		{"Async", ModeAsync, nil, defaultBufferSize},
		{"Sync", ModeSync, nil, 0},
	}
	for _, format := range formats {
		sinkFile, err := NewFileSink(logFile,
			WithFileMaxSize(10),
		)
		if err != nil {
			b.Fatal(err)
		}
		defer sinkFile.Close()
		teeSink := NewTeeSink(sinkFile)
		defer teeSink.Close()
		b.Run("Simple "+format.name, func(b *testing.B) {
			logger := NewLogger()
			if b.N == 1 {
				logger.ErrorWithContext(ctx, "test error simple message")
			}
			logger.SetMode(format.mode, teeSink, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.ErrorWithContext(ctx, "test error simple message")
			}
		})
		b.Run("Format "+format.name, func(b *testing.B) {
			logger := NewLogger(
				WithExtractor("trace_id"),
			)
			if b.N == 1 {
				logger.ErrorWithContext(ctx, "test error format message")
			}
			logger.SetMode(format.mode, teeSink, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				logger.ErrorWithContext(ctx, "test error format message")
			}
		})
	}
}
