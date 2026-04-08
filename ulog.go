// Copyright (c) 2026 Mikhail Dadaev
// All rights reserved.
//
// This source code is licensed under the MIT License found in the
// LICENSE file in the root directory of this source tree.
package ulog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Публичные типы
type TypeLevel int
type TypeField int
type TypeFormat int
type TypeTheme int

// Публичные константы
const (
	Author  = "Mikhail Dadaev"
	Version = "1.26.5"
)
const (
	LevelDebug TypeLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)
const (
	TypeBool TypeField = iota
	TypeBools
	TypeDuration
	TypeDurations
	TypeFloat64
	TypeFloats64
	TypeInt
	TypeInts
	TypeInt64
	TypeInts64
	TypeString
	TypeStrings
	TypeTime
	TypeTimes
)
const (
	TypeJson TypeFormat = iota
	TypeText
)
const (
	ThemeDark TypeTheme = iota
	ThemeLight
)

// Публичные интерфейсы
type Logger interface {
	Debug(message string, fields ...Field)
	DebugWithContext(ctx context.Context, msg string, fields ...Field)
	Error(message string, fields ...Field)
	ErrorWithContext(ctx context.Context, msg string, fields ...Field)
	Fatal(message string, fields ...Field)
	FatalWithContext(ctx context.Context, msg string, fields ...Field)
	Info(message string, fields ...Field)
	InfoWithContext(ctx context.Context, msg string, fields ...Field)
	Warn(message string, fields ...Field)
	WarnWithContext(ctx context.Context, msg string, fields ...Field)
	SetLevel(level TypeLevel)
	SetOutput(writer io.Writer)
	SetTheme(theme string)
	Sync() error
}

// Публичные структуры
type Field struct {
	keyName        string
	valueBool      bool
	valueBools     []bool
	valueDuration  time.Duration
	valueDurations []time.Duration
	valueInt       int
	valueInts      []int
	valueInt64     int64
	valueInts64    []int64
	valueFloat64   float64
	valueFloats64  []float64
	valueString    string
	valueStrings   []string
	valueTime      time.Time
	valueTimes     []time.Time
	valueType      TypeField
}
type StandardLogger struct {
	flags  int
	level  atomic.Int32
	logger Logger
	mutex  sync.Mutex
	scheme colorScheme
}
type UniversalLogger struct {
	async  bool
	cache  sync.Map
	format TypeFormat
	level  atomic.Int32
	mutex  sync.RWMutex
	scheme colorScheme
	writer io.Writer
}
type OptionLogger func(*UniversalLogger)

