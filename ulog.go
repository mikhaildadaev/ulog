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
type TypeField int
type TypeFormat int
type TypeLevel int
type TypeMode int
type TypeTheme int

// Публичные константы
const (
	Author  = "Mikhail Dadaev"
	Version = "1.26.5"
)
const (
	FieldBool TypeField = iota
	FieldBools
	FieldDuration
	FieldDurations
	FieldFloat64
	FieldFloats64
	FieldInt
	FieldInts
	FieldInt64
	FieldInts64
	FieldString
	FieldStrings
	FieldTime
	FieldTimes
)
const (
	FormatJson TypeFormat = iota
	FormatText
)
const (
	LevelDebug TypeLevel = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelFatal
)
const (
	ModeAsync TypeMode = iota
	ModeSync
)
const (
	ThemeDark TypeTheme = iota
	ThemeLight
)

// Публичные интерфейсы
type Logger interface {
	Close() error
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
	SetExtractor(extractor ContextExtractor)
	SetFormat(format TypeFormat)
	SetLevel(level TypeLevel)
	SetMode(mode TypeMode, writer io.Writer, bufferSize ...int)
	SetTheme(theme TypeTheme)
	Sync() error
}

// Публичные структуры
type Field struct {
	nameKey        string
	typeValue      TypeField
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
}

type ContextExtractor func(context context.Context) []Field
type OptionLogger func(*universalLogger)

