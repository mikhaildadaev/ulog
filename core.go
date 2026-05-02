package ulog

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
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
	Version = "1.26.11"
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

// Публичные переменные
var (
	DefaultWriterErr  = os.Stderr
	DefaultWriterNull = io.Discard
	DefaultWriterOut  = os.Stdout
)

// Публичные интерфейсы
type Telemetry interface {
	Close() error
	SetExtractor(keys ...string)
	SetFormat(format TypeFormat)
	SetLevel(level TypeLevel)
	SetMode(mode TypeMode, writer io.Writer, bufferSize ...int)
	SetTheme(theme TypeTheme)
	Sync() error
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
}
type SinkWriter interface {
	io.Writer
	WriteWithAttributes(attributes writeAttributes, fields []Field) (n int, err error)
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

// Публичные конструкторы
func NewTelemetry(options ...telemetryOptions) Telemetry {
	universalTelemetry := &universalTelemetry{
		mode:   defaultMode,
		theme:  getDefaultTheme(),
		writer: DefaultWriterOut,
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
		typeValue:  FieldBools,
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
		typeValue:      FieldDurations,
		valueDurations: valueDurations,
	}
}
func Error(valueError error) Field {
	if valueError == nil {
		return Field{
			nameKey:     "error",
			typeValue:   FieldString,
			valueString: "nil",
		}
	}
	return Field{
		nameKey:     "error",
		typeValue:   FieldString,
		valueString: valueError.Error(),
	}
}
func Errors(valueErrors []error) Field {
	valueStrings := make([]string, len(valueErrors))
	for i, err := range valueErrors {
		if err == nil {
			valueStrings[i] = "nil"
		} else {
			valueStrings[i] = err.Error()
		}
	}
	return Field{
		nameKey:      "errors",
		typeValue:    FieldStrings,
		valueStrings: valueStrings,
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
		typeValue:     FieldFloats64,
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
		typeValue: FieldInts,
		valueInts: valueInts,
	}
}
func Int64(nameKey string, valueInt64 int64) Field {
	return Field{
		nameKey:    nameKey,
		typeValue:  FieldInt64,
		valueInt64: valueInt64,
	}
}
func Ints64(nameKey string, valueInts64 []int64) Field {
	return Field{
		nameKey:     nameKey,
		typeValue:   FieldInts64,
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
		typeValue:    FieldStrings,
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
		typeValue:  FieldTimes,
		valueTimes: valueTimes,
	}
}

// Публичные функции
func WithExtractor(keys ...string) telemetryOptions {
	return func(universalTelemetry *universalTelemetry) {
		universalTelemetry.extractor = func(ctx context.Context) []Field {
			if ctx == nil {
				return nil
			}
			fields := make([]Field, 0, len(keys))
			for _, key := range keys {
				if val := ctx.Value(key); val != nil {
					switch v := val.(type) {
					case string:
						fields = append(fields, String(key, v))
					case int:
						fields = append(fields, Int(key, v))
					case int64:
						fields = append(fields, Int64(key, v))
					case float32:
						fields = append(fields, Float64(key, float64(v)))
					case float64:
						fields = append(fields, Float64(key, v))
					case bool:
						fields = append(fields, Bool(key, v))
					case time.Time:
						fields = append(fields, Time(key, v))
					case time.Duration:
						fields = append(fields, Duration(key, v))
					default:
						fields = append(fields, String(key, fmt.Sprintf("%v", v)))
					}
				}
			}
			return fields
		}
	}
}
func WithFormat(format TypeFormat) telemetryOptions {
	return func(universalTelemetry *universalTelemetry) {
		universalTelemetry.format.Store(int32(format))
	}
}
func WithLevel(level TypeLevel) telemetryOptions {
	return func(universalTelemetry *universalTelemetry) {
		universalTelemetry.level.Store(int32(level))
	}
}
func WithMode(mode TypeMode, writer io.Writer, bufferSize ...int) telemetryOptions {
	return func(universalTelemetry *universalTelemetry) {
		switch mode {
		case ModeAsync:
			size := defaultBufferSize
			if len(bufferSize) > 0 && bufferSize[0] >= 0 {
				size = bufferSize[0]
			}
			universalTelemetry.mode = ModeAsync
			universalTelemetry.writer = newAsyncWriter(writer, size)
		case ModeSync:
			universalTelemetry.mode = ModeSync
			universalTelemetry.writer = writer
		}
	}
}
func WithTheme(theme TypeTheme) telemetryOptions {
	return func(universalTelemetry *universalTelemetry) {
		switch theme {
		case ThemeDark:
			universalTelemetry.theme = themeDark
		case ThemeLight:
			universalTelemetry.theme = themeLight
		}
	}
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
func (standardTelemetry *standardTelemetry) Write(p []byte) (n int, err error) {
	standardTelemetry.mutex.Lock()
	defer standardTelemetry.mutex.Unlock()
	start := 0
	end := len(p)
	for start < end && p[start] <= ' ' {
		start++
	}
	for end > start && p[end-1] <= ' ' {
		end--
	}
	if start >= end {
		return 0, nil
	}
	if standardTelemetry.isIgnored(p[start:end]) {
		return len(p), nil
	}
	message := string(p[start:end])
	switch TypeLevel(standardTelemetry.level.Load()) {
	case LevelDebug:
		standardTelemetry.telemetry.Debug(DataLog, String("message", message))
	case LevelInfo:
		standardTelemetry.telemetry.Info(DataLog, String("message", message))
	case LevelWarn:
		standardTelemetry.telemetry.Warn(DataLog, String("message", message))
	case LevelError:
		standardTelemetry.telemetry.Error(DataLog, String("message", message))
	case LevelFatal:
		standardTelemetry.telemetry.Fatal(DataLog, String("message", message))
	}
	return len(p), nil
}
func (universalTelemetry *universalTelemetry) Close() error {
	universalTelemetry.mutex.RLock()
	writer := universalTelemetry.writer
	universalTelemetry.mutex.RUnlock()
	if universalTelemetry.mode == ModeAsync {
		if asyncWriter, ok := writer.(*asyncWriter); ok {
			return asyncWriter.Close()
		}
	}
	if closer, ok := writer.(io.Closer); ok {
		if writer == DefaultWriterErr || writer == DefaultWriterOut {
			return nil
		}
		return closer.Close()
	}
	return nil
}
func (universalTelemetry *universalTelemetry) Debug(typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelDebug),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelDebug,
	}
	universalTelemetry.route(context.Background(), attributes, fields)
}
func (universalTelemetry *universalTelemetry) DebugWithContext(context context.Context, typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelDebug),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelDebug,
	}
	universalTelemetry.route(context, attributes, fields)
}
func (universalTelemetry *universalTelemetry) Error(typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelError),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelError,
	}
	universalTelemetry.route(context.Background(), attributes, fields)
}
func (universalTelemetry *universalTelemetry) ErrorWithContext(context context.Context, typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelError),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelError,
	}
	universalTelemetry.route(context, attributes, fields)
}
func (universalTelemetry *universalTelemetry) Fatal(typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelFatal),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelFatal,
	}
	universalTelemetry.route(context.Background(), attributes, fields)
	if universalTelemetry.mode == ModeAsync {
		universalTelemetry.Sync()
	}
	osExit(1)
}
func (universalTelemetry *universalTelemetry) FatalWithContext(context context.Context, typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelFatal),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelFatal,
	}
	universalTelemetry.route(context, attributes, fields)
	if universalTelemetry.mode == ModeAsync {
		universalTelemetry.Sync()
	}
	osExit(1)
}
func (universalTelemetry *universalTelemetry) Info(typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelInfo),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelInfo,
	}
	universalTelemetry.route(context.Background(), attributes, fields)
}
func (universalTelemetry *universalTelemetry) InfoWithContext(context context.Context, typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelInfo),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelInfo,
	}
	universalTelemetry.route(context, attributes, fields)
}
func (universalTelemetry *universalTelemetry) SetExtractor(keys ...string) {
	universalTelemetry.mutex.Lock()
	defer universalTelemetry.mutex.Unlock()
	universalTelemetry.extractor = func(ctx context.Context) []Field {
		if ctx == nil {
			return nil
		}
		fields := make([]Field, 0, len(keys))
		for _, key := range keys {
			if val := ctx.Value(key); val != nil {
				switch v := val.(type) {
				case string:
					fields = append(fields, String(key, v))
				case int:
					fields = append(fields, Int(key, v))
				case int64:
					fields = append(fields, Int64(key, v))
				case float32:
					fields = append(fields, Float64(key, float64(v)))
				case float64:
					fields = append(fields, Float64(key, v))
				case bool:
					fields = append(fields, Bool(key, v))
				case time.Time:
					fields = append(fields, Time(key, v))
				case time.Duration:
					fields = append(fields, Duration(key, v))
				default:
					fields = append(fields, String(key, fmt.Sprintf("%v", v)))
				}
			}
		}
		return fields
	}
}
func (universalTelemetry *universalTelemetry) SetFormat(format TypeFormat) {
	universalTelemetry.format.Store(int32(format))
}
func (universalTelemetry *universalTelemetry) SetLevel(level TypeLevel) {
	universalTelemetry.level.Store(int32(level))
}
func (universalTelemetry *universalTelemetry) SetMode(mode TypeMode, writer io.Writer, bufferSize ...int) {
	universalTelemetry.mutex.Lock()
	defer universalTelemetry.mutex.Unlock()
	if universalTelemetry.mode == ModeAsync {
		if asyncWriter, ok := universalTelemetry.writer.(*asyncWriter); ok {
			asyncWriter.Close()
		}
	}
	switch mode {
	case ModeAsync:
		size := defaultBufferSize
		if len(bufferSize) > 0 && bufferSize[0] >= 0 {
			size = bufferSize[0]
		}
		universalTelemetry.mode = ModeAsync
		universalTelemetry.writer = newAsyncWriter(writer, size)
	case ModeSync:
		universalTelemetry.mode = ModeSync
		universalTelemetry.writer = writer
	}
}
func (universalTelemetry *universalTelemetry) SetTheme(theme TypeTheme) {
	universalTelemetry.mutex.Lock()
	defer universalTelemetry.mutex.Unlock()
	switch theme {
	case ThemeDark:
		universalTelemetry.theme = themeDark
	case ThemeLight:
		universalTelemetry.theme = themeLight
	}
}
func (universalTelemetry *universalTelemetry) Sync() error {
	universalTelemetry.mutex.RLock()
	currentWriter := universalTelemetry.writer
	universalTelemetry.mutex.RUnlock()
	if universalTelemetry.mode == ModeAsync {
		if asyncWriter, ok := currentWriter.(*asyncWriter); ok {
			if err := asyncWriter.Sync(); err != nil {
				return err
			}
			currentWriter = asyncWriter.writer
		}
	}
	if syncer, ok := currentWriter.(interface{ Sync() error }); ok {
		return syncer.Sync()
	}
	return nil
}
func (universalTelemetry *universalTelemetry) Warn(typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelWarn),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelWarn,
	}
	universalTelemetry.route(context.Background(), attributes, fields)
}
func (universalTelemetry *universalTelemetry) WarnWithContext(context context.Context, typeData TypeData, fields ...Field) {
	attributes := writeAttributes{
		caller:     universalTelemetry.getCaller(LevelWarn),
		theme:      universalTelemetry.getTheme(),
		typeData:   typeData,
		typeFormat: TypeFormat(universalTelemetry.format.Load()),
		typeLevel:  LevelWarn,
	}
	universalTelemetry.route(context, attributes, fields)
}
func (universalTelemetry *universalTelemetry) Write(p []byte) (n int, err error) {
	universalTelemetry.mutex.RLock()
	writer := universalTelemetry.writer
	universalTelemetry.mutex.RUnlock()
	return writer.Write(p)
}
