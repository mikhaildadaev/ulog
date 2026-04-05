// Copyright (c) 2026 Mikhail Dadaev
// All rights reserved.
//
// This source code is licensed under the MIT License found in the
// LICENSE file in the root directory of this source tree.
package ulog

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// Публичные константы
const (
	Author  = "Mikhail Dadaev"
	Version = "1.26.3"
)
const (
	ErrorEOF          = "EOF"
	ErrorTLSHandshake = "TLS handshake error"
)
const (
	LevelDebug = iota // 0 - отладочная информация
	LevelInfo         // 1 - штатные операции
	LevelWarn         // 2 - нештатные ситуации, но не ошибки
	LevelError        // 3 - ошибки, требующие внимания
	LevelFatal        // 4 - критические ошибки (с остановкой приложения)
)

// Публичные интерфейсы
type Logger interface {
	Debug(message string)
	Debugf(format string, args ...any)
	Error(message string)
	Errorf(format string, args ...any)
	Fatal(message string)
	Fatalf(format string, args ...any)
	Info(message string)
	Infof(format string, args ...any)
	Warn(message string)
	Warnf(format string, args ...any)
	SetLevel(level int)
	SetTheme(theme string)
}

// Публичные структуры
type LoggerStandard struct {
	*log.Logger
	mutex  sync.RWMutex
	level  int
	scheme colorScheme
}
type LoggerWriter struct {
	logger Logger
	mutex  sync.Mutex
	level  int
}

// Публичные конструкторы
func New() Logger {
	return &LoggerStandard{
		level:  getLogerLevel(),
		Logger: log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds),
		scheme: getColorScheme(),
	}
}
func NewErrorLog(logger Logger) *log.Logger {
	return log.New(
		&LoggerWriter{
			logger: logger,
			level:  LevelError,
		},
		"",
		log.LstdFlags|log.Lmicroseconds,
	)
}
func NewWithWriter(logger Logger, level int) io.Writer {
	return &LoggerWriter{
		level:  level,
		logger: logger,
	}
}

// Публичные функции
func GetAuthor() string {
	return Author
}
func GetCopyright() string {
	Copyright := fmt.Sprintf("Copyright © 2022-%d %s. All rights reserved.", time.Now().Year(), Author)
	return Copyright
}
func GetVersion() string {
	return Version
}

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
		loggerStandard.scheme = getColorScheme()
	}
}
func (loggerWriter *LoggerWriter) Write(p []byte) (n int, err error) {
	loggerWriter.mutex.Lock()
	defer loggerWriter.mutex.Unlock()
	msg := strings.TrimSpace(string(p))
	switch {
	case isHandshakeError(msg):
		return len(p), nil
	default:
		loggerWriter.setMessage(msg)
		return len(p), nil
	}
}

// Приватные константы
const (
	colorReset = "\033[0m"
	// Темная тема (ANSI коды 90-97)
	colorDarkRed    = "\033[91m"
	colorDarkGreen  = "\033[92m"
	colorDarkYellow = "\033[93m"
	colorDarkBlue   = "\033[94m"
	colorDarkPurple = "\033[95m"
	colorDarkCyan   = "\033[96m"
	colorDarkWhite  = "\033[97m"
	// Светлая тема (ANSI коды 30-37)
	colorLightBlack  = "\033[30m"
	colorLightRed    = "\033[31m"
	colorLightGreen  = "\033[32m"
	colorLightYellow = "\033[33m"
	colorLightBlue   = "\033[34m"
	colorLightPurple = "\033[35m"
	colorLightCyan   = "\033[36m"
)

// Приватные структуры
type colorScheme struct {
	colorCyan   string
	colorGreen  string
	colorPurple string
	colorRed    string
	colorYellow string
	fileLine    string
	message     string
	reset       string
}

// Приватные переменные
var (
	bufPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
	darkScheme = colorScheme{
		colorRed:    colorDarkRed,
		colorGreen:  colorDarkGreen,
		colorYellow: colorDarkYellow,
		colorPurple: colorDarkPurple,
		colorCyan:   colorDarkCyan,
		fileLine:    colorDarkBlue,
		message:     colorDarkWhite,
		reset:       colorReset,
	}
	lightScheme = colorScheme{
		colorRed:    colorLightRed,
		colorGreen:  colorLightGreen,
		colorYellow: colorLightYellow,
		colorPurple: colorLightPurple,
		colorCyan:   colorLightCyan,
		fileLine:    colorLightBlue,
		message:     colorLightBlack,
		reset:       colorReset,
	}
	osExit = os.Exit
)