// Публичные конструкторы
func Bool(nameKey string, valueBool bool) Field {
	return Field{
		nameKey:   nameKey,
		typeValue: FieldBool,
		valueBool: valueBool,
	}
}
func Bools(nameKey string, valueBools []bool) Field {
	return Field{
		nameKey:    nameKey,
		typeValue:  FieldBool,
		valueBools: valueBools,
	}
}
func Duration(nameKey string, valueDuration time.Duration) Field {
	return Field{
		nameKey:       nameKey,
		typeValue:     FieldDuration,
		valueDuration: valueDuration,
	}
}
func Durations(nameKey string, valueDurations []time.Duration) Field {
	return Field{
		nameKey:        nameKey,
		typeValue:      FieldDuration,
		valueDurations: valueDurations,
	}
}
func Err(err error) Field {
	if err == nil {
		return Field{
			nameKey:     "error",
			typeValue:   FieldString,
			valueString: "nil",
		}
	}
	return Field{
		nameKey:     "error",
		typeValue:   FieldString,
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
		nameKey:      "errors",
		typeValue:    FieldString,
		valueStrings: values,
	}
}
func Float64(nameKey string, valueFloat64 float64) Field {
	return Field{
		nameKey:      nameKey,
		typeValue:    FieldFloat64,
		valueFloat64: valueFloat64,
	}
}
func Floats64(nameKey string, valueFloats64 []float64) Field {
	return Field{
		nameKey:       nameKey,
		typeValue:     FieldFloat64,
		valueFloats64: valueFloats64,
	}
}
func Int(nameKey string, valueInt int) Field {
	return Field{
		nameKey:   nameKey,
		typeValue: FieldInt,
		valueInt:  valueInt,
	}
}
func Ints(nameKey string, valueInts []int) Field {
	return Field{
		nameKey:   nameKey,
		typeValue: FieldInt,
		valueInts: valueInts,
	}
}
func Int64(nameKey string, valueInt64 int64) Field {
	return Field{
		nameKey:    nameKey,
		typeValue:  FieldInt,
		valueInt64: valueInt64,
	}
}
func Ints64(nameKey string, valueInts64 []int64) Field {
	return Field{
		nameKey:     nameKey,
		typeValue:   FieldInt,
		valueInts64: valueInts64,
	}
}
func String(nameKey string, valueString string) Field {
	return Field{
		nameKey:     nameKey,
		typeValue:   FieldString,
		valueString: valueString,
	}
}
func Strings(nameKey string, valueStrings []string) Field {
	return Field{
		nameKey:      nameKey,
		typeValue:    FieldString,
		valueStrings: valueStrings,
	}
}
func Time(nameKey string, valueTime time.Time) Field {
	return Field{
		nameKey:   nameKey,
		typeValue: FieldTime,
		valueTime: valueTime,
	}
}
func Times(nameKey string, valueTimes []time.Time) Field {
	return Field{
		nameKey:    nameKey,
		typeValue:  FieldTime,
		valueTimes: valueTimes,
	}
}
func NewLogger(options ...OptionLogger) Logger {
	universalLogger := &universalLogger{
		mode:   defaultMode,
		theme:  getLoggerTheme(),
		writer: os.Stderr,
	}
	universalLogger.format.Store(int32(defaultFormat))
	universalLogger.level.Store(int32(getLoggerLevel()))
	for _, option := range options {
		option(universalLogger)
	}
	return universalLogger
}
func NewLoggerLog(level TypeLevel, logger Logger) *log.Logger {
	standardLogger := &standardLogger{
		logger: logger,
	}
	standardLogger.level.Store(int32(level))
	return log.New(standardLogger, "", 0)
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
func (asyncWriter *asyncWriter) Sync() error {
	asyncWriter.wg.Wait()
	return nil
}
func (asyncWriter *asyncWriter) Write(p []byte) (n int, err error) {
	buf := make([]byte, len(p))
	copy(buf, p)
	asyncWriter.wg.Add(1)
	select {
	case asyncWriter.ch <- buf:
		return len(p), nil
	default:
		return asyncWriter.writer.Write(p)
	}
}

// Приватные константы
const (
	defaultBufferSize = 10000
	defaultFormat     = FormatText
	defaultLevel      = LevelInfo
	defaultMode       = ModeSync
)
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
type colorTheme struct {
	caller      string
	message     string
	prefixDebug string
	prefixError string
	prefixFatal string
	prefixInfo  string
	prefixWarn  string
	reset       string
}
type standardLogger struct {
	level  atomic.Int32
	logger Logger
	mutex  sync.Mutex
}
type universalLogger struct {
	cache     sync.Map
	extractor ContextExtractor
	format    atomic.Int32
	level     atomic.Int32
	mode      TypeMode
	mutex     sync.RWMutex
	theme     colorTheme
	writer    io.Writer
}

// Приватные переменные
var (
	dataPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
	darkTheme = colorTheme{
		caller:      colorDarkBlue,
		message:     colorDarkWhite,
		prefixDebug: colorDarkCyan + "[DEBUG]",
		prefixError: colorDarkRed + "[ERROR]",
		prefixFatal: colorDarkPurple + "[FATAL]",
		prefixInfo:  colorDarkGreen + "[INFO]",
		prefixWarn:  colorDarkYellow + "[WARN]",
		reset:       colorReset,
	}
	lightTheme = colorTheme{
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
			escapeJSON(dataBuf, field.nameKey)
			dataBuf.WriteString(`":`)
			formatFieldValue(dataBuf, field)
		}
		dataBuf.WriteByte('}')
	}
}
func formatDataText(dataBuf *bytes.Buffer, message string, fields []Field, theme colorTheme) {
	dataBuf.WriteString(theme.message)
	dataBuf.WriteString(message)
	if len(fields) != 0 {
		dataBuf.WriteByte(':')
		for _, field := range fields {
			dataBuf.WriteByte(' ')
			dataBuf.WriteString(field.nameKey)
			dataBuf.WriteByte('=')
			formatFieldValue(dataBuf, field)
		}
	}
	dataBuf.WriteString(theme.reset)
	dataBuf.WriteByte('\n')
}
func formatFieldValue(dataBuf *bytes.Buffer, field Field) {
	switch field.typeValue {
	case FieldBool:
		dataBuf.WriteString(strconv.FormatBool(field.valueBool))
	case FieldBools:
		dataBuf.WriteByte('[')
		for i, value := range field.valueBools {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(strconv.FormatBool(value))
		}
		dataBuf.WriteByte(']')
	case FieldDuration:
		dataBuf.WriteString(field.valueDuration.String())
	case FieldDurations:
		dataBuf.WriteByte('[')
		for i, value := range field.valueDurations {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(value.String())
		}
		dataBuf.WriteByte(']')
	case FieldFloat64:
		dataBuf.WriteString(strconv.FormatFloat(field.valueFloat64, 'f', -1, 64))
	case FieldFloats64:
		dataBuf.WriteByte('[')
		for i, value := range field.valueFloats64 {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
		}
		dataBuf.WriteByte(']')
	case FieldInt:
		dataBuf.WriteString(strconv.Itoa(field.valueInt))
	case FieldInts:
		dataBuf.WriteByte('[')
		for i, value := range field.valueInts {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(strconv.Itoa(value))
		}
		dataBuf.WriteByte(']')
	case FieldInt64:
		dataBuf.WriteString(strconv.FormatInt(field.valueInt64, 10))
	case FieldInts64:
		dataBuf.WriteByte('[')
		for i, value := range field.valueInts64 {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteString(strconv.FormatInt(value, 10))
		}
		dataBuf.WriteByte(']')
	case FieldString:
		dataBuf.WriteByte('"')
		dataBuf.WriteString(field.valueString)
		dataBuf.WriteByte('"')
	case FieldStrings:
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
	case FieldTime:
		dataBuf.Write(field.valueTime.AppendFormat(nil, "2006-01-02T15:04:05.000000-07:00"))
	case FieldTimes:
		dataBuf.WriteByte('[')
		for i, value := range field.valueTimes {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.Write(value.AppendFormat(nil, "2006-01-02T15:04:05.000000-07:00"))
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
func formatPrefixText(dataBuf *bytes.Buffer, level TypeLevel, caller string, theme colorTheme) {
	switch level {
	case LevelDebug:
		dataBuf.WriteString(theme.prefixDebug)
	case LevelInfo:
		dataBuf.WriteString(theme.prefixInfo)
	case LevelWarn:
		dataBuf.WriteString(theme.prefixWarn)
	case LevelError:
		dataBuf.WriteString(theme.prefixError)
	case LevelFatal:
		dataBuf.WriteString(theme.prefixFatal)
	}
	if caller != "" {
		dataBuf.WriteByte(' ')
		dataBuf.WriteString(theme.caller)
		dataBuf.WriteString(caller)
	}
}
func formatTimeJson(dataBuf *bytes.Buffer, timestamp time.Time) {
	dataBuf.WriteString(`"time":"`)
	timeBuf := timePool.Get().([]byte)
	timeBuf = timestamp.AppendFormat(timeBuf[:0], "2006-01-02T15:04:05.000000-07:00")
	dataBuf.Write(timeBuf)
	timePool.Put(timeBuf)
	dataBuf.WriteByte('"')
}
func formatTimeText(dataBuf *bytes.Buffer, timestamp time.Time) {
	timeBuf := timePool.Get().([]byte)
	timeBuf = timestamp.AppendFormat(timeBuf[:0], "2006-01-02T15:04:05.000000-07:00")
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
	return defaultLevel
}
func getLoggerTheme() colorTheme {
	switch strings.ToLower(os.Getenv("TERM_THEME")) {
	case "dark":
		return darkTheme
	case "light":
		return lightTheme
	}
	if os.Getenv("COLORFGBG") != "" {
		parts := strings.Split(os.Getenv("COLORFGBG"), ";")
		if len(parts) >= 2 {
			bg, _ := strconv.Atoi(parts[1])
			if bg < 8 {
				return darkTheme
			}
			return lightTheme
		}
	}
	return darkTheme
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
	for buf := range asyncWriter.ch {
		if _, err := asyncWriter.writer.Write(buf); err != nil {
			// Дописать место логирования ошибки
		}
		asyncWriter.wg.Done()
	}
}
func (universalLogger *universalLogger) getCaller(level TypeLevel) string {
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
func (universalLogger *universalLogger) getLevel() TypeLevel {
	return TypeLevel(universalLogger.level.Load())
}
func (universalLogger *universalLogger) getTheme() colorTheme {
	universalLogger.mutex.RLock()
	defer universalLogger.mutex.RUnlock()
	return universalLogger.theme
}
func (universalLogger *universalLogger) writeJson(level TypeLevel, context context.Context, message string, fields []Field) {
	if universalLogger.getLevel() > level {
		return
	}
	if universalLogger.extractor != nil && context != nil {
		fields = append(fields, universalLogger.extractor(context)...)
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
func (universalLogger *universalLogger) writeText(level TypeLevel, context context.Context, message string, fields []Field) {
	if universalLogger.getLevel() > level {
		return
	}
	if universalLogger.extractor != nil && context != nil {
		fields = append(fields, universalLogger.extractor(context)...)
	}
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	caller := universalLogger.getCaller(level)
	theme := universalLogger.getTheme()
	time := time.Now()
	formatTimeText(dataBuf, time)
	dataBuf.WriteByte(' ')
	formatPrefixText(dataBuf, level, caller, theme)
	dataBuf.WriteByte(' ')
	formatDataText(dataBuf, message, fields, theme)
	universalLogger.mutex.RLock()
	writer := universalLogger.writer
	universalLogger.mutex.RUnlock()
	writer.Write(dataBuf.Bytes())
}
