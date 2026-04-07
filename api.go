package ulog

import (
	"io"
	"strings"
)

// Публичные методы
func (loggerStandard *LoggerStandard) Debug(message string, fields ...Field) {
	switch loggerStandard.format {
	case JsonType:
		if len(fields) == 0 {
			loggerStandard.writeJson(LevelDebug, message)
		} else {
			loggerStandard.writeJsonFields(LevelDebug, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			loggerStandard.writeText(LevelDebug, message)
		} else {
			loggerStandard.writeTextFields(LevelDebug, message, fields)
		}
	}
}
func (loggerStandard *LoggerStandard) Error(message string, fields ...Field) {
	switch loggerStandard.format {
	case JsonType:
		if len(fields) == 0 {
			loggerStandard.writeJson(LevelError, message)
		} else {
			loggerStandard.writeJsonFields(LevelError, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			loggerStandard.writeText(LevelError, message)
		} else {
			loggerStandard.writeTextFields(LevelError, message, fields)
		}
	}
}
func (loggerStandard *LoggerStandard) Fatal(message string, fields ...Field) {
	switch loggerStandard.format {
	case JsonType:
		if len(fields) == 0 {
			loggerStandard.writeJson(LevelFatal, message)
		} else {
			loggerStandard.writeJsonFields(LevelFatal, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			loggerStandard.writeText(LevelFatal, message)
		} else {
			loggerStandard.writeTextFields(LevelFatal, message, fields)
		}
	}
	osExit(1)
}
func (loggerStandard *LoggerStandard) Info(message string, fields ...Field) {
	switch loggerStandard.format {
	case JsonType:
		if len(fields) == 0 {
			loggerStandard.writeJson(LevelInfo, message)
		} else {
			loggerStandard.writeJsonFields(LevelInfo, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			loggerStandard.writeText(LevelInfo, message)
		} else {
			loggerStandard.writeTextFields(LevelInfo, message, fields)
		}
	}
}
func (loggerStandard *LoggerStandard) Warn(message string, fields ...Field) {
	switch loggerStandard.format {
	case JsonType:
		if len(fields) == 0 {
			loggerStandard.writeJson(LevelWarn, message)
		} else {
			loggerStandard.writeJsonFields(LevelWarn, message, fields)
		}
	case TextType:
		if len(fields) == 0 {
			loggerStandard.writeText(LevelWarn, message)
		} else {
			loggerStandard.writeTextFields(LevelWarn, message, fields)
		}
	}
}
func (loggerStandard *LoggerStandard) SetLevel(level TypeLevel) {
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.level = level
}
func (loggerStandard *LoggerStandard) SetOutput(w io.Writer) {
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	if loggerStandard.async {
		if asyncWriter, ok := loggerStandard.writer.(*AsyncWriter); ok {
			go asyncWriter.Close()
		}
		loggerStandard.writer = NewAsyncWriter(w, 10000)
	} else {
		loggerStandard.writer = w
	}
}
func (loggerStandard *LoggerStandard) SetTheme(theme string) {
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	switch strings.ToLower(theme) {
	case "dark":
		loggerStandard.scheme = darkScheme
	case "light":
		loggerStandard.scheme = lightScheme
	default:
		loggerStandard.scheme = getLoggerScheme()
	}
}
func (loggerStandard *LoggerStandard) Sync() error {
	if !loggerStandard.async {
		return nil
	}
	loggerStandard.mutex.RLock()
	currentWriter := loggerStandard.writer
	loggerStandard.mutex.RUnlock()
	if asyncWriter, ok := currentWriter.(*AsyncWriter); ok {
		return asyncWriter.Close()
	}
	return nil
}
func (loggerStandard *LoggerStandard) Write(p []byte) (n int, err error) {
	loggerStandard.mutex.RLock()
	writer := loggerStandard.writer
	loggerStandard.mutex.RUnlock()
	return writer.Write(p)
}
func (loggerWriter *LoggerWriter) Write(p []byte) (n int, err error) {
	loggerWriter.mutex.Lock()
	defer loggerWriter.mutex.Unlock()
	start := 0
	end := len(p)
	for start < end && p[start] <= ' ' {
		start++
	}
	for end > start && p[end-1] <= ' ' {
		end--
	}
	if start >= end {
		return len(p), nil
	}
	if isIgnoredError(p[start:end]) {
		return len(p), nil
	}
	message := string(p[start:end])
	loggerWriter.setMessage(message)
	return len(p), nil
}