// Публичные конструкторы
func Bool(keyName string, valueBool bool) Field {
	return Field{
		valueType: TypeBool,
		keyName:   keyName,
		valueBool: valueBool,
	}
}
func Bools(keyName string, valueBools []bool) Field {
	return Field{
		valueType:  TypeBool,
		keyName:    keyName,
		valueBools: valueBools,
	}
}
func Duration(keyName string, valueDuration time.Duration) Field {
	return Field{
		valueType:     TypeDuration,
		keyName:       keyName,
		valueDuration: valueDuration,
	}
}
func Durations(keyName string, valueDurations []time.Duration) Field {
	return Field{
		valueType:      TypeDuration,
		keyName:        keyName,
		valueDurations: valueDurations,
	}
}
func Err(err error) Field {
	if err == nil {
		return Field{
			valueType:   TypeString,
			keyName:     "error",
			valueString: "nil",
		}
	}
	return Field{
		valueType:   TypeString,
		keyName:     "error",
		valueString: err.Error(),
	}
}
func Errs(errs []error) Field {
	values := make([]string, len(errs))
	for i, err := range errs {
		if err == nil {
			values[i] = "nil"
		} else {
			values[i] = err.Error()
		}
	}
	return Field{
		valueType:    TypeString,
		keyName:      "errors",
		valueStrings: values,
	}
}
func Float64(keyName string, valueFloat64 float64) Field {
	return Field{
		valueType:    TypeFloat64,
		keyName:      keyName,
		valueFloat64: valueFloat64,
	}
}
func Floats64(keyName string, valueFloats64 []float64) Field {
	return Field{
		valueType:     TypeFloat64,
		keyName:       keyName,
		valueFloats64: valueFloats64,
	}
}
func Int(keyName string, valueInt int) Field {
	return Field{
		valueType: TypeInt,
		keyName:   keyName,
		valueInt:  valueInt,
	}
}
func Ints(keyName string, valueInts []int) Field {
	return Field{
		valueType: TypeInt,
		keyName:   keyName,
		valueInts: valueInts,
	}
}
func Int64(keyName string, valueInt64 int64) Field {
	return Field{
		valueType:  TypeInt,
		keyName:    keyName,
		valueInt64: valueInt64,
	}
}
func Ints64(keyName string, valueInts64 []int64) Field {
	return Field{
		valueType:   TypeInt,
		keyName:     keyName,
		valueInts64: valueInts64,
	}
}
func String(keyName string, valueString string) Field {
	return Field{
		valueType:   TypeString,
		keyName:     keyName,
		valueString: valueString,
	}
}
func Strings(keyName string, valueStrings []string) Field {
	return Field{
		valueType:    TypeString,
		keyName:      keyName,
		valueStrings: valueStrings,
	}
}
func Time(keyName string, valueTime time.Time) Field {
	return Field{
		valueType: TypeTime,
		keyName:   keyName,
		valueTime: valueTime,
	}
}
func Times(keyName string, valueTimes []time.Time) Field {
	return Field{
		valueType:  TypeTime,
		keyName:    keyName,
		valueTimes: valueTimes,
	}
}
func NewLogger() Logger {
	universalLogger := &UniversalLogger{
		async:  false,
		format: TypeText,
		scheme: getLoggerScheme(),
		writer: os.Stderr,
	}
	universalLogger.level.Store(int32(getLoggerLevel()))
	return universalLogger
}
func NewLoggerAsync() Logger {
	universalLogger := &UniversalLogger{
		async:  true,
		format: TypeText,
		scheme: getLoggerScheme(),
		writer: newAsyncWriter(os.Stderr, 10000),
	}
	universalLogger.level.Store(int32(getLoggerLevel()))
	return universalLogger
}
func NewLoggerError(logger Logger) *log.Logger {
	standardLogger := &StandardLogger{
		flags:  log.LstdFlags | log.Lmicroseconds,
		logger: logger,
		scheme: getLoggerScheme(),
	}
	standardLogger.level.Store(int32(LevelError))
	return log.New(standardLogger, "", 0)
}
func NewWithWriter(level TypeLevel, logger Logger) io.Writer {
	standardLogger := &StandardLogger{
		flags:  log.LstdFlags | log.Lmicroseconds,
		logger: logger,
		scheme: getLoggerScheme(),
	}
	standardLogger.level.Store(int32(level))
	return standardLogger
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
func (asyncWriter *asyncWriter) Close() error {
	close(asyncWriter.ch)
	asyncWriter.wg.Wait()
	return nil
}
func (asyncWriter *asyncWriter) Write(p []byte) (n int, err error) {
	buf := make([]byte, len(p))
	copy(buf, p)
	select {
	case asyncWriter.ch <- buf:
		return len(p), nil
	default:
		return asyncWriter.writer.Write(p)
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
	caller      string
	message     string
	prefixDebug string
	prefixError string
	prefixFatal string
	prefixInfo  string
	prefixWarn  string
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
		caller:      colorDarkBlue,
		message:     colorDarkWhite,
		prefixDebug: colorDarkCyan + "[DEBUG]",
		prefixError: colorDarkRed + "[ERROR]",
		prefixFatal: colorDarkPurple + "[FATAL]",
		prefixInfo:  colorDarkGreen + "[INFO]",
		prefixWarn:  colorDarkYellow + "[WARN]",
		reset:       colorReset,
	}
	lightScheme = colorScheme{
		caller:      colorLightBlue,
		message:     colorLightBlack,
		prefixDebug: colorLightCyan + "[DEBUG]",
		prefixError: colorLightRed + "[ERROR]",
		prefixFatal: colorLightPurple + "[FATAL]",
		prefixInfo:  colorLightGreen + "[INFO]",
		prefixWarn:  colorLightYellow + "[WARN]",
		reset:       colorReset,
	}
	osExit   = os.Exit
	timePool = sync.Pool{
		New: func() any {
			return make([]byte, 0, 26)
		},
	}
)

// Приватная структура
type asyncWriter struct {
	ch     chan []byte
	limit  int
	wg     sync.WaitGroup
	writer io.Writer
}

// Приватные конструкторы
func newAsyncWriter(writer io.Writer, bufferSize int) *asyncWriter {
	asyncWriter := &asyncWriter{
		ch:     make(chan []byte, bufferSize),
		limit:  bufferSize,
		writer: writer,
	}
	asyncWriter.wg.Add(1)
	go asyncWriter.run()
	return asyncWriter
}

// Приватные функции
func escapeJSON(buf *bytes.Buffer, s string) {
	start := 0
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch >= 0x20 && ch != '"' && ch != '\\' {
			continue
		}
		if start < i {
			buf.WriteString(s[start:i])
		}
		switch ch {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if ch < 0x20 {
				buf.WriteString(fmt.Sprintf(`\u%04x`, ch))
			} else {
				buf.WriteByte(ch)
			}
		}
		start = i + 1
	}
	if start < len(s) {
		buf.WriteString(s[start:])
	}
}
func formatDataJson(dataBuf *bytes.Buffer, message string, fields []Field) {
	dataBuf.WriteString(`"message":"`)
	escapeJSON(dataBuf, message)
	dataBuf.WriteByte('"')
	if len(fields) != 0 {
		dataBuf.WriteString(`,"fields":{`)
		for i, field := range fields {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteByte('"')
			escapeJSON(dataBuf, field.keyName)
			dataBuf.WriteString(`":`)
			formatFieldValue(dataBuf, field)
		}
		dataBuf.WriteByte('}')
	}
}
func formatDataText(dataBuf *bytes.Buffer, message string, fields []Field, scheme colorScheme) {
	dataBuf.WriteString(scheme.message)
	dataBuf.WriteString(message)
	if len(fields) != 0 {
		dataBuf.WriteByte(':')
		for _, field := range fields {
			dataBuf.WriteByte(' ')
			dataBuf.WriteString(field.keyName)
			dataBuf.WriteByte('=')
			formatFieldValue(dataBuf, field)
		}
	}
	dataBuf.WriteString(scheme.reset)
	dataBuf.WriteByte('\n')
}
func formatFieldValue(dataBuf *bytes.Buffer, field Field) {
	switch field.valueType {
	case TypeBool:
		dataBuf.WriteString(strconv.FormatBool(field.valueBool))
	case TypeBools:
		dataBuf.WriteByte('[')
		for i, value := range field.valueBools {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(strconv.FormatBool(value))
		}
		dataBuf.WriteByte(']')
	case TypeDuration:
		dataBuf.WriteString(field.valueDuration.String())
	case TypeDurations:
		dataBuf.WriteByte('[')
		for i, value := range field.valueDurations {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(value.String())
		}
		dataBuf.WriteByte(']')
	case TypeFloat64:
		dataBuf.WriteString(strconv.FormatFloat(field.valueFloat64, 'f', -1, 64))
	case TypeFloats64:
		dataBuf.WriteByte('[')
		for i, value := range field.valueFloats64 {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
		}
		dataBuf.WriteByte(']')
	case TypeInt:
		dataBuf.WriteString(strconv.Itoa(field.valueInt))
	case TypeInts:
		dataBuf.WriteByte('[')
		for i, value := range field.valueInts {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(strconv.Itoa(value))
		}
		dataBuf.WriteByte(']')
	case TypeInt64:
		dataBuf.WriteString(strconv.FormatInt(field.valueInt64, 10))
	case TypeInts64:
		dataBuf.WriteByte('[')
		for i, value := range field.valueInts64 {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(strconv.FormatInt(value, 10))
		}
		dataBuf.WriteByte(']')
	case TypeString:
		dataBuf.WriteByte('"')
		dataBuf.WriteString(field.valueString)
		dataBuf.WriteByte('"')
	case TypeStrings:
		dataBuf.WriteByte('[')
		for i, value := range field.valueStrings {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteByte('"')
			dataBuf.WriteString(value)
			dataBuf.WriteByte('"')
		}
		dataBuf.WriteByte(']')
	case TypeTime:
		dataBuf.Write(field.valueTime.AppendFormat(nil, time.RFC3339Nano))
	case TypeTimes:
		dataBuf.WriteByte('[')
		for i, value := range field.valueTimes {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.Write(value.AppendFormat(nil, time.RFC3339Nano))
		}
		dataBuf.WriteByte(']')
	}
}
func formatPrefixJson(dataBuf *bytes.Buffer, level TypeLevel, caller string) {
	dataBuf.WriteString(`"level":"`)
	switch level {
	case LevelDebug:
		dataBuf.WriteString("debug")
	case LevelInfo:
		dataBuf.WriteString("info")
	case LevelWarn:
		dataBuf.WriteString("warn")
	case LevelError:
		dataBuf.WriteString("error")
	case LevelFatal:
		dataBuf.WriteString("fatal")
	}
	dataBuf.WriteByte('"')
	if caller != "" {
		dataBuf.WriteString(`,"caller":"`)
		escapeJSON(dataBuf, caller)
		dataBuf.WriteByte('"')
	}
}
func formatPrefixText(dataBuf *bytes.Buffer, level TypeLevel, caller string, scheme colorScheme) {
	switch level {
	case LevelDebug:
		dataBuf.WriteString(scheme.prefixDebug)
	case LevelError:
		dataBuf.WriteString(scheme.prefixError)
	case LevelFatal:
		dataBuf.WriteString(scheme.prefixFatal)
	case LevelInfo:
		dataBuf.WriteString(scheme.prefixInfo)
	case LevelWarn:
		dataBuf.WriteString(scheme.prefixWarn)
	}
	if caller != "" {
		dataBuf.WriteByte(' ')
		dataBuf.WriteString(scheme.caller)
		dataBuf.WriteString(caller)
	}
}
func formatTimeJson(dataBuf *bytes.Buffer, timestamp time.Time) {
	dataBuf.WriteString(`"time":"`)
	timeBuf := timePool.Get().([]byte)
	timeBuf = timestamp.AppendFormat(timeBuf[:0], time.RFC3339Nano)
	dataBuf.Write(timeBuf)
	timePool.Put(timeBuf)
	dataBuf.WriteByte('"')
}
func formatTimeText(dataBuf *bytes.Buffer, timestamp time.Time) {
	timeBuf := timePool.Get().([]byte)
	timeBuf = timestamp.AppendFormat(timeBuf[:0], time.RFC3339Nano)
	dataBuf.Write(timeBuf)
	timePool.Put(timeBuf)
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
func getLoggerLevel() TypeLevel {
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
func (asyncWriter *asyncWriter) run() {
	defer asyncWriter.wg.Done()
	for buf := range asyncWriter.ch {
		asyncWriter.writer.Write(buf)
	}
}
func (universalLogger *UniversalLogger) getCaller(level TypeLevel) string {
	if level != LevelDebug {
		return ""
	}
	pc, file, line, _ := runtime.Caller(2)
	if val, ok := universalLogger.cache.Load(pc); ok {
		return val.(string)
	}
	caller := getLoggerCaller(file) + ":" + strconv.Itoa(line)
	universalLogger.cache.Store(pc, caller)
	return caller
}
func (universalLogger *UniversalLogger) getLevel() TypeLevel {
	return TypeLevel(universalLogger.level.Load())
}
func (universalLogger *UniversalLogger) getScheme() colorScheme {
	universalLogger.mutex.RLock()
	defer universalLogger.mutex.RUnlock()
	return universalLogger.scheme
}
func (universalLogger *UniversalLogger) writeJson(level TypeLevel, context context.Context, message string, fields []Field) {
	if universalLogger.getLevel() > level {
		return
	}
	if context != nil {
		//
	}
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	caller := universalLogger.getCaller(level)
	time := time.Now()
	dataBuf.WriteByte('{')
	formatTimeJson(dataBuf, time)
	dataBuf.WriteByte(',')
	formatPrefixJson(dataBuf, level, caller)
	dataBuf.WriteByte(',')
	formatDataJson(dataBuf, message, fields)
	dataBuf.WriteByte('}')
	universalLogger.mutex.RLock()
	writer := universalLogger.writer
	universalLogger.mutex.RUnlock()
	writer.Write(dataBuf.Bytes())
}
func (universalLogger *UniversalLogger) writeText(level TypeLevel, context context.Context, message string, fields []Field) {
	if universalLogger.getLevel() > level {
		return
	}
	if context != nil {
		//
	}
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	caller := universalLogger.getCaller(level)
	scheme := universalLogger.getScheme()
	time := time.Now()
	formatTimeText(dataBuf, time)
	dataBuf.WriteByte(' ')
	formatPrefixText(dataBuf, level, caller, scheme)
	dataBuf.WriteByte(' ')
	formatDataText(dataBuf, message, fields, scheme)
	universalLogger.mutex.RLock()
	writer := universalLogger.writer
	universalLogger.mutex.RUnlock()
	writer.Write(dataBuf.Bytes())
}
