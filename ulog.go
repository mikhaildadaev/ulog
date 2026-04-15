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
type TypeData int
type TypeField int
type TypeFormat int
type TypeLevel int
type TypeMode int
type TypeTheme int

// Публичные константы
const (
	Author  = "Mikhail Dadaev"
	Version = "1.26.7"
)
const (
	DataLog TypeData = iota
	DataMetric
	DataTrace
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
type Telemetry interface {
	Close() error
	Debug(typeData TypeData, fields ...Field)
	DebugWithContext(ctx context.Context, typeData TypeData, fields ...Field)
	Error(typeData TypeData, fields ...Field)
	ErrorWithContext(ctx context.Context, typeData TypeData, fields ...Field)
	Fatal(typeData TypeData, fields ...Field)
	FatalWithContext(ctx context.Context, typeData TypeData, fields ...Field)
	Info(typeData TypeData, fields ...Field)
	InfoWithContext(ctx context.Context, typeData TypeData, fields ...Field)
	Warn(typeData TypeData, fields ...Field)
	WarnWithContext(ctx context.Context, typeData TypeData, fields ...Field)
	SetExtractor(keys ...string)
	SetFormat(format TypeFormat)
	SetLevel(level TypeLevel)
	SetMode(mode TypeMode, writer io.Writer, bufferSize ...int)
	SetTheme(theme TypeTheme)
	Sync() error
}
type SinkWriter interface {
	io.Writer
	WriteWithAttributes(attributes writeAttributes, p []byte) (n int, err error)
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

// Публичные конструкторы
func NewTelemetry(options ...optionTelemetry) Telemetry {
	universalTelemetry := &universalTelemetry{
		mode:   defaultMode,
		theme:  getDefaultTheme(),
		writer: defaultWriterOut,
	}
	universalTelemetry.format.Store(int32(defaultFormat))
	universalTelemetry.level.Store(int32(getDefaultLevel()))
	for _, option := range options {
		option(universalTelemetry)
	}
	return universalTelemetry
}
func NewTelemetryLog(level TypeLevel, telemetry Telemetry) *log.Logger {
	standardTelemetry := &standardTelemetry{
		telemetry: telemetry,
	}
	standardTelemetry.level.Store(int32(level))
	return log.New(standardTelemetry, "", 0)
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
	if closer, ok := asyncWriter.writer.(io.Closer); ok {
		return closer.Close()
	}
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
var (
	defaultBufferSize = 10000
	defaultFormat     = FormatText
	defaultLevel      = LevelInfo
	defaultMode       = ModeSync
	defaultType       = -1
	defaultWriterErr  = os.Stderr
	defaultWriterOut  = os.Stdout
)
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
type asyncWriter struct {
	ch     chan []byte
	limit  int
	wg     sync.WaitGroup
	writer io.Writer
}
type colorTheme struct {
	caller      string
	data        string
	prefixDebug string
	prefixError string
	prefixFatal string
	prefixInfo  string
	prefixWarn  string
	reset       string
}
type standardTelemetry struct {
	level     atomic.Int32
	mutex     sync.Mutex
	telemetry Telemetry
}
type universalTelemetry struct {
	cache     sync.Map
	extractor ContextExtractor
	format    atomic.Int32
	level     atomic.Int32
	mode      TypeMode
	mutex     sync.RWMutex
	theme     colorTheme
	writer    io.Writer
}
type writeAttributes struct {
	typeData  TypeData
	typeLevel TypeLevel
}
type optionTelemetry func(*universalTelemetry)

// Приватные переменные
var (
	dataPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
	osExit    = os.Exit
	themeDark = colorTheme{
		caller:      colorDarkBlue,
		data:        colorDarkWhite,
		prefixDebug: colorDarkCyan + "[DEBUG]",
		prefixError: colorDarkRed + "[ERROR]",
		prefixFatal: colorDarkPurple + "[FATAL]",
		prefixInfo:  colorDarkGreen + "[INFO]",
		prefixWarn:  colorDarkYellow + "[WARN]",
		reset:       colorReset,
	}
	themeLight = colorTheme{
		caller:      colorLightBlue,
		data:        colorLightBlack,
		prefixDebug: colorLightCyan + "[DEBUG]",
		prefixError: colorLightRed + "[ERROR]",
		prefixFatal: colorLightPurple + "[FATAL]",
		prefixInfo:  colorLightGreen + "[INFO]",
		prefixWarn:  colorLightYellow + "[WARN]",
		reset:       colorReset,
	}
	timePool = sync.Pool{
		New: func() any {
			return make([]byte, 0, 26)
		},
	}
)

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
func getTypeData(buf *bytes.Buffer, typeData TypeData) {
	switch typeData {
	case 0:
		buf.WriteString(`log`)
	case 1:
		buf.WriteString(`metric`)
	case 2:
		buf.WriteString(`trace`)
	}
}
func escapeJson(buf *bytes.Buffer, s string) {
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
func formatDataJson(dataBuf *bytes.Buffer, typeData TypeData, fields []Field) {
	dataBuf.WriteString(`"type":"`)
	getTypeData(dataBuf, typeData)
	dataBuf.WriteByte('"')
	if len(fields) != 0 {
		dataBuf.WriteString(`,`)
		for i, field := range fields {
			if i > 0 {
				dataBuf.WriteByte(',')
			}
			dataBuf.WriteByte('"')
			escapeJson(dataBuf, field.nameKey)
			dataBuf.WriteString(`":`)
			formatFieldValue(dataBuf, field)
		}
	}
}
func formatDataText(dataBuf *bytes.Buffer, typeData TypeData, fields []Field, theme colorTheme) {
	dataBuf.WriteString(theme.data)
	dataBuf.WriteString(`type="`)
	getTypeData(dataBuf, typeData)
	dataBuf.WriteByte('"')
	if len(fields) != 0 {
		for _, field := range fields {
			dataBuf.WriteByte(' ')
			dataBuf.WriteString(field.nameKey)
			dataBuf.WriteByte('=')
			formatFieldValue(dataBuf, field)
		}
	}
	dataBuf.WriteString(theme.reset)
}
func formatFieldValue(dataBuf *bytes.Buffer, field Field) {
	switch field.typeValue {
	case FieldString:
		formatValueString(dataBuf, field.valueString)
	case FieldStrings:
		formatValueStrings(dataBuf, field.valueStrings)
	case FieldInt:
		formatValueInt(dataBuf, field.valueInt)
	case FieldInts:
		formatValueInts(dataBuf, field.valueInts)
	case FieldInt64:
		formatValueInt64(dataBuf, field.valueInt64)
	case FieldInts64:
		formatValueInts64(dataBuf, field.valueInts64)
	case FieldFloat64:
		formatValueFloat64(dataBuf, field.valueFloat64)
	case FieldFloats64:
		formatValueFloats64(dataBuf, field.valueFloats64)
	case FieldBool:
		formatValueBool(dataBuf, field.valueBool)
	case FieldBools:
		formatValueBools(dataBuf, field.valueBools)
	case FieldTime:
		formatValueTime(dataBuf, field.valueTime)
	case FieldTimes:
		formatValueTimes(dataBuf, field.valueTimes)
	case FieldDuration:
		formatValueDuration(dataBuf, field.valueDuration)
	case FieldDurations:
		formatValueDurations(dataBuf, field.valueDurations)
	}
}
func formatPrefixJson(dataBuf *bytes.Buffer, level TypeLevel, caller string) {
	dataBuf.WriteString(`"level":"`)
	switch level {
	case LevelDebug:
		dataBuf.WriteString(`debug`)
	case LevelInfo:
		dataBuf.WriteString(`info`)
	case LevelWarn:
		dataBuf.WriteString(`warn`)
	case LevelError:
		dataBuf.WriteString(`error`)
	case LevelFatal:
		dataBuf.WriteString(`fatal`)
	}
	dataBuf.WriteByte('"')
	if caller != "" {
		dataBuf.WriteString(`,"caller":"`)
		escapeJson(dataBuf, caller)
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
func formatValueBool(dataBuf *bytes.Buffer, v bool) {
	dataBuf.WriteString(strconv.FormatBool(v))
}
func formatValueBools(dataBuf *bytes.Buffer, v []bool) {
	dataBuf.WriteByte('[')
	for i, b := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.WriteString(strconv.FormatBool(b))
	}
	dataBuf.WriteByte(']')
}
func formatValueDuration(dataBuf *bytes.Buffer, v time.Duration) {
	dataBuf.WriteString(v.String())
}
func formatValueDurations(dataBuf *bytes.Buffer, v []time.Duration) {
	dataBuf.WriteByte('[')
	for i, d := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.WriteString(d.String())
	}
	dataBuf.WriteByte(']')
}
func formatValueFloat64(dataBuf *bytes.Buffer, v float64) {
	dataBuf.WriteString(strconv.FormatFloat(v, 'f', -1, 64))
}
func formatValueFloats64(dataBuf *bytes.Buffer, v []float64) {
	dataBuf.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.WriteString(strconv.FormatFloat(f, 'f', -1, 64))
	}
	dataBuf.WriteByte(']')
}
func formatValueInt(dataBuf *bytes.Buffer, v int) {
	dataBuf.WriteString(strconv.Itoa(v))
}
func formatValueInts(dataBuf *bytes.Buffer, v []int) {
	dataBuf.WriteByte('[')
	for i, n := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.WriteString(strconv.Itoa(n))
	}
	dataBuf.WriteByte(']')
}
func formatValueInt64(dataBuf *bytes.Buffer, v int64) {
	dataBuf.WriteString(strconv.FormatInt(v, 10))
}
func formatValueInts64(dataBuf *bytes.Buffer, v []int64) {
	dataBuf.WriteByte('[')
	for i, n := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.WriteString(strconv.FormatInt(n, 10))
	}
	dataBuf.WriteByte(']')
}
func formatValueString(dataBuf *bytes.Buffer, v string) {
	dataBuf.WriteByte('"')
	dataBuf.WriteString(v)
	dataBuf.WriteByte('"')
}
func formatValueStrings(dataBuf *bytes.Buffer, v []string) {
	dataBuf.WriteByte('[')
	for i, s := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.WriteByte('"')
		dataBuf.WriteString(s)
		dataBuf.WriteByte('"')
	}
	dataBuf.WriteByte(']')
}
func formatValueTime(dataBuf *bytes.Buffer, v time.Time) {
	dataBuf.Write(v.AppendFormat(nil, "2006-01-02T15:04:05.000000-07:00"))
}
func formatValueTimes(dataBuf *bytes.Buffer, v []time.Time) {
	dataBuf.WriteByte('[')
	for i, t := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.Write(t.AppendFormat(nil, "2006-01-02T15:04:05.000000-07:00"))
	}
	dataBuf.WriteByte(']')
}
func getDefaultLevel() TypeLevel {
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
func getDefaultTheme() colorTheme {
	switch strings.ToLower(os.Getenv("TERM_THEME")) {
	case "dark":
		return themeDark
	case "light":
		return themeLight
	}
	if os.Getenv("COLORFGBG") != "" {
		parts := strings.Split(os.Getenv("COLORFGBG"), ";")
		if len(parts) >= 2 {
			bg, _ := strconv.Atoi(parts[1])
			if bg < 8 {
				return themeDark
			}
			return themeLight
		}
	}
	return themeDark
}

// Приватные методы
func (asyncWriter *asyncWriter) run() {
	for buf := range asyncWriter.ch {
		if _, err := asyncWriter.writer.Write(buf); err != nil {
			fmt.Fprintf(defaultWriterErr, "ulog: async write failed: %v\n", err)
		}
		asyncWriter.wg.Done()
	}
}
func (standardTelemetry *standardTelemetry) isIgnored(data []byte) bool {
	for _, err := range ignoredErrors {
		if bytes.Contains(data, err) {
			return true
		}
	}
	return false
}
func (universalTelemetry *universalTelemetry) getCaller(level TypeLevel) string {
	if level != LevelDebug {
		return ""
	}
	pc, file, line, _ := runtime.Caller(2)
	if val, ok := universalTelemetry.cache.Load(pc); ok {
		return val.(string)
	}
	for i := len(file) - 1; i >= 0; i-- {
		if file[i] == '/' || file[i] == '\\' {
			file = file[i+1:]
			break
		}
	}
	caller := file + ":" + strconv.Itoa(line)
	universalTelemetry.cache.Store(pc, caller)
	return caller
}
func (universalTelemetry *universalTelemetry) getLevel() TypeLevel {
	return TypeLevel(universalTelemetry.level.Load())
}
func (universalTelemetry *universalTelemetry) getTheme() colorTheme {
	universalTelemetry.mutex.RLock()
	defer universalTelemetry.mutex.RUnlock()
	return universalTelemetry.theme
}
func (universalTelemetry *universalTelemetry) writeJson(context context.Context, attributes writeAttributes, fields []Field) {
	if universalTelemetry.getLevel() > attributes.typeLevel {
		return
	}
	if universalTelemetry.extractor != nil && context != nil {
		fields = append(fields, universalTelemetry.extractor(context)...)
	}
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	caller := universalTelemetry.getCaller(attributes.typeLevel)
	time := time.Now()
	dataBuf.WriteByte('{')
	formatTimeJson(dataBuf, time)
	dataBuf.WriteByte(',')
	formatPrefixJson(dataBuf, attributes.typeLevel, caller)
	dataBuf.WriteByte(',')
	formatDataJson(dataBuf, attributes.typeData, fields)
	dataBuf.WriteByte('}')
	dataBuf.WriteByte('\n')
	universalTelemetry.mutex.RLock()
	writer := universalTelemetry.writer
	universalTelemetry.mutex.RUnlock()
	if sinks, ok := writer.(SinkWriter); ok {
		_, err := sinks.WriteWithAttributes(attributes, dataBuf.Bytes())
		if err != nil {
			fmt.Fprintf(defaultWriterErr, "ulog: failed to write: %v\n", err)
		}
		return
	}
	if _, err := writer.Write(dataBuf.Bytes()); err != nil {
		fmt.Fprintf(defaultWriterErr, "ulog: failed to write: %v\n", err)
	}
}
func (universalTelemetry *universalTelemetry) writeText(context context.Context, attributes writeAttributes, fields []Field) {
	if universalTelemetry.getLevel() > attributes.typeLevel {
		return
	}
	if universalTelemetry.extractor != nil && context != nil {
		fields = append(fields, universalTelemetry.extractor(context)...)
	}
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	caller := universalTelemetry.getCaller(attributes.typeLevel)
	theme := universalTelemetry.getTheme()
	time := time.Now()
	formatTimeText(dataBuf, time)
	dataBuf.WriteByte(' ')
	formatPrefixText(dataBuf, attributes.typeLevel, caller, theme)
	dataBuf.WriteByte(' ')
	formatDataText(dataBuf, attributes.typeData, fields, theme)
	dataBuf.WriteByte('\n')
	universalTelemetry.mutex.RLock()
	writer := universalTelemetry.writer
	universalTelemetry.mutex.RUnlock()
	if sinks, ok := writer.(SinkWriter); ok {
		_, err := sinks.WriteWithAttributes(attributes, dataBuf.Bytes())
		if err != nil {
			fmt.Fprintf(defaultWriterErr, "ulog: failed to write: %v\n", err)
		}
		return
	}
	if _, err := writer.Write(dataBuf.Bytes()); err != nil {
		fmt.Fprintf(defaultWriterErr, "ulog: failed to write: %v\n", err)
	}
}
