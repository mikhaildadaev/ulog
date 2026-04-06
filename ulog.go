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

// Публичные типы
type FieldType uint8

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
const (
	StringType FieldType = iota
	IntType
	Int64Type
	BoolType
	Float64Type
	DurationType
	TimeType
)

// Публичные интерфейсы
type Logger interface {
	Debug(message string, fields ...Field)
	Error(message string, fields ...Field)
	Fatal(message string, fields ...Field)
	Info(message string, fields ...Field)
	Warn(message string, fields ...Field)
	SetLevel(level int)
	SetOutput(writer io.Writer)
	SetTheme(theme string)
	Sync() error
}

// Публичные структуры
type Field struct {
	typ   FieldType
	key   string
	value any
}
type LoggerStandard struct {
	*log.Logger
	asyncWriter *AsyncWriter
	cache       sync.Map
	caller      bool
	level       int
	mutex       sync.RWMutex
	scheme      colorScheme
}
type AsyncWriter struct {
	ch     chan []byte
	limit  int
	wg     sync.WaitGroup
	writer io.Writer
}
type LoggerWriter struct {
	level  int
	logger Logger
	mutex  sync.Mutex
	scheme colorScheme
}

// Публичные конструкторы
func Bool(key string, value bool) Field {
	return Field{
		typ:   BoolType,
		key:   key,
		value: value,
	}
}
func Duration(key string, value time.Duration) Field {
	return Field{
		typ:   DurationType,
		key:   key,
		value: value,
	}
}
func Float64(key string, value float64) Field {
	return Field{
		typ:   Float64Type,
		key:   key,
		value: value,
	}
}
func Int(key string, value int) Field {
	return Field{
		typ:   IntType,
		key:   key,
		value: value,
	}
}
func Int64(key string, value int64) Field {
	return Field{
		typ:   Int64Type,
		key:   key,
		value: value,
	}
}
func String(key, value string) Field {
	return Field{
		typ:   StringType,
		key:   key,
		value: value,
	}
}
func Time(key string, value time.Time) Field {
	return Field{
		typ:   TimeType,
		key:   key,
		value: value,
	}
}
func New() Logger {
	asyncWriter := NewAsyncWriter(os.Stderr, 10000)
	return &LoggerStandard{
		asyncWriter: asyncWriter,
		caller:      true,
		level:       getLoggerLevel(),
		Logger:      log.New(asyncWriter, "", 0),
		scheme:      getLoggerScheme(),
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
func NewAsyncWriter(writer io.Writer, bufferSize int) *AsyncWriter {
	asyncWriter := &AsyncWriter{
		ch:     make(chan []byte, bufferSize),
		limit:  bufferSize,
		writer: writer,
	}
	asyncWriter.wg.Add(1)
	go asyncWriter.run()
	return asyncWriter
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
func formatDataf(dataBuf *bytes.Buffer, scheme colorScheme, fields []Field) {
	for _, field := range fields {
		dataBuf.WriteByte(' ')
		dataBuf.WriteString(field.key)
		dataBuf.WriteByte('=')
		switch field.typ {
		case BoolType:
			dataBuf.WriteString(strconv.FormatBool(field.value.(bool)))
		case DurationType:
			dataBuf.WriteString(field.value.(time.Duration).String())
		case Float64Type:
			dataBuf.WriteString(strconv.FormatFloat(field.value.(float64), 'f', -1, 64))
		case IntType:
			dataBuf.WriteString(strconv.Itoa(field.value.(int)))
		case Int64Type:
			dataBuf.WriteString(strconv.FormatInt(field.value.(int64), 10))
		case StringType:
			dataBuf.WriteString(field.value.(string))
		case TimeType:
			dataBuf.Write(field.value.(time.Time).AppendFormat(nil, "2006-01-02T15:04:05.000Z07:00"))
		}
	}
	dataBuf.WriteString(scheme.reset)
	dataBuf.WriteByte('\n')
}
func formatPrefix(dataBuf *bytes.Buffer, scheme colorScheme, level int) {
	switch level {
	case LevelDebug:
		dataBuf.WriteString(scheme.colorCyan)
		dataBuf.WriteString("[DEBUG]")
	case LevelError:
		dataBuf.WriteString(scheme.colorRed)
		dataBuf.WriteString("[ERROR]")
	case LevelFatal:
		dataBuf.WriteString(scheme.colorPurple)
		dataBuf.WriteString("[FATAL]")
	case LevelInfo:
		dataBuf.WriteString(scheme.colorGreen)
		dataBuf.WriteString("[INFO]")
	case LevelWarn:
		dataBuf.WriteString(scheme.colorYellow)
		dataBuf.WriteString("[WARN]")
	}
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
func (asyncWriter *AsyncWriter) run() {
	defer asyncWriter.wg.Done()
	for buf := range asyncWriter.ch {
		asyncWriter.writer.Write(buf)
	}
}
func (asyncWriter *AsyncWriter) Write(p []byte) (n int, err error) {
	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case asyncWriter.ch <- buf:
		return len(p), nil
	default:
		asyncWriter.ch <- buf
		return len(p), nil
	}
}
func (asyncWriter *AsyncWriter) Close() error {
	close(asyncWriter.ch)
	asyncWriter.wg.Wait()
	return nil
}
func (loggerStandard *LoggerStandard) Close() error {
	return loggerStandard.asyncWriter.Close()
}
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
func (loggerStandard *LoggerStandard) setLog(level int, message string) {
	if loggerStandard.getLevel() > level {
		return
	}
	caller := loggerStandard.getCaller()
	scheme := loggerStandard.getScheme()
	time := time.Now()
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	formatTime(dataBuf, time)
	formatPrefix(dataBuf, scheme, level)
	formatCaller(dataBuf, scheme, caller)
	formatData(dataBuf, scheme, message)
	loggerStandard.mutex.Lock()
	defer loggerStandard.mutex.Unlock()
	loggerStandard.Writer().Write(dataBuf.Bytes())
}
func (loggerStandard *LoggerStandard) setLogf(level int, message string, fields []Field) {
	if loggerStandard.getLevel() > level {
		return
	}
	caller := loggerStandard.getCaller()
	scheme := loggerStandard.getScheme()
	time := time.Now()
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	formatTime(dataBuf, time)
	formatPrefix(dataBuf, scheme, level)
	formatCaller(dataBuf, scheme, caller)
	formatDataf(dataBuf, scheme, fields)
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
