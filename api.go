package ulog

import (
	"context"
	"fmt"
	"io"
	"time"
)

// Публичные функции
func WithExtractor(keys ...string) OptionLogger {
	return func(universalLogger *universalLogger) {
		universalLogger.extractor = func(ctx context.Context) []Field {
			if ctx == nil {
				return nil
			}
			fields := make([]Field, 0, len(keys))
			for _, key := range keys {
				if val := ctx.Value(key); val != nil {
					switch v := val.(type) {
					case string:
						fields = append(fields, String(key, v))
					case int:
						fields = append(fields, Int(key, v))
					case int64:
						fields = append(fields, Int64(key, v))
					case float32:
						fields = append(fields, Float64(key, float64(v)))
					case float64:
						fields = append(fields, Float64(key, v))
					case bool:
						fields = append(fields, Bool(key, v))
					case time.Time:
						fields = append(fields, Time(key, v))
					case time.Duration:
						fields = append(fields, Duration(key, v))
					default:
						fields = append(fields, String(key, fmt.Sprintf("%v", v)))
					}
				}
			}
			return fields
		}
	}
}
func WithFormat(format TypeFormat) OptionLogger {
	return func(universalLogger *universalLogger) {
		universalLogger.format.Store(int32(format))
	}
}
func WithLevel(level TypeLevel) OptionLogger {
	return func(universalLogger *universalLogger) {
		universalLogger.level.Store(int32(level))
	}
}
func WithMode(mode TypeMode, writer io.Writer, bufferSize ...int) OptionLogger {
	return func(universalLogger *universalLogger) {
		switch mode {
		case ModeAsync:
			if bufferSize == nil || bufferSize[0] <= 0 {
				bufferSize[0] = defaultBufferSize
			}
			universalLogger.mode = ModeAsync
			universalLogger.writer = newAsyncWriter(writer, bufferSize[0])
		case ModeSync:
			universalLogger.mode = ModeSync
			universalLogger.writer = writer
		}
	}
}
func WithTheme(theme TypeTheme) OptionLogger {
	return func(universalLogger *universalLogger) {
		switch theme {
		case ThemeDark:
			universalLogger.theme = darkTheme
		case ThemeLight:
			universalLogger.theme = lightTheme
		}
	}
}

