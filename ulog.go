// Copyright [2026] [Mikhail Dadaev]
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package ulog

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

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
	defaultFormat     = FormatJson
	defaultLevel      = LevelInfo
	defaultMode       = ModeSync
	defaultType       = -1
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
	caller     string
	theme      colorTheme
	typeData   TypeData
	typeFormat TypeFormat
	typeLevel  TypeLevel
}
type telemetryOptions func(*universalTelemetry)

// Приватные переменные
var (
	dataPool = sync.Pool{
		New: func() any {
			return new(bytes.Buffer)
		},
	}
	osExit          = os.Exit
	timeCacheMu     sync.Mutex
	timeCacheSec    int64
	timeCachePrefix = make([]byte, 0, 32)
	timeCacheTZ     string
	timeOnce        sync.Once
	themeDark       = colorTheme{
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
func getData(typeData TypeData) string {
	switch typeData {
	case DataLog:
		return "LOG"
	case DataMetric:
		return "METRIC"
	case DataTrace:
		return "TRACE"
	default:
		return "UNKNOWN"
	}
}
func getLevel(typeLevel TypeLevel) string {
	switch typeLevel {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	case LevelFatal:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}
func getTime(dataBuf *bytes.Buffer, timestamp time.Time) {
	unixSec := timestamp.Unix()
	unixNano := timestamp.UnixNano()
	if atomic.LoadInt64(&timeCacheSec) == unixSec {
		dataBuf.Write(timeCachePrefix)
	} else {
		timeCacheMu.Lock()
		if timeCacheSec != unixSec {
			timeBuf := timePool.Get().([]byte)
			timeBuf = timestamp.AppendFormat(timeBuf[:0], "2006-01-02T15:04:05")
			timeCachePrefix = timeCachePrefix[:0]
			timeCachePrefix = append(timeCachePrefix, timeBuf...)
			timePool.Put(timeBuf)
			timeOnce.Do(func() {
				tzBuf := timePool.Get().([]byte)
				tzBuf = timestamp.AppendFormat(tzBuf[:0], "-07:00")
				timeCacheTZ = string(tzBuf)
				timePool.Put(tzBuf)
			})
			atomic.StoreInt64(&timeCacheSec, unixSec)
		}
		timeCacheMu.Unlock()
		dataBuf.Write(timeCachePrefix)
	}
	millis := (unixNano / 1_000_000) % 1000
	micros := (unixNano / 1_000) % 1000
	dataBuf.WriteByte('.')
	dataBuf.WriteByte(byte('0' + (millis/100)%10))
	dataBuf.WriteByte(byte('0' + (millis/10)%10))
	dataBuf.WriteByte(byte('0' + millis%10))
	dataBuf.WriteByte(byte('0' + (micros/100)%10))
	dataBuf.WriteByte(byte('0' + (micros/10)%10))
	dataBuf.WriteByte(byte('0' + micros%10))
	dataBuf.WriteString(timeCacheTZ)
}
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
func formatJson(dataBuf *bytes.Buffer, attributes writeAttributes, fields []Field) {
	time := time.Now()
	dataBuf.WriteByte('{')
	formatJsonTime(dataBuf, time)
	dataBuf.WriteByte(',')
	formatJsonPrefix(dataBuf, attributes.typeLevel, attributes.caller)
	dataBuf.WriteByte(',')
	formatJsonData(dataBuf, attributes.typeData, fields)
	dataBuf.WriteByte('}')
	dataBuf.WriteByte('\n')
}
func formatJsonData(dataBuf *bytes.Buffer, typeData TypeData, fields []Field) {
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
func formatJsonPrefix(dataBuf *bytes.Buffer, level TypeLevel, caller string) {
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
func formatJsonTime(dataBuf *bytes.Buffer, timestamp time.Time) {
	dataBuf.WriteString(`"timestamp":"`)
	getTime(dataBuf, timestamp)
	dataBuf.WriteByte('"')
}
func formatText(dataBuf *bytes.Buffer, attributes writeAttributes, fields []Field) {
	time := time.Now()
	formatTextTime(dataBuf, time)
	dataBuf.WriteByte(' ')
	formatTextPrefix(dataBuf, attributes.typeLevel, attributes.caller, attributes.theme)
	dataBuf.WriteByte(' ')
	formatTextData(dataBuf, attributes.typeData, fields, attributes.theme)
	dataBuf.WriteByte('\n')
}
func formatTextData(dataBuf *bytes.Buffer, typeData TypeData, fields []Field, theme colorTheme) {
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
func formatTextPrefix(dataBuf *bytes.Buffer, level TypeLevel, caller string, theme colorTheme) {
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
func formatTextTime(dataBuf *bytes.Buffer, timestamp time.Time) {
	getTime(dataBuf, timestamp)
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
	dataBuf.WriteByte('"')
	dataBuf.WriteString(v.String())
	dataBuf.WriteByte('"')
}
func formatValueDurations(dataBuf *bytes.Buffer, v []time.Duration) {
	dataBuf.WriteByte('[')
	for i, d := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.WriteByte('"')
		dataBuf.WriteString(d.String())
		dataBuf.WriteByte('"')
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
	dataBuf.WriteByte('"')
	dataBuf.Write(v.AppendFormat(nil, "2006-01-02T15:04:05.000000-07:00"))
	dataBuf.WriteByte('"')
}
func formatValueTimes(dataBuf *bytes.Buffer, v []time.Time) {
	dataBuf.WriteByte('[')
	for i, t := range v {
		if i > 0 {
			dataBuf.WriteByte(',')
		}
		dataBuf.WriteByte('"')
		dataBuf.Write(t.AppendFormat(nil, "2006-01-02T15:04:05.000000-07:00"))
		dataBuf.WriteByte('"')
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
			fmt.Fprintf(DefaultWriterErr, "ulog: async write failed: %v\n", err)
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
func (universalTelemetry *universalTelemetry) route(context context.Context, attributes writeAttributes, fields []Field) {
	if universalTelemetry.getLevel() > attributes.typeLevel {
		return
	}
	if universalTelemetry.extractor != nil && context != nil {
		fields = append(fields, universalTelemetry.extractor(context)...)
	}
	universalTelemetry.mutex.RLock()
	writer := universalTelemetry.writer
	universalTelemetry.mutex.RUnlock()
	if sinks, ok := writer.(SinkWriter); ok {
		_, err := sinks.WriteWithAttributes(attributes, fields)
		if err != nil {
			fmt.Fprintf(DefaultWriterErr, "ulog: failed to write: %v\n", err)
		}
		return
	}
	dataBuf := dataPool.Get().(*bytes.Buffer)
	dataBuf.Reset()
	defer dataPool.Put(dataBuf)
	switch attributes.typeFormat {
	case FormatJson:
		formatJson(dataBuf, attributes, fields)
	case FormatText:
		formatText(dataBuf, attributes, fields)
	default:
		fmt.Fprintf(DefaultWriterErr, "ulog: unsupported format: %v\n", attributes.typeFormat)
	}
	if _, err := writer.Write(dataBuf.Bytes()); err != nil {
		fmt.Fprintf(DefaultWriterErr, "ulog: failed to write: %v\n", err)
	}
}
