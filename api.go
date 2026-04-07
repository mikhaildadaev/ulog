package ulog

import (
	"io"
	"strings"
	"time"
)

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
	case JsonType:
		if len(fields) == 0 {
			universalLogger.writeJson(LevelDebug, message)
		} else {
			universalLogger.writeJsonFields(LevelDebug, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			universalLogger.writeText(LevelDebug, message)
		} else {
			universalLogger.writeTextFields(LevelDebug, message, fields)
		}
	}
}
func (universalLogger *UniversalLogger) Error(message string, fields ...Field) {
	switch universalLogger.format {
	case JsonType:
		if len(fields) == 0 {
			universalLogger.writeJson(LevelError, message)
		} else {
			universalLogger.writeJsonFields(LevelError, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			universalLogger.writeText(LevelError, message)
		} else {
			universalLogger.writeTextFields(LevelError, message, fields)
		}
	}
}
func (universalLogger *UniversalLogger) Fatal(message string, fields ...Field) {
	switch universalLogger.format {
	case JsonType:
		if len(fields) == 0 {
			universalLogger.writeJson(LevelFatal, message)
		} else {
			universalLogger.writeJsonFields(LevelFatal, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			universalLogger.writeText(LevelFatal, message)
		} else {
			universalLogger.writeTextFields(LevelFatal, message, fields)
		}
	}
	osExit(1)
}
func (universalLogger *UniversalLogger) Info(message string, fields ...Field) {
	switch universalLogger.format {
	case JsonType:
		if len(fields) == 0 {
			universalLogger.writeJson(LevelInfo, message)
		} else {
			universalLogger.writeJsonFields(LevelInfo, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			universalLogger.writeText(LevelInfo, message)
		} else {
			universalLogger.writeTextFields(LevelInfo, message, fields)
		}
	}
}
func (universalLogger *UniversalLogger) Warn(message string, fields ...Field) {
	switch universalLogger.format {
	case JsonType:
		if len(fields) == 0 {
			universalLogger.writeJson(LevelWarn, message)
		} else {
			universalLogger.writeJsonFields(LevelWarn, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			universalLogger.writeText(LevelWarn, message)
		} else {
			universalLogger.writeTextFields(LevelWarn, message, fields)
		}
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
