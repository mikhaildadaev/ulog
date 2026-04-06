package ulog

import (
	"io"
	"log"
	"strings"
)

// Публичные методы
func (loggerStandard *LoggerStandard) Debug(message string, fields ...Field) {
	if len(fields) == 0 {
		loggerStandard.setLog(LevelDebug, message)
	} else {
		loggerStandard.setLogf(LevelDebug, message, fields)
	}
}
func (loggerStandard *LoggerStandard) Error(message string, fields ...Field) {
	if len(fields) == 0 {
		loggerStandard.setLog(LevelError, message)
	} else {
		loggerStandard.setLogf(LevelError, message, fields)
	}
}
func (loggerStandard *LoggerStandard) Fatal(message string, fields ...Field) {
	if len(fields) == 0 {
		loggerStandard.setLog(LevelFatal, message)
	} else {
		loggerStandard.setLogf(LevelFatal, message, fields)
	}
	osExit(1)
}
func (loggerStandard *LoggerStandard) Info(message string, fields ...Field) {
	if len(fields) == 0 {
		loggerStandard.setLog(LevelInfo, message)
	} else {
		loggerStandard.setLogf(LevelInfo, message, fields)
	}
}
func (loggerStandard *LoggerStandard) Warn(message string, fields ...Field) {
	if len(fields) == 0 {
		loggerStandard.setLog(LevelWarn, message)
	} else {
		loggerStandard.setLogf(LevelWarn, message, fields)
	}
}
func (loggerStandard *LoggerStandard) SetLevel(level int) {
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
