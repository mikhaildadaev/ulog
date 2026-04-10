package ulog

import (
	"context"
	"fmt"
	"io"
	"time"
)

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
func WithExtractor(keys ...string) OptionLogger {
	return func(universalLogger *universalLogger) {
		universalLogger.extractor = func(ctx context.Context) []Field {
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
func WithFormat(format TypeFormat) OptionLogger {
	return func(universalLogger *universalLogger) {
		universalLogger.format.Store(int32(format))
	}
}
func WithLevel(level TypeLevel) OptionLogger {
	return func(universalLogger *universalLogger) {
		universalLogger.level.Store(int32(level))
	}
}
func WithMode(mode TypeMode, writer io.Writer, bufferSize ...int) OptionLogger {
	return func(universalLogger *universalLogger) {
		switch mode {
		case ModeAsync:
			if bufferSize == nil || bufferSize[0] <= 0 {
				bufferSize[0] = defaultBufferSize
			}
			universalLogger.mode = ModeAsync
			universalLogger.writer = newAsyncWriter(writer, bufferSize[0])
		case ModeSync:
			universalLogger.mode = ModeSync
			universalLogger.writer = writer
		}
	}
}
func WithTheme(theme TypeTheme) OptionLogger {
	return func(universalLogger *universalLogger) {
		switch theme {
		case ThemeDark:
			universalLogger.theme = darkTheme
		case ThemeLight:
			universalLogger.theme = lightTheme
		}
	}
}

// Публичные методы
func (standardLogger *standardLogger) Write(p []byte) (n int, err error) {
	standardLogger.mutex.Lock()
	defer standardLogger.mutex.Unlock()
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
	if isIgnoredError(p[start:end]) {
		return len(p), nil
	}
	message := string(p[start:end])
	switch TypeLevel(standardLogger.level.Load()) {
	case LevelDebug:
		standardLogger.logger.Debug(message)
	case LevelInfo:
		standardLogger.logger.Info(message)
	case LevelWarn:
		standardLogger.logger.Warn(message)
	case LevelError:
		standardLogger.logger.Error(message)
	case LevelFatal:
		standardLogger.logger.Fatal(message)
	}
	return len(p), nil
}
func (universalLogger *universalLogger) Close() error {
	if universalLogger.mode != ModeAsync {
		return nil
	}
	universalLogger.mutex.RLock()
	currentWriter := universalLogger.writer
	universalLogger.mutex.RUnlock()
	if asyncWriter, ok := currentWriter.(*asyncWriter); ok {
		return asyncWriter.Close()
	}
	return nil
}
func (universalLogger *universalLogger) Debug(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelDebug, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelDebug, context.Background(), message, fields)
	}
}
func (universalLogger *universalLogger) DebugWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelDebug, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelDebug, context, message, fields)
	}
}
func (universalLogger *universalLogger) Error(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelError, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelError, context.Background(), message, fields)
	}
}
func (universalLogger *universalLogger) ErrorWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelError, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelError, context, message, fields)
	}
}
func (universalLogger *universalLogger) Fatal(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelFatal, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelFatal, context.Background(), message, fields)
	}
	osExit(1)
}
func (universalLogger *universalLogger) FatalWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelFatal, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelFatal, context, message, fields)
	}
	osExit(1)
}
func (universalLogger *universalLogger) Info(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelInfo, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelInfo, context.Background(), message, fields)
	}
}
func (universalLogger *universalLogger) InfoWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelInfo, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelInfo, context, message, fields)
	}
}
func (universalLogger *universalLogger) Warn(message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelWarn, context.Background(), message, fields)
	case FormatText:
		universalLogger.writeText(LevelWarn, context.Background(), message, fields)
	}
}
func (universalLogger *universalLogger) WarnWithContext(context context.Context, message string, fields ...Field) {
	switch TypeFormat(universalLogger.format.Load()) {
	case FormatJson:
		universalLogger.writeJson(LevelWarn, context, message, fields)
	case FormatText:
		universalLogger.writeText(LevelWarn, context, message, fields)
	}
}
func (universalLogger *universalLogger) SetExtractor(keys ...string) {
	universalLogger.mutex.Lock()
	defer universalLogger.mutex.Unlock()
	universalLogger.extractor = func(ctx context.Context) []Field {
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
func (universalLogger *universalLogger) SetFormat(format TypeFormat) {
	universalLogger.format.Store(int32(format))
}
func (universalLogger *universalLogger) SetLevel(level TypeLevel) {
	universalLogger.level.Store(int32(level))
}
func (universalLogger *universalLogger) SetMode(mode TypeMode, writer io.Writer, bufferSize ...int) {
	universalLogger.mutex.Lock()
	defer universalLogger.mutex.Unlock()
	if universalLogger.mode == ModeAsync {
		if asyncWriter, ok := universalLogger.writer.(*asyncWriter); ok {
			asyncWriter.Close()
			time.Sleep(time.Millisecond)
		}
	}
	switch mode {
	case ModeAsync:
		if bufferSize == nil || bufferSize[0] <= 0 {
			bufferSize[0] = defaultBufferSize
		}
		universalLogger.mode = ModeAsync
		universalLogger.writer = newAsyncWriter(writer, bufferSize[0])
	case ModeSync:
		universalLogger.mode = ModeSync
		universalLogger.writer = writer
	}
}
func (universalLogger *universalLogger) SetTheme(theme TypeTheme) {
	universalLogger.mutex.Lock()
	defer universalLogger.mutex.Unlock()
	switch theme {
	case ThemeDark:
		universalLogger.theme = darkTheme
	case ThemeLight:
		universalLogger.theme = lightTheme
	}
}
func (universalLogger *universalLogger) Sync() error {
	if universalLogger.mode != ModeAsync {
		return nil
	}
	universalLogger.mutex.RLock()
	currentWriter := universalLogger.writer
	universalLogger.mutex.RUnlock()
	if asyncWriter, ok := currentWriter.(*asyncWriter); ok {
		return asyncWriter.Sync()
	}
	return nil
}
func (universalLogger *universalLogger) Write(p []byte) (n int, err error) {
	universalLogger.mutex.RLock()
	writer := universalLogger.writer
	universalLogger.mutex.RUnlock()
	return writer.Write(p)
}
