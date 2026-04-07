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
	"sync/atomic"
	"time"
)

// Публичные типы
type TypeLevel int
type TypeField int
type TypeFormat int

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
	BoolType TypeField = iota
	BoolsType
	DurationType
	DurationsType
	FloatType
	FloatsType
	IntType
	IntsType
	StringType
	StringsType
	TimeType
	TimesType
)
const (
	JsonType TypeFormat = iota
	TextType
)

// Публичные интерфейсы
type Logger interface {
	Debug(message string, fields ...Field)
	Error(message string, fields ...Field)
	Fatal(message string, fields ...Field)
	Info(message string, fields ...Field)
	Warn(message string, fields ...Field)
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
	valueInt       int64
	valueInts      []int64
	valueFloat     float64
	valueFloats    []float64
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

// Публичные конструкторы
func Bool(keyName string, valueBool bool) Field {
	return Field{
		valueType: BoolType,
		keyName:   keyName,
		valueBool: valueBool,
	}
}
func Bools(keyName string, valueBools []bool) Field {
	return Field{
		valueType:  BoolType,
		keyName:    keyName,
		valueBools: valueBools,
	}
}
func Duration(keyName string, valueDuration time.Duration) Field {
	return Field{
		valueType:     DurationType,
		keyName:       keyName,
		valueDuration: valueDuration,
	}
}
func Durations(keyName string, valueDurations []time.Duration) Field {
	return Field{
		valueType:      DurationType,
		keyName:        keyName,
		valueDurations: valueDurations,
	}
}
func Err(err error) Field {
	if err == nil {
		return Field{
			valueType:   StringType,
			keyName:     "error",
			valueString: "nil",
		}
	}
	return Field{
		valueType:   StringType,
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
		valueType:    StringType,
		keyName:      "errors",
		valueStrings: values,
	}
}
func Float(keyName string, valueFloat float64) Field {
	return Field{
		valueType:  FloatType,
		keyName:    keyName,
		valueFloat: valueFloat,
	}
}
func Floats(keyName string, valueFloats []float64) Field {
	return Field{
		valueType:   FloatType,
		keyName:     keyName,
		valueFloats: valueFloats,
	}
}
func Int(keyName string, valueInt int64) Field {
	return Field{
		valueType: IntType,
		keyName:   keyName,
		valueInt:  valueInt,
	}
}
func Ints(keyName string, valueInts []int64) Field {
	return Field{
		valueType: IntType,
		keyName:   keyName,
		valueInts: valueInts,
	}
}
func String(keyName string, valueString string) Field {
	return Field{
		valueType:   StringType,
		keyName:     keyName,
		valueString: valueString,
	}
}
func Strings(keyName string, valueStrings []string) Field {
	return Field{
		valueType:    StringType,
		keyName:      keyName,
		valueStrings: valueStrings,
	}
}
func Time(keyName string, valueTime time.Time) Field {
	return Field{
		valueType: TimeType,
		keyName:   keyName,
		valueTime: valueTime,
	}
}
func Times(keyName string, valueTimes []time.Time) Field {
	return Field{
		valueType:  TimeType,
		keyName:    keyName,
		valueTimes: valueTimes,
	}
}
func NewLogger() Logger {
	universalLogger := &UniversalLogger{
		async:  false,
		format: TextType,
		scheme: getLoggerScheme(),
		writer: os.Stderr,
	}
	universalLogger.level.Store(int32(getLoggerLevel()))
	return universalLogger
}
func NewLoggerAsync() Logger {
	universalLogger := &UniversalLogger{
		async:  true,
		format: TextType,
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
func formatData(dataBuf *bytes.Buffer, scheme colorScheme, message string) {
	dataBuf.WriteString(scheme.message)
	dataBuf.WriteString(message)
	dataBuf.WriteString(scheme.reset)
	dataBuf.WriteByte('\n')
}
func formatDataf(dataBuf *bytes.Buffer, scheme colorScheme, message string, fields []Field) {
	dataBuf.WriteByte(' ')
	dataBuf.WriteString(scheme.message)
	dataBuf.WriteString(message)
	dataBuf.WriteByte(':')
	for _, field := range fields {
		dataBuf.WriteByte(' ')
		dataBuf.WriteString(field.keyName)
		dataBuf.WriteByte('=')
		switch field.valueType {
		case BoolType:
			dataBuf.WriteString(strconv.FormatBool(field.valueBool))
		case BoolsType:
			dataBuf.WriteByte('[')
			for i, value := range field.valueBools {
				if i > 0 {
					dataBuf.WriteByte(' ')
				}
				dataBuf.WriteString(strconv.FormatBool(value))
			}
			dataBuf.WriteByte(']')
		case DurationType:
			dataBuf.WriteString(field.valueDuration.String())
		case DurationsType:
			dataBuf.WriteByte('[')
			for i, value := range field.valueDurations {
				if i > 0 {
					dataBuf.WriteByte(' ')
				}
				dataBuf.WriteString(value.String())
			}
			dataBuf.WriteByte(']')
		case FloatType:
			dataBuf.WriteString(strconv.FormatFloat(field.valueFloat, 'f', -1, 64))
		case FloatsType:
			dataBuf.WriteByte('[')
			for i, value := range field.valueFloats {
				if i > 0 {
					dataBuf.WriteByte(' ')
				}
				dataBuf.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
			}
			dataBuf.WriteByte(']')
		case IntType:
			dataBuf.WriteString(strconv.FormatInt(field.valueInt, 10))
		case IntsType:
			dataBuf.WriteByte('[')
			for i, value := range field.valueInts {
				if i > 0 {
					dataBuf.WriteByte(' ')
				}
				dataBuf.WriteString(strconv.FormatInt(value, 10))
			}
			dataBuf.WriteByte(']')
		case StringType:
			dataBuf.WriteString(field.valueString)
		case StringsType:
			dataBuf.WriteByte('[')
			for i, value := range field.valueStrings {
				if i > 0 {
					dataBuf.WriteByte(' ')
				}
				dataBuf.WriteString(value)
			}
			dataBuf.WriteByte(']')
		case TimeType:
			dataBuf.Write(field.valueTime.AppendFormat(nil, "2006-01-02T15:04:05.000Z07:00"))
		case TimesType:
			dataBuf.WriteByte('[')
			for i, value := range field.valueTimes {
				if i > 0 {
					dataBuf.WriteByte(' ')
				}
				dataBuf.Write(value.AppendFormat(nil, "2006-01-02T15:04:05.000Z07:00"))
			}
			dataBuf.WriteByte(']')
		}
	}
	dataBuf.WriteString(scheme.reset)
	dataBuf.WriteByte('\n')
}
func formatPrefix(dataBuf *bytes.Buffer, scheme colorScheme, level TypeLevel, caller string) {
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
	if caller != "" {
		dataBuf.WriteString(scheme.caller)
		dataBuf.WriteString(caller)
		dataBuf.WriteByte(' ')
	}
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
func (universalLogger *UniversalLogger) writeJson(level TypeLevel, message string) {
	if universalLogger.getLevel() > level {
		return
	}
	// Дописать
}
func (universalLogger *UniversalLogger) writeText(level TypeLevel, message string) {
	if universalLogger.getLevel() > level {
		return
	}
	caller := universalLogger.getCaller(level)
	scheme := universalLogger.getScheme()
	time := time.Now()
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	formatTime(dataBuf, time)
	formatPrefix(dataBuf, scheme, level, caller)
	formatData(dataBuf, scheme, message)
	universalLogger.mutex.RLock()
	writer := universalLogger.writer
	universalLogger.mutex.RUnlock()
	writer.Write(dataBuf.Bytes())
}
func (universalLogger *UniversalLogger) writeJsonFields(level TypeLevel, message string, fields []Field) {
	if universalLogger.getLevel() > level {
		return
	}
	// Дописать
}
func (universalLogger *UniversalLogger) writeTextFields(level TypeLevel, message string, fields []Field) {
	if universalLogger.getLevel() > level {
		return
	}
	caller := universalLogger.getCaller(level)
	scheme := universalLogger.getScheme()
	time := time.Now()
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	formatTime(dataBuf, time)
	formatPrefix(dataBuf, scheme, level, caller)
	formatDataf(dataBuf, scheme, message, fields)
	universalLogger.mutex.RLock()
	writer := universalLogger.writer
	universalLogger.mutex.RUnlock()
	writer.Write(dataBuf.Bytes())
}
