package ulog

import (
	"context"
	"io"
	"strings"
	"time"
)

// Публичные функции
func WithAsync(bufferSize int) OptionLogger {
	return func(universalLogger *UniversalLogger) {
		if bufferSize <= 0 {
			bufferSize = 10000
		}
		universalLogger.async = true
		universalLogger.writer = newAsyncWriter(universalLogger.writer, bufferSize)
	}
}
func WithFormat(format TypeFormat) OptionLogger {
	return func(universalLogger *UniversalLogger) {
		universalLogger.format = format
	}
}
func WithLevel(level TypeLevel) OptionLogger {
	return func(universalLogger *UniversalLogger) {
		universalLogger.level.Store(int32(level))
	}
}
func WithOutput(writer io.Writer) OptionLogger {
	return func(universalLogger *UniversalLogger) {
		universalLogger.writer = writer
	}
}
func WithTheme(theme TypeTheme) OptionLogger {
	return func(universalLogger *UniversalLogger) {
		switch theme {
		case ThemeDark:
			universalLogger.scheme = darkScheme
		case ThemeLight:
			universalLogger.scheme = lightScheme
		default:
			universalLogger.scheme = getLoggerScheme()
		}
	}
}

// Публичные методы
func (standardLogger *StandardLogger) Write(p []byte) (n int, err error) {
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
func (universalLogger *UniversalLogger) Debug(message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelDebug, context.Background(), message, fields)
	case TypeText:
		universalLogger.writeText(LevelDebug, context.Background(), message, fields)
	}
}
func (universalLogger *UniversalLogger) DebugWithContext(context context.Context, message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelDebug, context, message, fields)
	case TypeText:
		universalLogger.writeText(LevelDebug, context, message, fields)
	}
}
func (universalLogger *UniversalLogger) Error(message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelError, context.Background(), message, fields)
	case TypeText:
		universalLogger.writeText(LevelError, context.Background(), message, fields)
	}
}
func (universalLogger *UniversalLogger) ErrorWithContext(context context.Context, message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelError, context, message, fields)
	case TypeText:
		universalLogger.writeText(LevelError, context, message, fields)
	}
}
func (universalLogger *UniversalLogger) Fatal(message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelFatal, context.Background(), message, fields)
	case TypeText:
		universalLogger.writeText(LevelFatal, context.Background(), message, fields)
	}
	osExit(1)
}
func (universalLogger *UniversalLogger) FatalWithContext(context context.Context, message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelFatal, context, message, fields)
	case TypeText:
		universalLogger.writeText(LevelFatal, context, message, fields)
	}
	osExit(1)
}
func (universalLogger *UniversalLogger) Info(message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelInfo, context.Background(), message, fields)
	case TypeText:
		universalLogger.writeText(LevelInfo, context.Background(), message, fields)
	}
}
func (universalLogger *UniversalLogger) InfoWithContext(context context.Context, message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelInfo, context, message, fields)
	case TypeText:
		universalLogger.writeText(LevelInfo, context, message, fields)
	}
}
func (universalLogger *UniversalLogger) Warn(message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelWarn, context.Background(), message, fields)
	case TypeText:
		universalLogger.writeText(LevelWarn, context.Background(), message, fields)
	}
}
func (universalLogger *UniversalLogger) WarnWithContext(context context.Context, message string, fields ...Field) {
	switch universalLogger.format {
	case TypeJson:
		universalLogger.writeJson(LevelWarn, context, message, fields)
	case TypeText:
		universalLogger.writeText(LevelWarn, context, message, fields)
	}
}
func (universalLogger *UniversalLogger) SetLevel(level TypeLevel) {
	universalLogger.level.Store(int32(level))
}
func (universalLogger *UniversalLogger) SetOutput(writer io.Writer) {
	universalLogger.mutex.Lock()
	defer universalLogger.mutex.Unlock()
	if universalLogger.async {
		if asyncWriter, ok := universalLogger.writer.(*asyncWriter); ok {
			asyncWriter.Close()
			time.Sleep(time.Millisecond)
		}
		universalLogger.writer = newAsyncWriter(writer, 10000)
	} else {
		universalLogger.writer = writer
	}
}
func (universalLogger *UniversalLogger) SetTheme(theme string) {
	universalLogger.mutex.Lock()
	defer universalLogger.mutex.Unlock()
	switch strings.ToLower(theme) {
	case "dark":
		universalLogger.scheme = darkScheme
	case "light":
		universalLogger.scheme = lightScheme
	default:
		universalLogger.scheme = getLoggerScheme()
	}
}
func (universalLogger *UniversalLogger) Sync() error {
	if !universalLogger.async {
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
func (universalLogger *UniversalLogger) Write(p []byte) (n int, err error) {
	universalLogger.mutex.RLock()
	writer := universalLogger.writer
	universalLogger.mutex.RUnlock()
	return writer.Write(p)
}
