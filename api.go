package ulog

import (
	"io"
	"log"
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
func (loggerStandard *LoggerStandard) SetOutput(writer io.Writer) {
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	if loggerStandard.asyncWriter != nil {
		loggerStandard.asyncWriter.Close()
	}
	bufferSize := 10000
	if loggerStandard.asyncWriter != nil {
		bufferSize = loggerStandard.asyncWriter.limit
	}
	loggerStandard.asyncWriter = NewAsyncWriter(writer, bufferSize)
	loggerStandard.Logger = log.New(loggerStandard.asyncWriter, "", 0)
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
	return loggerStandard.asyncWriter.Close()
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
