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
	cache  sync.Map
	caller bool
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
		caller: true,
		level:  getLoggerLevel(),
		Logger: log.New(os.Stderr, "", log.LstdFlags|log.Lmicroseconds),
		scheme: getLoggerScheme(),
	}
}
func NewErrorLog(logger Logger) *log.Logger {
	return log.New(
		&LoggerWriter{
			level:  LevelError,
			logger: logger,
			scheme: getLoggerScheme(),
		},
		"",
		log.LstdFlags|log.Lmicroseconds,
	)
}
func NewWithWriter(logger Logger, level int) io.Writer {
	return &LoggerWriter{
		level:  level,
		logger: logger,
		scheme: getLoggerScheme(),
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

// Приватные переменные
var ignoredErrors = [][]byte{
	[]byte("EOF"),
	[]byte("TLS handshake error"),
	[]byte("connection refused"),
	[]byte("timeout"),
	[]byte("broken pipe"),
	[]byte("i/o timeout"),
	[]byte("no such host"),
}

// Приватные структуры
type colorScheme struct {
	colorCyan   string
	colorGreen  string
	colorPurple string
	colorRed    string
	colorYellow string
	caller      string
	message     string
	prefix      string
	reset       string
}

// Приватные переменные
var (
	dataPool = sync.Pool{
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
		caller:      colorDarkBlue,
		message:     colorDarkWhite,
		reset:       colorReset,
	}
	lightScheme = colorScheme{
		colorRed:    colorLightRed,
		colorGreen:  colorLightGreen,
		colorYellow: colorLightYellow,
		colorPurple: colorLightPurple,
		colorCyan:   colorLightCyan,
		caller:      colorLightBlue,
		message:     colorLightBlack,
		reset:       colorReset,
	}
	osExit   = os.Exit
	timePool = sync.Pool{
		New: func() any {
			return make([]byte, 0, 26)
		},
	}
)

// Приватные функции
func formatCaller(dataBuf *bytes.Buffer, scheme colorScheme, caller string) {
	dataBuf.WriteString(scheme.caller)
	dataBuf.WriteString(caller)
	dataBuf.WriteByte(' ')
}
func formatData(dataBuf *bytes.Buffer, scheme colorScheme, message string) {
	dataBuf.WriteString(scheme.message)
	dataBuf.WriteString(message)
	dataBuf.WriteString(scheme.reset)
	dataBuf.WriteByte('\n')
}
func formatDataf(dataBuf *bytes.Buffer, scheme colorScheme, format string, args []any) {
	argIdx := 0
	for i := 0; i < len(format); i++ {
		ch := format[i]
		if ch == '%' && i+1 < len(format) {
			i++
			switch format[i] {
			case 's':
				if argIdx < len(args) {
					if s, ok := args[argIdx].(string); ok {
						dataBuf.WriteString(s)
					} else {
						fmt.Fprint(dataBuf, args[argIdx])
					}
				}
				argIdx++
			case 'd':
				if argIdx < len(args) {
					switch v := args[argIdx].(type) {
					case int:
						dataBuf.WriteString(strconv.Itoa(v))
					case int64:
						dataBuf.WriteString(strconv.FormatInt(v, 10))
					case uint:
						dataBuf.WriteString(strconv.FormatUint(uint64(v), 10))
					default:
						fmt.Fprint(dataBuf, v)
					}
				}
				argIdx++
			case 'f':
				if argIdx < len(args) {
					switch v := args[argIdx].(type) {
					case float64:
						dataBuf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
					case float32:
						dataBuf.WriteString(strconv.FormatFloat(float64(v), 'f', -1, 32))
					default:
						fmt.Fprint(dataBuf, v)
					}
				}
				argIdx++
			case 'v':
				if argIdx < len(args) {
					fmt.Fprint(dataBuf, args[argIdx])
				}
				argIdx++
			case '%':
				dataBuf.WriteByte('%')
			default:
				dataBuf.WriteByte(ch)
				dataBuf.WriteByte(format[i])
			}
		} else {
			dataBuf.WriteByte(ch)
		}
	}
	dataBuf.WriteString(scheme.reset)
	dataBuf.WriteByte('\n')
}
func formatPrefix(dataBuf *bytes.Buffer, scheme colorScheme, prefix string) {
	dataBuf.WriteString(scheme.prefix)
	dataBuf.WriteString(prefix)
	dataBuf.WriteByte(' ')
}
func formatTime(dataBuf *bytes.Buffer, time time.Time) {
	timeBuf := timePool.Get().([]byte)
	timeBuf = time.AppendFormat(timeBuf[:0], "2006/01/02 15:04:05.000000")
	dataBuf.Write(timeBuf)
	timePool.Put(timeBuf)
	dataBuf.WriteByte(' ')
}
func getLoggerCaller(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[i+1:]
		}
	}
	return path
}
func getLoggerScheme() colorScheme {
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
func getLoggerLevel() int {
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
func isIgnoredError(data []byte) bool {
	for _, err := range ignoredErrors {
		if bytes.Contains(data, err) {
			return true
		}
	}
	return false
}

// Приватные методы
func (loggerStandard *LoggerStandard) getCaller() string {
	if !loggerStandard.caller {
		return "disabled:0"
	}
	pc, file, line, _ := runtime.Caller(2)
	if val, ok := loggerStandard.cache.Load(pc); ok {
		return val.(string)
	}
	caller := getLoggerCaller(file) + ":" + strconv.Itoa(line)
	loggerStandard.cache.Store(pc, caller)
	return caller
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
	caller := loggerStandard.getCaller()
	scheme := loggerStandard.getScheme()
	scheme.prefix = loggerStandard.getColor(level, scheme)
	time := time.Now()
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	formatTime(dataBuf, time)
	formatPrefix(dataBuf, scheme, prefix)
	formatCaller(dataBuf, scheme, caller)
	formatData(dataBuf, scheme, message)
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.Writer().Write(dataBuf.Bytes())
}
func (loggerStandard *LoggerStandard) setLogf(level int, prefix, format string, args ...any) {
	if loggerStandard.getLevel() > level {
		return
	}
	caller := loggerStandard.getCaller()
	scheme := loggerStandard.getScheme()
	scheme.prefix = loggerStandard.getColor(level, scheme)
	time := time.Now()
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	formatTime(dataBuf, time)
	formatPrefix(dataBuf, scheme, prefix)
	formatCaller(dataBuf, scheme, caller)
	formatDataf(dataBuf, scheme, format, args)
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.Writer().Write(dataBuf.Bytes())
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
