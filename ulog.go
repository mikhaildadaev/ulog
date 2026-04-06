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
	BoolType FieldType = iota
	DurationType
	FloatType
	IntType
	StringType
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
	valueType      FieldType
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
func New() Logger {
	asyncWriter := NewAsyncWriter(os.Stderr, 10000)
	return &LoggerStandard{
		asyncWriter: asyncWriter,
		caller:      false,
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
		case DurationType:
			dataBuf.WriteString(field.valueDuration.String())
		case FloatType:
			dataBuf.WriteString(strconv.FormatFloat(field.valueFloat, 'f', -1, 64))
		case IntType:
			dataBuf.WriteString(strconv.FormatInt(field.valueInt, 10))
		case StringType:
			dataBuf.WriteString(field.valueString)
		case TimeType:
			dataBuf.Write(field.valueTime.AppendFormat(nil, "2006-01-02T15:04:05.000Z07:00"))
		}
	}
	dataBuf.WriteString(scheme.reset)
	dataBuf.WriteByte('\n')
}
func formatPrefix(dataBuf *bytes.Buffer, scheme colorScheme, level int, caller string) {
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
		return ""
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
	formatPrefix(dataBuf, scheme, level, caller)
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
	formatPrefix(dataBuf, scheme, level, caller)
	formatDataf(dataBuf, scheme, message, fields)
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
