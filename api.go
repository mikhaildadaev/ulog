// Copyright (c) 2026 Mikhail Dadaev
// All rights reserved.
//
// This source code is licensed under the MIT License found in the
// LICENSE file in the root directory of this source tree.
package ulog

import "strings"

// Публичные методы
func (loggerStandard *LoggerStandard) Debug(message string) {
	loggerStandard.setLog(LevelDebug, "[DEBUG] ", message)
}
func (loggerStandard *LoggerStandard) Debugf(format string, args ...any) {
	loggerStandard.setLogf(LevelDebug, "[DEBUG] ", format, args...)
}
func (loggerStandard *LoggerStandard) Error(message string) {
	loggerStandard.setLog(LevelError, "[ERROR] ", message)
}
func (loggerStandard *LoggerStandard) Errorf(format string, args ...any) {
	loggerStandard.setLogf(LevelError, "[ERROR] ", format, args...)
}
func (loggerStandard *LoggerStandard) Fatal(message string) {
	loggerStandard.setLog(LevelError, "[FATAL] ", message)
	osExit(1)
}
func (loggerStandard *LoggerStandard) Fatalf(format string, args ...any) {
	loggerStandard.setLogf(LevelError, "[FATAL] ", format, args...)
	osExit(1)
}
func (loggerStandard *LoggerStandard) Info(message string) {
	loggerStandard.setLog(LevelInfo, "[INFO] ", message)
}
func (loggerStandard *LoggerStandard) Infof(format string, args ...any) {
	loggerStandard.setLogf(LevelInfo, "[INFO] ", format, args...)
}
func (loggerStandard *LoggerStandard) Warn(message string) {
	loggerStandard.setLog(LevelWarn, "[WARN] ", message)
}
func (loggerStandard *LoggerStandard) Warnf(format string, args ...any) {
	loggerStandard.setLogf(LevelWarn, "[WARN] ", format, args...)
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
		loggerStandard.scheme = getLogerScheme()
	}
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
	if isError(p[start:end]) {
		return len(p), nil
	}
	message := string(p[start:end])
	loggerWriter.setMessage(message)
	return len(p), nil
}
