package ulog

import "strings"

// Публичные методы
func (loggerStandard *LoggerStandard) Debug(message string) {
	loggerStandard.setLog(LevelDebug, message)
}
func (loggerStandard *LoggerStandard) Debugf(format string, args ...any) {
	loggerStandard.setLogf(LevelDebug, format, args...)
}
func (loggerStandard *LoggerStandard) Error(message string) {
	loggerStandard.setLog(LevelError, message)
}
func (loggerStandard *LoggerStandard) Errorf(format string, args ...any) {
	loggerStandard.setLogf(LevelError, format, args...)
}
func (loggerStandard *LoggerStandard) Fatal(message string) {
	loggerStandard.setLog(LevelError, message)
	osExit(1)
}
func (loggerStandard *LoggerStandard) Fatalf(format string, args ...any) {
	loggerStandard.setLogf(LevelError, format, args...)
	osExit(1)
}
func (loggerStandard *LoggerStandard) Info(message string) {
	loggerStandard.setLog(LevelInfo, message)
}
func (loggerStandard *LoggerStandard) Infof(format string, args ...any) {
	loggerStandard.setLogf(LevelInfo, format, args...)
}
func (loggerStandard *LoggerStandard) Warn(message string) {
	loggerStandard.setLog(LevelWarn, message)
}
func (loggerStandard *LoggerStandard) Warnf(format string, args ...any) {
	loggerStandard.setLogf(LevelWarn, format, args...)
}
func (loggerStandard *LoggerStandard) SetLevel(level int) {
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.level = level
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