// Приватные функции
func getColorScheme() colorScheme {
	switch strings.ToLower(os.Getenv("TERM_THEME")) {
	case "dark":
		return darkScheme
	case "light":
		return lightScheme
	}
	if os.Getenv("COLORFGBG") != "" {
		parts := strings.Split(os.Getenv("COLORFGBG"), ";")
		if len(parts) >= 2 {
			bg, _ := strconv.Atoi(parts[1])
			if bg < 8 {
				return darkScheme
			}
			return lightScheme
		}
	}
	return darkScheme
}
func getLogerLevel() int {
	switch strings.ToLower(os.Getenv("LOG_LEVEL")) {
	case "debug":
		return LevelDebug
	case "info":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error":
		return LevelError
	case "fatal":
		return LevelFatal
	}
	if os.Getenv("DEBUG") == "true" {
		return LevelDebug
	}
	return LevelInfo
}
func isHandshakeError(message string) bool {
	lowerMessage := strings.ToLower(message)
	return strings.Contains(lowerMessage, strings.ToLower(ErrorEOF)) && strings.Contains(lowerMessage, strings.ToLower(ErrorTLSHandshake))
}

// Приватные методы
func (loggerStandard *LoggerStandard) getLevel() int {
	loggerStandard.mutex.RLock()
	defer loggerStandard.mutex.RUnlock()
	return loggerStandard.level
}
func (loggerStandard *LoggerStandard) getColor(level int, scheme colorScheme) string {
	switch level {
	case LevelDebug:
		return scheme.colorCyan
	case LevelError:
		return scheme.colorRed
	case LevelFatal:
		return scheme.colorPurple
	case LevelInfo:
		return scheme.colorGreen
	case LevelWarn:
		return scheme.colorYellow
	default:
		return scheme.colorGreen
	}
}
func (loggerStandard *LoggerStandard) getScheme() colorScheme {
	loggerStandard.mutex.RLock()
	defer loggerStandard.mutex.RUnlock()
	return loggerStandard.scheme
}
func (loggerStandard *LoggerStandard) setLog(level int, prefix, message string) {
	if loggerStandard.getLevel() > level {
		return
	}
	_, filename, linenumber, ok := runtime.Caller(2)
	if !ok {
		filename = "unknown"
		linenumber = 0
	} else {
		filename = filepath.Base(filename)
	}
	scheme := loggerStandard.getScheme()
	colorPrefix := loggerStandard.getColor(level, scheme)
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.Printf("%s%s%s%s[%d] %s%s%s", colorPrefix, prefix, scheme.fileLine, filename, linenumber, scheme.message, message, scheme.reset)
}
func (loggerStandard *LoggerStandard) setLogf(level int, prefix, format string, args ...any) {
	if loggerStandard.getLevel() > level {
		return
	}
	_, filename, linenumber, ok := runtime.Caller(2)
	if !ok {
		filename = "unknown"
		linenumber = 0
	} else {
		filename = filepath.Base(filename)
	}
	scheme := loggerStandard.getScheme()
	colorPrefix := loggerStandard.getColor(level, scheme)
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	fmt.Fprintf(buf, format, args...)
	message := buf.String()
	message = strings.ReplaceAll(message, "%", "%%")
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.Printf("%s%s%s%s[%d] %s%s%s", colorPrefix, prefix, scheme.fileLine, filename, linenumber, scheme.message, message, scheme.reset)
}
func (loggerWriter *LoggerWriter) setMessage(msg string) {
	switch loggerWriter.level {
	case LevelDebug:
		loggerWriter.logger.Debug(msg)
	case LevelInfo:
		loggerWriter.logger.Info(msg)
	case LevelWarn:
		loggerWriter.logger.Warn(msg)
	case LevelError:
		loggerWriter.logger.Error(msg)
	case LevelFatal:
		loggerWriter.logger.Fatal(msg)
	default:
		loggerWriter.logger.Info(msg)
	}
}
