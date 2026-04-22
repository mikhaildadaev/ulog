package ulog

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

// Бенчмарки компонентов
func Benchmark_Telemetry_Debug_Multi(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
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
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
			)
			if b.N == 1 {
				telemetry.DebugWithContext(ctx, DataLog, String("message", "test debug text"))
			}
			telemetry.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					telemetry.DebugWithContext(ctx, DataLog, String("message", "test debug text"))
				}
			})
		})
	}
}
func Benchmark_Telemetry_Debug_Single(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
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
			telemetry := NewTelemetry()
			if b.N == 1 {
				telemetry.DebugWithContext(ctx, DataLog, String("message", "test debug text"))
			}
			telemetry.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				telemetry.DebugWithContext(ctx, DataLog, String("message", "test debug text"))
			}
		})
	}
}
func Benchmark_Telemetry_Error_Multi(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
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
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
			)
			if b.N == 1 {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
			telemetry.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
				}
			})
		})
	}
}
func Benchmark_Telemetry_Error_Single(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
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
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
			)
			if b.N == 1 {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
			telemetry.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
		})
	}
}
func Benchmark_Telemetry_Info_Multi(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
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
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
			)
			if b.N == 1 {
				telemetry.InfoWithContext(ctx, DataLog, String("message", "test info text"))
			}
			telemetry.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					telemetry.InfoWithContext(ctx, DataLog, String("message", "test info text"))
				}
			})
		})
	}
}
func Benchmark_Telemetry_Info_Single(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
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
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
			)
			if b.N == 1 {
				telemetry.InfoWithContext(ctx, DataLog, String("message", "test info text"))
			}
			telemetry.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				telemetry.InfoWithContext(ctx, DataLog, String("message", "test info text"))
			}
		})
	}
}
func Benchmark_Telemetry_Warn_Multi(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
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
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
			)
			if b.N == 1 {
				telemetry.WarnWithContext(ctx, DataLog, String("message", "test warn text"))
			}
			telemetry.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					telemetry.WarnWithContext(ctx, DataLog, String("message", "test warn text"))
				}
			})
		})
	}
}
func Benchmark_Telemetry_Warn_Single(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
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
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
			)
			if b.N == 1 {
				telemetry.WarnWithContext(ctx, DataLog, String("message", "test warn text"))
			}
			telemetry.SetMode(format.mode, format.writer, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				telemetry.WarnWithContext(ctx, DataLog, String("message", "test warn text"))
			}
		})
	}
}
func Benchmark_TelemetryLog_Error_Multi(b *testing.B) {
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
			telemetry := NewTelemetry(
				WithMode(format.mode, format.writer, format.bufferSize),
			)
			telemetryLog := NewTelemetryLog(LevelError, telemetry)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					telemetryLog.Print("test error text")
				}
			})
		})
	}
}
func Benchmark_TelemetryLog_Error_Single(b *testing.B) {
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
			telemetry := NewTelemetry(
				WithMode(format.mode, format.writer, format.bufferSize),
			)
			telemetryLog := NewTelemetryLog(LevelError, telemetry)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				telemetryLog.Print("test error text")
			}
		})
	}
}
func Benchmark_SinkFile_Multi(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	tmpDir := filepath.Join(cwd, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		b.Fatal(err)
	}
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
		b.Run(format.name, func(b *testing.B) {
			logFile := filepath.Join(tmpDir, "ulog_file.log")
			sinkFile, err := NewFileSink(logFile,
				WithFileMaxSize(15),
			)
			if err != nil {
				b.Fatal(err)
			}
			teeSink := NewTeeSink(sinkFile)
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
				WithFormat(FormatJson),
			)
			if b.N == 1 {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
			telemetry.SetMode(format.mode, teeSink, format.bufferSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
				}
			})
		})
	}
}
func Benchmark_SinkFile_Single(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	cwd, err := os.Getwd()
	if err != nil {
		b.Fatal(err)
	}
	tmpDir := filepath.Join(cwd, "tmp")
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		b.Fatal(err)
	}
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
		b.Run(format.name, func(b *testing.B) {
			logFile := filepath.Join(tmpDir, "ulog_file.log")
			sinkFile, err := NewFileSink(logFile,
				WithFileMaxSize(15),
			)
			if err != nil {
				b.Fatal(err)
			}
			teeSink := NewTeeSink(sinkFile)
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
				WithFormat(FormatJson),
			)
			if b.N == 1 {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
			telemetry.SetMode(format.mode, teeSink, format.bufferSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
		})
	}
}
func Benchmark_SinkHttp_Multi(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	formats := []struct {
		name    string
		mode    TypeMode
		bufSize int
	}{
		{"Async", ModeAsync, defaultBufferSize},
		{"Sync", ModeSync, 0},
	}
	for _, format := range formats {
		b.Run(format.name, func(b *testing.B) {
			sink := NewHttpSink(server.URL,
				WithHttpDisabledBatch(),
				WithHttpFilterLevel(LevelDebug),
			)
			tee := NewTeeSink(sink)
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
				WithFormat(FormatJson),
				WithLevel(LevelDebug),
			)
			if b.N == 1 {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
			telemetry.SetMode(format.mode, tee, format.bufSize)
			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				for pb.Next() {
					telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
				}
			})
		})
	}
}
func Benchmark_SinkHttp_Single(b *testing.B) {
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	formats := []struct {
		name    string
		mode    TypeMode
		bufSize int
	}{
		{"Async", ModeAsync, defaultBufferSize},
		{"Sync", ModeSync, 0},
	}
	for _, format := range formats {
		b.Run(format.name, func(b *testing.B) {
			sink := NewHttpSink(server.URL,
				WithHttpDisabledBatch(),
				WithHttpFilterLevel(LevelDebug),
			)
			tee := NewTeeSink(sink)
			telemetry := NewTelemetry(
				WithExtractor("node_id", "trace_id"),
				WithFormat(FormatJson),
				WithLevel(LevelDebug),
			)
			if b.N == 1 {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
			telemetry.SetMode(format.mode, tee, format.bufSize)
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				telemetry.ErrorWithContext(ctx, DataLog, String("message", "test error text"))
			}
		})
	}
}
