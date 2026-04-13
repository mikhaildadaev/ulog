package ulog

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

// Тесты публичных копонентов
func TestClose(t *testing.T) {
	t.Run("Close/Async", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := NewLogger(WithMode(ModeAsync, buf, 100))
		logger.Info("test message")
		err := logger.Close()
		if err != nil {
			t.Errorf("Close() returned error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "test message") {
			t.Error("Message not written after Close")
		}
	})
	t.Run("Close/Sync", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := NewLogger(WithMode(ModeSync, buf))
		err := logger.Close()
		if err != nil {
			t.Errorf("Close() returned error: %v", err)
		}
		logger.Info("after close")
		if buf.Len() == 0 {
			t.Error("Logger stopped working after Close in sync mode")
		}
	})
}
func TestExtractor(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		context   context.Context
		wantKey   string
		wantValue string
		shouldAdd bool
	}{
		{
			name:      "NullContext",
			keys:      []string{"test_empty"},
			context:   context.Background(),
			wantKey:   "",
			wantValue: "",
			shouldAdd: false,
		},
		{
			name:      "NullKeys",
			keys:      nil,
			context:   context.WithValue(context.Background(), "trace_id", "abc-123"),
			wantKey:   "",
			wantValue: "",
			shouldAdd: false,
		},
		{
			name:      "Bool",
			keys:      []string{"test_bool"},
			context:   context.WithValue(context.Background(), "test_bool", true),
			wantKey:   "test_bool",
			wantValue: "true",
			shouldAdd: true,
		},
		{
			name:      "Duration",
			keys:      []string{"test_duration"},
			context:   context.WithValue(context.Background(), "test_duration", 5*time.Second),
			wantKey:   "test_duration",
			wantValue: "5s",
			shouldAdd: true,
		},
		{
			name:      "Float64",
			keys:      []string{"test_float64"},
			context:   context.WithValue(context.Background(), "test_float64", 3.14159),
			wantKey:   "test_float64",
			wantValue: "3.14159",
			shouldAdd: true,
		},
		{
			name:      "Int",
			keys:      []string{"test_int"},
			context:   context.WithValue(context.Background(), "test_int", int(12345)),
			wantKey:   "test_int",
			wantValue: "12345",
			shouldAdd: true,
		},
		{
			name:      "Int64",
			keys:      []string{"test_int64"},
			context:   context.WithValue(context.Background(), "test_int64", int64(12345)),
			wantKey:   "test_int64",
			wantValue: "12345",
			shouldAdd: true,
		},
		{
			name:      "String",
			keys:      []string{"test_string"},
			context:   context.WithValue(context.Background(), "test_string", "abc-123"),
			wantKey:   "test_string",
			wantValue: "abc-123",
			shouldAdd: true,
		},
		{
			name:      "Time",
			keys:      []string{"test_time"},
			context:   context.WithValue(context.Background(), "test_time", time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)),
			wantKey:   "test_time",
			wantValue: "2026-04-10T12:00:00.000000+00:00",
			shouldAdd: true,
		},
		{
			name:      "Multiple",
			keys:      []string{"trace_id", "user_id"},
			context:   context.WithValue(context.WithValue(context.Background(), "trace_id", "abc-123"), "user_id", int64(12345)),
			wantKey:   "trace_id",
			wantValue: "abc-123",
			shouldAdd: true,
		},
	}
	for _, elem := range tests {
		t.Run("WithExtractor/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(
				WithExtractor(elem.keys...),
				WithFormat(FormatJson),
				WithMode(ModeSync, buf),
			)
			logger.InfoWithContext(elem.context, "test message")
			logger.Sync()
			output := buf.String()
			checkExtractor(t, elem, output)
		})
		t.Run("SetExtractor/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger()
			logger.SetExtractor(elem.keys...)
			logger.SetFormat(FormatJson)
			logger.SetMode(ModeSync, buf)
			logger.InfoWithContext(elem.context, "test message")
			logger.Sync()
			output := buf.String()
			checkExtractor(t, elem, output)
		})
	}
}
func TestField(t *testing.T) {
	t.Run("Bool", func(t *testing.T) {
		val := bool(true)
		field := Bool("test", val)
		checkFieldBool(t, field, val)
	})
	t.Run("Bools", func(t *testing.T) {
		vals := []bool{true, false}
		field := Bools("test", vals)
		checkFieldBools(t, field, vals)
	})
	t.Run("Duration", func(t *testing.T) {
		val := time.Duration(5 * time.Second)
		field := Duration("test", val)
		checkFieldDuration(t, field, val)
	})
	t.Run("Durations", func(t *testing.T) {
		vals := []time.Duration{1 * time.Second, 2 * time.Second}
		field := Durations("test", vals)
		checkFieldDurations(t, field, vals)
	})
	t.Run("Error", func(t *testing.T) {
		val := error(errors.New("err"))
		field := Error(val)
		checkFieldError(t, field, val)
	})
	t.Run("Errors", func(t *testing.T) {
		vals := []error{errors.New("err1"), errors.New("err2")}
		field := Errors(vals)
		checkFieldErrors(t, field, vals)
	})
	t.Run("Float64", func(t *testing.T) {
		val := float64(3.14159)
		field := Float64("test", val)
		checkFieldFloat64(t, field, val)
	})
	t.Run("Floats64", func(t *testing.T) {
		vals := []float64{1.1, 2.2}
		field := Floats64("test", vals)
		checkFieldFloats64(t, field, vals)
	})
	t.Run("Int", func(t *testing.T) {
		val := int(42)
		field := Int("test", val)
		checkFieldInt(t, field, val)
	})
	t.Run("Ints", func(t *testing.T) {
		vals := []int{1, 2}
		field := Ints("test", vals)
		checkFieldInts(t, field, vals)
	})
	t.Run("Int64", func(t *testing.T) {
		val := int64(1 << 62)
		field := Int64("test", val)
		checkFieldInt64(t, field, val)
	})
	t.Run("Ints64", func(t *testing.T) {
		vals := []int64{1, 2}
		field := Ints64("test", vals)
		checkFieldInts64(t, field, vals)
	})
	t.Run("String", func(t *testing.T) {
		val := string("John")
		field := String("test", val)
		checkFieldString(t, field, val)
	})
	t.Run("Strings", func(t *testing.T) {
		vals := []string{"a", "b"}
		field := Strings("test", vals)
		checkFieldStrings(t, field, vals)
	})
	t.Run("Time", func(t *testing.T) {
		val := time.Time(time.Now())
		field := Time("test", val)
		checkFieldTime(t, field, val)
	})
	t.Run("Times", func(t *testing.T) {
		vals := []time.Time{time.Now(), time.Now().Add(time.Hour)}
		field := Times("test", vals)
		checkFieldTimes(t, field, vals)
	})
}
func TestFormat(t *testing.T) {
	array := []struct {
		name   string
		format TypeFormat
		expect string
	}{
		{"Json", FormatJson, `"message":"test message"`},
		{"Text", FormatText, "test message"},
	}
	for _, elem := range array {
		t.Run("WithFormat/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(
				WithFormat(elem.format),
				WithMode(ModeSync, buf, 0),
			)
			logger.Info("test message")
			logger.Sync()
			output := buf.String()
			if !strings.Contains(output, elem.expect) {
				t.Errorf("Expected output to contain %q, got %q", elem.expect, output)
			}
		})
		t.Run("SetFormat/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger()
			logger.SetFormat(elem.format)
			logger.SetMode(ModeSync, buf, 0)
			logger.Info("test message")
			logger.Sync()
			output := buf.String()
			if !strings.Contains(output, elem.expect) {
				t.Errorf("Expected output to contain %q, got %q", elem.expect, output)
			}
		})
	}
}
func TestLevel(t *testing.T) {
	array := []struct {
		name      string
		level     TypeLevel
		logFunc   func(Logger)
		shouldLog bool
	}{
		// DEBUG
		{"DEBUG->DEBUG", LevelDebug, testDebug, true},
		{"DEBUG->INFO", LevelDebug, testInfo, true},
		{"DEBUG->WARN", LevelDebug, testWarn, true},
		{"DEBUG->ERROR", LevelDebug, testError, true},
		{"DEBUG->FATAL", LevelDebug, testFatal, true},
		// INFO
		{"INFO->DEBUG", LevelInfo, testDebug, false},
		{"INFO->INFO", LevelInfo, testInfo, true},
		{"INFO->WARN", LevelInfo, testWarn, true},
		{"INFO->ERROR", LevelInfo, testError, true},
		{"INFO->FATAL", LevelInfo, testFatal, true},
		// WARN
		{"WARN->DEBUG", LevelWarn, testDebug, false},
		{"WARN->INFO", LevelWarn, testInfo, false},
		{"WARN->WARN", LevelWarn, testWarn, true},
		{"WARN->ERROR", LevelWarn, testError, true},
		{"WARN->FATAL", LevelWarn, testFatal, true},
		// ERROR
		{"ERROR->DEBUG", LevelError, testDebug, false},
		{"ERROR->INFO", LevelError, testInfo, false},
		{"ERROR->WARN", LevelError, testWarn, false},
		{"ERROR->ERROR", LevelError, testError, true},
		{"ERROR->FATAL", LevelError, testFatal, true},
		// FATAL
		{"FATAL->DEBUG", LevelFatal, testDebug, false},
		{"FATAL->INFO", LevelFatal, testInfo, false},
		{"FATAL->WARN", LevelFatal, testWarn, false},
		{"FATAL->ERROR", LevelFatal, testError, false},
		{"FATAL->FATAL", LevelFatal, testFatal, true},
	}
	for _, elem := range array {
		t.Run("WithLevel/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(
				WithLevel(elem.level),
				WithMode(ModeSync, buf, 0),
			)
			elem.logFunc(logger)
			logger.Sync()
			if elem.shouldLog && buf.Len() == 0 {
				t.Error("Expected log to be written, but got nothing")
			}
			if !elem.shouldLog && buf.Len() > 0 {
				t.Error("Expected no log, but got output")
			}
		})
		t.Run("SetLevel/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger()
			logger.SetLevel(elem.level)
			logger.SetMode(ModeSync, buf, 0)
			elem.logFunc(logger)
			logger.Sync()
			if elem.shouldLog && buf.Len() == 0 {
				t.Error("Expected log to be written, but got nothing")
			}
			if !elem.shouldLog && buf.Len() > 0 {
				t.Error("Expected no log, but got output")
			}
		})
	}
}
func TestLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(
		WithFormat(FormatText),
		WithMode(ModeSync, buf),
	)
	if logger == nil {
		t.Fatal("NewLogger returned nil")
	}
	logger.Info("test message")
	logger.Sync()
	if buf.Len() == 0 {
		t.Error("Logger produced no buf")
	}
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Expected 'test message', got %q", buf.String())
	}
}
func TestLoggerLog(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := NewLogger(
		WithFormat(FormatText),
		WithMode(ModeSync, buf),
	)
	loggerLog := NewLoggerLog(LevelInfo, logger)
	loggerLog.Print("test message")
	if buf.Len() == 0 {
		t.Error("Logger produced no buf")
	}
	if !strings.Contains(buf.String(), "test message") {
		t.Errorf("Expected 'test message', got %q", buf.String())
	}
}
func TestMethod(t *testing.T) {
	array := []struct {
		name      string
		logFunc   func(Logger)
		level     TypeLevel
		shouldLog bool
	}{
		// DEBUG
		{"DEBUG", testDebug, LevelDebug, true},
		{"DEBUG/WithContext", testDebugWithContext, LevelDebug, true},
		// ERROR
		{"ERROR", testError, LevelError, true},
		{"ERROR/WithContext", testErrorWithContext, LevelError, true},
		// FATAL
		{"FATAL", testFatal, LevelFatal, true},
		{"FATAL/WithContext", testFatalWithContext, LevelFatal, true},
		// INFO
		{"INFO", testInfo, LevelInfo, true},
		{"INFO/WithContext", testInfoWithContext, LevelInfo, true},
		// WARN
		{"WARN", testWarn, LevelWarn, true},
		{"WARN/WithContext", testWarnWithContext, LevelWarn, true},
	}
	for _, elem := range array {
		t.Run(elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(
				WithMode(ModeSync, buf),
				WithLevel(elem.level),
			)
			elem.logFunc(logger)
			logger.Sync()
			output := buf.String()
			if elem.shouldLog && !strings.Contains(output, "test message") {
				t.Errorf("Expected message not found in output: %q", output)
			}
		})
	}
}
func TestMode(t *testing.T) {
	t.Run("WithMode/Async", func(t *testing.T) {
		writerBuf := &bytes.Buffer{}
		logger := NewLogger(
			WithMode(ModeAsync, writerBuf, 1000),
		)
		logger.Info("test message")
		logger.Sync()
		if writerBuf.Len() == 0 {
			t.Error("Async mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test message") {
			t.Error("Async mode: expected message not found")
		}
	})
	t.Run("SetMode/Async", func(t *testing.T) {
		writerBuf := &bytes.Buffer{}
		logger := NewLogger()
		logger.SetMode(ModeAsync, writerBuf, 1000)
		logger.Info("test message")
		logger.Sync()
		if writerBuf.Len() == 0 {
			t.Error("Async mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test message") {
			t.Error("Async mode: expected message not found")
		}
	})
	t.Run("WithMode/Sync", func(t *testing.T) {
		writerBuf := &bytes.Buffer{}
		logger := NewLogger(
			WithMode(ModeSync, writerBuf),
		)
		logger.Info("test message")
		logger.Sync()
		if writerBuf.Len() == 0 {
			t.Error("Sync mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test message") {
			t.Error("Sync mode: expected message not found")
		}
	})
	t.Run("SetMode/Sync", func(t *testing.T) {
		writerBuf := &bytes.Buffer{}
		logger := NewLogger()
		logger.SetMode(ModeSync, writerBuf)
		logger.Info("test message")
		logger.Sync()
		if writerBuf.Len() == 0 {
			t.Error("Sync mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test message") {
			t.Error("Sync mode: expected message not found")
		}
	})
}
func TestSync(t *testing.T) {
	t.Run("Sync/Async", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := NewLogger(WithMode(ModeAsync, buf, 1000))
		defer logger.Close()
		logger.Info("test message")
		err := logger.Sync()
		if err != nil {
			t.Errorf("Sync() returned error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "test message") {
			t.Error("Message not written after Sync")
		}
	})
	t.Run("Sync/Sync", func(t *testing.T) {
		buf := &bytes.Buffer{}
		logger := NewLogger(WithMode(ModeSync, buf))
		defer logger.Close()
		logger.Info("test message")
		err := logger.Sync()
		if err != nil {
			t.Errorf("Sync() returned error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "test message") {
			t.Error("Message not written after Sync")
		}
	})
}
func TestTheme(t *testing.T) {
	array := []struct {
		name         string
		theme        TypeTheme
		callerColor  string
		messageColor string
		prefixDebug  string
		prefixError  string
		prefixFatal  string
		prefixInfo   string
		prefixWarn   string
		reset        string
	}{
		{
			name:         "Dark",
			theme:        ThemeDark,
			callerColor:  colorDarkBlue,
			messageColor: colorDarkWhite,
			prefixDebug:  colorDarkCyan + "[DEBUG]",
			prefixError:  colorDarkRed + "[ERROR]",
			prefixFatal:  colorDarkPurple + "[FATAL]",
			prefixInfo:   colorDarkGreen + "[INFO]",
			prefixWarn:   colorDarkYellow + "[WARN]",
			reset:        colorReset,
		},
		{
			name:         "Light",
			theme:        ThemeLight,
			callerColor:  colorLightBlue,
			messageColor: colorLightBlack,
			prefixDebug:  colorLightCyan + "[DEBUG]",
			prefixError:  colorLightRed + "[ERROR]",
			prefixFatal:  colorLightPurple + "[FATAL]",
			prefixInfo:   colorLightGreen + "[INFO]",
			prefixWarn:   colorLightYellow + "[WARN]",
			reset:        colorReset,
		},
	}
	for _, elem := range array {
		t.Run("WithTheme/"+elem.name, func(t *testing.T) {
			testLevel := func(level string, logFunc func(Logger), expectedPrefix string) {
				buf := &bytes.Buffer{}
				logger := NewLogger(
					WithLevel(LevelDebug),
					WithMode(ModeSync, buf),
					WithTheme(elem.theme),
				)
				logFunc(logger)
				logger.Sync()
				output := buf.String()
				checkTheme(t, level, expectedPrefix, elem, output)
			}
			testLevel("Debug", testDebug, elem.prefixDebug)
			testLevel("Error", testError, elem.prefixError)
			testLevel("Fatal", testFatal, elem.prefixFatal)
			testLevel("Info", testInfo, elem.prefixInfo)
			testLevel("Warn", testWarn, elem.prefixWarn)
		})
		t.Run("SetTheme/"+elem.name, func(t *testing.T) {
			testLevel := func(level string, logFunc func(Logger), expectedPrefix string) {
				buf := &bytes.Buffer{}
				logger := NewLogger()
				logger.SetLevel(LevelDebug)
				logger.SetMode(ModeSync, buf)
				logger.SetTheme(elem.theme)
				logFunc(logger)
				logger.Sync()
				output := buf.String()
				checkTheme(t, level, expectedPrefix, elem, output)
			}
			testLevel("Debug", testDebug, elem.prefixDebug)
			testLevel("Error", testError, elem.prefixError)
			testLevel("Fatal", testFatal, elem.prefixFatal)
			testLevel("Info", testInfo, elem.prefixInfo)
			testLevel("Warn", testWarn, elem.prefixWarn)
		})
	}
}

// Тесты приватных копонентов
func TestIsIgnoredError(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected bool
	}{
		{"EOF", []byte("read: EOF"), true},
		{"TLS handshake", []byte("TLS handshake error"), true},
		{"Connection refused", []byte("dial: connection refused"), true},
		{"Timeout", []byte("i/o timeout"), true},
		{"Normal message", []byte("user logged in"), false},
		{"Empty", []byte{}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isIgnoredError(tt.data); got != tt.expected {
				t.Errorf("isIgnoredError(%q) = %v, want %v", tt.data, got, tt.expected)
			}
		})
	}
}

// Приватные функции
func checkExtractor(t *testing.T, elem struct {
	name      string
	keys      []string
	context   context.Context
	wantKey   string
	wantValue string
	shouldAdd bool
}, output string) {
	t.Helper()
	if elem.shouldAdd {
		if !strings.Contains(output, elem.wantKey) {
			t.Errorf("extractor with keys %v: expected field %q not found in output: %s",
				elem.keys, elem.wantKey, output)
		}
		if !strings.Contains(output, elem.wantValue) {
			t.Errorf("extractor with keys %v: expected value %q for key %q not found in output: %s",
				elem.keys, elem.wantValue, elem.wantKey, output)
		}
	} else {
		for _, key := range elem.keys {
			if strings.Contains(output, key) {
				t.Errorf("extractor with keys %v: unexpected field %q found in output: %s",
					elem.keys, key, output)
			}
		}
		if elem.keys == nil && strings.Contains(output, "trace_id") {
			t.Errorf("extractor with nil keys: unexpected field 'trace_id' found in output: %s", output)
		}
	}
}
func checkFieldBool(t *testing.T, field Field, val bool) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldBool {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if field.valueBool != val {
		t.Errorf("Expected valueBool, got %v", field.valueBool)
	}
}
func checkFieldBools(t *testing.T, field Field, vals []bool) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldBools {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if len(field.valueBools) != len(vals) {
		t.Errorf("Expected valueBools, got %v", len(field.valueBools))
	}
}
func checkFieldDuration(t *testing.T, field Field, val time.Duration) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldDuration {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if field.valueDuration != val {
		t.Errorf("Expected valueDuration, got %v", field.valueDuration)
	}
}
func checkFieldDurations(t *testing.T, field Field, vals []time.Duration) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldDurations {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if len(field.valueDurations) != len(vals) {
		t.Errorf("Expected valueDurations, got %v", field.valueDurations)
	}
}
func checkFieldError(t *testing.T, field Field, val error) {
	t.Helper()
	if field.nameKey != "error" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldString {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if field.valueString != val.Error() {
		t.Errorf("Expected valueString, got %v", field.valueString)
	}
}
func checkFieldErrors(t *testing.T, field Field, vals []error) {
	t.Helper()
	if field.nameKey != "errors" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldStrings {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if len(field.valueStrings) != len(vals) {
		t.Errorf("Expected valueStrings, got %v", field.valueStrings)
	}
}
func checkFieldFloat64(t *testing.T, field Field, val float64) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldFloat64 {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if field.valueFloat64 != val {
		t.Errorf("Expected valueFloat64, got %v", field.valueFloat64)
	}
}
func checkFieldFloats64(t *testing.T, field Field, vals []float64) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldFloats64 {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if len(field.valueFloats64) != len(vals) {
		t.Errorf("Expected valueFloats64, got %v", field.valueFloats64)
	}
}
func checkFieldInt(t *testing.T, field Field, val int) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldInt {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if field.valueInt != val {
		t.Errorf("Expected valueInt, got %v", field.valueInt)
	}
}
func checkFieldInts(t *testing.T, field Field, vals []int) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldInts {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if len(field.valueInts) != len(vals) {
		t.Errorf("Expected valueInts, got %v", field.valueInts)
	}
}
func checkFieldInt64(t *testing.T, field Field, val int64) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldInt64 {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if field.valueInt64 != val {
		t.Errorf("Expected valueInt64, got %v", field.valueInt64)
	}
}
func checkFieldInts64(t *testing.T, field Field, vals []int64) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldInts64 {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if len(field.valueInts64) != len(vals) {
		t.Errorf("Expected valueInts64, got %v", field.valueInts64)
	}
}
func checkFieldString(t *testing.T, field Field, val string) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldString {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if field.valueString != val {
		t.Errorf("Expected valueString, got %v", field.valueString)
	}
}
func checkFieldStrings(t *testing.T, field Field, vals []string) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldStrings {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if len(field.valueStrings) != len(vals) {
		t.Errorf("Expected valueStrings, got %v", field.valueStrings)
	}
}
func checkFieldTime(t *testing.T, field Field, val time.Time) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldTime {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if field.valueTime != val {
		t.Errorf("Expected valueTime, got %v", field.valueTime)
	}
}
func checkFieldTimes(t *testing.T, field Field, vals []time.Time) {
	t.Helper()
	if field.nameKey != "test" {
		t.Errorf("Expected nameKey, got '%s'", field.nameKey)
	}
	if field.typeValue != FieldTimes {
		t.Errorf("Expected typeValue, got %d", field.typeValue)
	}
	if len(field.valueTimes) != len(vals) {
		t.Errorf("Expected valueTimes, got %v", field.valueTimes)
	}
}
func checkTheme(t *testing.T, level, expectedPrefix string, elem struct {
	name         string
	theme        TypeTheme
	callerColor  string
	messageColor string
	prefixDebug  string
	prefixError  string
	prefixFatal  string
	prefixInfo   string
	prefixWarn   string
	reset        string
}, output string) {
	t.Helper()
	if !strings.Contains(output, elem.callerColor) && level == "DEBUG" {
		t.Errorf("%s: expected prefix %q not found in %q", level, elem.callerColor, output)
	}
	if !strings.Contains(output, expectedPrefix) {
		t.Errorf("%s: expected prefix %q not found in %q", level, expectedPrefix, output)
	}
	if !strings.Contains(output, elem.messageColor) {
		t.Errorf("%s: expected message color %q not found", level, elem.messageColor)
	}
	if !strings.Contains(output, elem.reset) {
		t.Errorf("%s: expected message color %q not found", level, elem.reset)
	}
}
func testDebug(l Logger) {
	l.Debug("test message")
}
func testDebugWithContext(l Logger) {
	l.DebugWithContext(context.Background(), "test message")
}
func testError(l Logger) {
	l.Error("test message")
}
func testErrorWithContext(l Logger) {
	l.ErrorWithContext(context.Background(), "test message")
}
func testFatal(l Logger) {
	oldExit := osExit
	osExit = func(int) {}
	defer func() { osExit = oldExit }()
	l.Fatal("test message")
}
func testFatalWithContext(l Logger) {
	oldExit := osExit
	osExit = func(int) {}
	defer func() { osExit = oldExit }()
	l.FatalWithContext(context.Background(), "test message")
}
func testInfo(l Logger) {
	l.Info("test message")
}
func testInfoWithContext(l Logger) {
	l.InfoWithContext(context.Background(), "test message")
}
func testWarn(l Logger) {
	l.Warn("test message")
}
func testWarnWithContext(l Logger) {
	l.WarnWithContext(context.Background(), "test message")
}
