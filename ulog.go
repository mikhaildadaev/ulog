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
	Version = "1.26.5"
)
const (
	LevelDebug = iota // 0 - отладочная информация
	LevelInfo         // 1 - штатные операции
	LevelWarn         // 2 - нештатные ситуации, но не ошибки
	LevelError        // 3 - ошибки, требующие внимания
	LevelFatal        // 4 - критические ошибки (с остановкой приложения)
)

// Публичные переменные
var (
	ErrorEOF          = []byte("EOF")
	ErrorTLSHandshake = []byte("TLS handshake error")
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
	level  int
	mutex  sync.RWMutex
	scheme colorScheme
}
type LoggerWriter struct {
	level  int
	logger Logger
	mutex  sync.Mutex
	scheme colorScheme
}

// Публичные конструкторы
func New() Logger {
	return &LoggerStandard{
		level:  getLogerLevel(),
		Logger: log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds),
		scheme: getLogerScheme(),
	}
}
func NewErrorLog(logger Logger) *log.Logger {
	return log.New(
		&LoggerWriter{
			level:  LevelError,
			logger: logger,
			scheme: getLogerScheme(),
		},
		"",
		log.LstdFlags|log.Lmicroseconds,
	)
}
func NewWithWriter(logger Logger, level int) io.Writer {
	return &LoggerWriter{
		level:  level,
		logger: logger,
		scheme: getLogerScheme(),
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
	prefix      string
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
func formatMessage(buf *bytes.Buffer, format string, args []any) {
	argIdx := 0
	for i := 0; i < len(format); i++ {
		ch := format[i]
		if ch == '%' && i+1 < len(format) {
			i++
			switch format[i] {
			case 's':
				if argIdx < len(args) {
					if s, ok := args[argIdx].(string); ok {
						buf.WriteString(s)
					} else {
						fmt.Fprint(buf, args[argIdx])
					}
				}
				argIdx++
			case 'd':
				if argIdx < len(args) {
					switch v := args[argIdx].(type) {
					case int:
						buf.WriteString(strconv.Itoa(v))
					case int64:
						buf.WriteString(strconv.FormatInt(v, 10))
					case uint:
						buf.WriteString(strconv.FormatUint(uint64(v), 10))
					default:
						fmt.Fprint(buf, v)
					}
				}
				argIdx++
			case 'f':
				if argIdx < len(args) {
					switch v := args[argIdx].(type) {
					case float64:
						buf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
					case float32:
						buf.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
					default:
						fmt.Fprint(buf, v)
					}
				}
				argIdx++
			case 'v':
				if argIdx < len(args) {
					fmt.Fprint(buf, args[argIdx])
				}
				argIdx++
			case '%':
				buf.WriteByte('%')
			default:
				buf.WriteByte(ch)
				buf.WriteByte(format[i])
			}
		} else {
			buf.WriteByte(ch)
		}
	}
}
func getLogerScheme() colorScheme {
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
func isError(data []byte) bool {
	return bytes.Contains(data, ErrorEOF) && bytes.Contains(data, ErrorTLSHandshake)
}

// Приватные методы
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
func (loggerStandard *LoggerStandard) getLevel() int {
	loggerStandard.mutex.RLock()
	defer loggerStandard.mutex.RUnlock()
	return loggerStandard.level
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
	now := time.Now()
	scheme := loggerStandard.getScheme()
	scheme.prefix = loggerStandard.getColor(level, scheme)
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	buf.WriteString(now.Format("2006/01/02 15:04:05.000000"))
	buf.WriteString(" ")
	buf.WriteString(scheme.prefix)
	buf.WriteString(prefix)
	buf.WriteString(scheme.fileLine)
	buf.WriteString(filename)
	buf.WriteString("[")
	buf.WriteString(strconv.Itoa(linenumber))
	buf.WriteString("] ")
	buf.WriteString(scheme.message)
	buf.WriteString(message)
	buf.WriteString(scheme.reset)
	buf.WriteByte('\n')
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.Writer().Write(buf.Bytes())
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
	now := time.Now()
	scheme := loggerStandard.getScheme()
	scheme.prefix = loggerStandard.getColor(level, scheme)
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Reset()
	defer bufPool.Put(buf)
	buf.WriteString(now.Format("2006/01/02 15:04:05.000000"))
	buf.WriteByte(' ')
	buf.WriteString(scheme.prefix)
	buf.WriteString(prefix)
	buf.WriteString(scheme.fileLine)
	buf.WriteString(filename)
	buf.WriteByte('[')
	buf.WriteString(strconv.Itoa(linenumber))
	buf.WriteString("] ")
	buf.WriteString(scheme.message)
	formatMessage(buf, format, args)
	buf.WriteString(scheme.reset)
	buf.WriteByte('\n')
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.Writer().Write(buf.Bytes())
}
func (loggerWriter *LoggerWriter) setMessage(message string) {
	switch loggerWriter.level {
	case LevelDebug:
		loggerWriter.logger.Debug(message)
	case LevelInfo:
		loggerWriter.logger.Info(message)
	case LevelWarn:
		loggerWriter.logger.Warn(message)
	case LevelError:
		loggerWriter.logger.Error(message)
	case LevelFatal:
		loggerWriter.logger.Fatal(message)
	default:
		loggerWriter.logger.Info(message)
	}
}