// Публичные методы
func (standardLogger *standardLogger) Write(p []byte) (n int, err error) {
	standardLogger.mutex.Lock()
	defer standardLogger.mutex.Unlock()
	start := 0
	end := len(p)
	for start < end && p[start] <= ' ' {
		start++
	}
	for end > start && p[end-1] <= ' ' {
		end--
	}
	if start >= end {
		return 0, nil
	}
	if isIgnoredError(p[start:end]) {
		return len(p), nil
	}
	message := string(p[start:end])
	switch TypeLevel(standardLogger.level.Load()) {
	case LevelDebug:
		standardLogger.logger.Debug(message)
	case LevelInfo:
		standardLogger.logger.Info(message)
	case LevelWarn:
		standardLogger.logger.Warn(message)
	case LevelError:
		standardLogger.logger.Error(message)
	case LevelFatal:
		standardLogger.logger.Fatal(message)
	}
	return len(p), nil
}
func (universalLogger *universalLogger) Close() error {
	if universalLogger.mode != ModeAsync {
		return nil
	}
	universalLogger.mutex.RLock()
	currentWriter := universalLogger.writer
	universalLogger.mutex.RUnlock()
	if asyncWriter, ok := currentWriter.(*asyncWriter); ok {
		return asyncWriter.Close()
	}
	return nil
}
func (universalLogger *universalLogger) Debug(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelDebug, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelDebug, context.Background(), message, fields)
	}
}
func (universalLogger *universalLogger) DebugWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelDebug, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelDebug, context, message, fields)
	}
}
func (universalLogger *universalLogger) Error(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelError, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelError, context.Background(), message, fields)
	}
}
func (universalLogger *universalLogger) ErrorWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelError, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelError, context, message, fields)
	}
}
func (universalLogger *universalLogger) Fatal(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelFatal, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelFatal, context.Background(), message, fields)
	}
	osExit(1)
}
func (universalLogger *universalLogger) FatalWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelFatal, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelFatal, context, message, fields)
	}
	osExit(1)
}
func (universalLogger *universalLogger) Info(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelInfo, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelInfo, context.Background(), message, fields)
	}
}
func (universalLogger *universalLogger) InfoWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelInfo, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelInfo, context, message, fields)
	}
}
func (universalLogger *universalLogger) Warn(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelWarn, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelWarn, context.Background(), message, fields)
	}
}
func (universalLogger *universalLogger) WarnWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelWarn, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelWarn, context, message, fields)
	}
}
func (universalLogger *universalLogger) SetExtractor(keys ...string) {
	universalLogger.mutex.Lock()
	defer universalLogger.mutex.Unlock()
	universalLogger.extractor = func(ctx context.Context) []Field {
		if ctx == nil {
			return nil
		}
		fields := make([]Field, 0, len(keys))
		for _, key := range keys {
			if val := ctx.Value(key); val != nil {
				switch v := val.(type) {
				case string:
					fields = append(fields, String(key, v))
				case int:
					fields = append(fields, Int(key, v))
				case int64:
					fields = append(fields, Int64(key, v))
				case float32:
					fields = append(fields, Float64(key, float64(v)))
				case float64:
					fields = append(fields, Float64(key, v))
				case bool:
					fields = append(fields, Bool(key, v))
				case time.Time:
					fields = append(fields, Time(key, v))
				case time.Duration:
					fields = append(fields, Duration(key, v))
				default:
					fields = append(fields, String(key, fmt.Sprintf("%v", v)))
				}
			}
		}
		return fields
	}
}
func (universalLogger *universalLogger) SetFormat(format TypeFormat) {
	universalLogger.format.Store(int32(format))
}
func (universalLogger *universalLogger) SetLevel(level TypeLevel) {
	universalLogger.level.Store(int32(level))
}
func (universalLogger *universalLogger) SetMode(mode TypeMode, writer io.Writer, bufferSize ...int) {
	universalLogger.mutex.Lock()
	defer universalLogger.mutex.Unlock()
	if universalLogger.mode == ModeAsync {
		if asyncWriter, ok := universalLogger.writer.(*asyncWriter); ok {
			asyncWriter.Close()
			time.Sleep(time.Millisecond)
		}
	}
	switch mode {
	case ModeAsync:
		if bufferSize == nil || bufferSize[0] <= 0 {
			bufferSize[0] = defaultBufferSize
		}
		universalLogger.mode = ModeAsync
		universalLogger.writer = newAsyncWriter(writer, bufferSize[0])
	case ModeSync:
		universalLogger.mode = ModeSync
		universalLogger.writer = writer
	}
}
func (universalLogger *universalLogger) SetTheme(theme TypeTheme) {
	universalLogger.mutex.Lock()
	defer universalLogger.mutex.Unlock()
	switch theme {
	case ThemeDark:
		universalLogger.theme = darkTheme
	case ThemeLight:
		universalLogger.theme = lightTheme
	}
}
func (universalLogger *universalLogger) Sync() error {
	if universalLogger.mode != ModeAsync {
		return nil
	}
	universalLogger.mutex.RLock()
	currentWriter := universalLogger.writer
	universalLogger.mutex.RUnlock()
	if asyncWriter, ok := currentWriter.(*asyncWriter); ok {
		return asyncWriter.Sync()
	}
	return nil
}
func (universalLogger *universalLogger) Write(p []byte) (n int, err error) {
	universalLogger.mutex.RLock()
	writer := universalLogger.writer
	universalLogger.mutex.RUnlock()
	return writer.Write(p)
}
