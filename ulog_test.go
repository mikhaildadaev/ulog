package ulog

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"
)

// Тесты публичных копонентов
func TestFields(t *testing.T) {
	t.Run("Bool", func(t *testing.T) {
		f := Bool("active", true)
		if f.nameKey != "active" {
			t.Errorf("Expected nameKey 'active', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldBool {
			t.Errorf("Expected typeValue FieldBool, got %d", f.typeValue)
		}
		if f.valueBool != true {
			t.Errorf("Expected valueBool true, got %t", f.valueBool)
		}
	})
	t.Run("Bools", func(t *testing.T) {
		vals := []bool{true, false, true}
		f := Bools("flags", vals)
		if f.nameKey != "flags" {
			t.Errorf("Expected nameKey 'flags', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldBool {
			t.Errorf("Expected typeValue FieldBool, got %d", f.typeValue)
		}
		if len(f.valueBools) != 3 {
			t.Errorf("Expected 3 values, got %d", len(f.valueBools))
		}
		if f.valueBools[0] != true || f.valueBools[1] != false {
			t.Error("Bools values incorrect")
		}
	})
	t.Run("Duration", func(t *testing.T) {
		d := 5 * time.Second
		f := Duration("timeout", d)
		if f.nameKey != "timeout" {
			t.Errorf("Expected nameKey 'timeout', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldDuration {
			t.Errorf("Expected typeValue FieldDuration, got %d", f.typeValue)
		}
		if f.valueDuration != d {
			t.Errorf("Expected duration %v, got %v", d, f.valueDuration)
		}
	})
	t.Run("Durations", func(t *testing.T) {
		vals := []time.Duration{1 * time.Second, 2 * time.Second}
		f := Durations("latencies", vals)
		if f.nameKey != "latencies" {
			t.Errorf("Expected nameKey 'latencies', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldDuration {
			t.Errorf("Expected typeValue FieldDuration, got %d", f.typeValue)
		}
		if len(f.valueDurations) != 2 {
			t.Errorf("Expected 2 values, got %d", len(f.valueDurations))
		}
	})
	t.Run("Error", func(t *testing.T) {
		f := Err(nil)
		if f.nameKey != "error" {
			t.Errorf("Expected nameKey 'error', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldString {
			t.Errorf("Expected typeValue FieldString, got %d", f.typeValue)
		}
		if f.valueString != "nil" {
			t.Errorf("Expected 'nil', got '%s'", f.valueString)
		}
		err := errors.New("something failed")
		f = Err(err)
		if f.valueString != "something failed" {
			t.Errorf("Expected 'something failed', got '%s'", f.valueString)
		}
	})
	t.Run("Errors", func(t *testing.T) {
		errs := []error{
			nil,
			errors.New("err1"),
			errors.New("err2"),
		}
		f := Errs(errs)
		if f.nameKey != "errors" {
			t.Errorf("Expected nameKey 'errors', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldString {
			t.Errorf("Expected typeValue FieldString, got %d", f.typeValue)
		}
		if len(f.valueStrings) != 3 {
			t.Errorf("Expected 3 values, got %d", len(f.valueStrings))
		}
		if f.valueStrings[0] != "nil" {
			t.Error("First error should be 'nil'")
		}
		if f.valueStrings[1] != "err1" {
			t.Error("Second error incorrect")
		}
	})
	t.Run("Float64", func(t *testing.T) {
		f := Float64("pi", 3.14159)
		if f.nameKey != "pi" {
			t.Errorf("Expected nameKey 'pi', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldFloat64 {
			t.Errorf("Expected typeValue FieldFloat64, got %d", f.typeValue)
		}
		if f.valueFloat64 != 3.14159 {
			t.Errorf("Expected 3.14159, got %f", f.valueFloat64)
		}
	})
	t.Run("Floats64", func(t *testing.T) {
		vals := []float64{1.1, 2.2, 3.3}
		f := Floats64("values", vals)
		if len(f.valueFloats64) != 3 {
			t.Errorf("Expected 3 values, got %d", len(f.valueFloats64))
		}
	})
	t.Run("Int", func(t *testing.T) {
		f := Int("count", 42)
		if f.nameKey != "count" {
			t.Errorf("Expected nameKey 'count', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldInt {
			t.Errorf("Expected typeValue FieldInt, got %d", f.typeValue)
		}
		if f.valueInt != 42 {
			t.Errorf("Expected 42, got %d", f.valueInt)
		}
	})
	t.Run("Ints", func(t *testing.T) {
		vals := []int{1, 2, 3}
		f := Ints("ids", vals)
		if len(f.valueInts) != 3 {
			t.Errorf("Expected 3 values, got %d", len(f.valueInts))
		}
	})
	t.Run("Int64", func(t *testing.T) {
		f := Int64("big", 1<<62)
		if f.typeValue != FieldInt64 {
			t.Errorf("Expected typeValue FieldInt, got %d", f.typeValue)
		}
		if f.valueInt64 != 1<<62 {
			t.Errorf("Expected %d, got %d", 1<<62, f.valueInt64)
		}
	})
	t.Run("Ints64", func(t *testing.T) {
		vals := []int64{1, 2, 3}
		f := Ints64("ids64", vals)
		if len(f.valueInts64) != 3 {
			t.Errorf("Expected 3 values, got %d", len(f.valueInts64))
		}
	})
	t.Run("String", func(t *testing.T) {
		f := String("name", "John")
		if f.nameKey != "name" {
			t.Errorf("Expected nameKey 'name', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldString {
			t.Errorf("Expected typeValue FieldString, got %d", f.typeValue)
		}
		if f.valueString != "John" {
			t.Errorf("Expected 'John', got '%s'", f.valueString)
		}
	})
	t.Run("Strings", func(t *testing.T) {
		vals := []string{"a", "b", "c"}
		f := Strings("letters", vals)
		if len(f.valueStrings) != 3 {
			t.Errorf("Expected 3 values, got %d", len(f.valueStrings))
		}
	})
	t.Run("Time", func(t *testing.T) {
		now := time.Now()
		f := Time("created", now)
		if f.nameKey != "created" {
			t.Errorf("Expected nameKey 'created', got '%s'", f.nameKey)
		}
		if f.typeValue != FieldTime {
			t.Errorf("Expected typeValue FieldTime, got %d", f.typeValue)
		}
		if !f.valueTime.Equal(now) {
			t.Errorf("Time values don't match")
		}
	})
	t.Run("Times", func(t *testing.T) {
		now := time.Now()
		vals := []time.Time{now, now.Add(time.Hour)}
		f := Times("timestamps", vals)
		if len(f.valueTimes) != 2 {
			t.Errorf("Expected 2 values, got %d", len(f.valueTimes))
		}
	})
}
func TestMethods(t *testing.T) {
	array := []struct {
		name      string
		logFunc   func(Logger)
		level     TypeLevel
		shouldLog bool
	}{
		// DEBUG
		{"Debug", testDebug, LevelDebug, true},
		{"DebugWithContext", testDebugWithContext, LevelDebug, true},
		// ERROR
		{"Error", testError, LevelError, true},
		{"ErrorWithContext", testErrorWithContext, LevelError, true},
		// FATAL
		{"Fatal", testFatal, LevelFatal, true},
		{"FatalWithContext", testFatalWithContext, LevelFatal, true},
		// INFO
		{"Info", testInfo, LevelInfo, true},
		{"InfoWithContext", testInfoWithContext, LevelInfo, true},
		// WARN
		{"Warn", testWarn, LevelWarn, true},
		{"WarnWithContext", testWarnWithContext, LevelWarn, true},
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
			if elem.shouldLog && !strings.Contains(output, "message") {
				t.Errorf("Expected message not found in output: %q", output)
			}
		})
	}
}
func TestWithExtractor(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		context   context.Context
		wantKey   string
		wantValue string
		shouldAdd bool
	}{
		{
			name:      "extractor with empty keys",
			keys:      nil,
			context:   context.WithValue(context.Background(), "trace_id", "abc-123"),
			wantKey:   "",
			wantValue: "",
			shouldAdd: false,
		},
		{
			name:      "extractor with empty field from context",
			keys:      []string{"test_empty"},
			context:   context.Background(),
			wantKey:   "",
			wantValue: "",
			shouldAdd: false,
		},
		{
			name:      "extract with bool field from context",
			keys:      []string{"test_bool"},
			context:   context.WithValue(context.Background(), "test_bool", true),
			wantKey:   "test_bool",
			wantValue: "true",
			shouldAdd: true,
		},
		{
			name:      "extract with duration field from context",
			keys:      []string{"test_duration"},
			context:   context.WithValue(context.Background(), "test_duration", 5*time.Second),
			wantKey:   "test_duration",
			wantValue: "5s",
			shouldAdd: true,
		},
		{
			name:      "extract with float64 field from context",
			keys:      []string{"test_float64"},
			context:   context.WithValue(context.Background(), "test_float64", 3.14159),
			wantKey:   "test_float64",
			wantValue: "3.14159",
			shouldAdd: true,
		},
		{
			name:      "extract with int field from context",
			keys:      []string{"test_int"},
			context:   context.WithValue(context.Background(), "test_int", int(12345)),
			wantKey:   "test_int",
			wantValue: "12345",
			shouldAdd: true,
		},
		{
			name:      "extract with int64 field from context",
			keys:      []string{"test_int64"},
			context:   context.WithValue(context.Background(), "test_int64", int64(12345)),
			wantKey:   "test_int64",
			wantValue: "12345",
			shouldAdd: true,
		},
		{
			name:      "extract with string field from context",
			keys:      []string{"test_string"},
			context:   context.WithValue(context.Background(), "test_string", "abc-123"),
			wantKey:   "test_string",
			wantValue: "abc-123",
			shouldAdd: true,
		},
		{
			name:      "extract with time field from context",
			keys:      []string{"test_time"},
			context:   context.WithValue(context.Background(), "test_time", time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)),
			wantKey:   "test_time",
			wantValue: "2026-04-10T12:00:00.000000+00:00",
			shouldAdd: true,
		},
		{
			name:      "extract with multiple fields",
			keys:      []string{"trace_id", "user_id"},
			context:   context.WithValue(context.WithValue(context.Background(), "trace_id", "abc-123"), "user_id", int64(12345)),
			wantKey:   "trace_id",
			wantValue: "abc-123",
			shouldAdd: true,
		},
	}
	for _, elem := range tests {
		t.Run(elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(
				WithExtractor(elem.keys...),
				WithFormat(FormatJson),
				WithMode(ModeSync, buf),
			)
			logger.InfoWithContext(elem.context, "test message")
			logger.Sync()
			output := buf.String()
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
		})
	}
}
func TestWithFormat(t *testing.T) {
	array := []struct {
		name   string
		format TypeFormat
		expect string
	}{
		{"Json", FormatJson, `"message":"test"`},
		{"Text", FormatText, "test"},
	}
	for _, elem := range array {
		t.Run(elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(
				WithFormat(elem.format),
				WithMode(ModeSync, buf, 0),
			)
			logger.Info("test")
			output := buf.String()
			if !strings.Contains(output, elem.expect) {
				t.Errorf("Expected output to contain %q, got %q", elem.expect, output)
			}
		})
	}
}
func TestWithLevel(t *testing.T) {
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
		t.Run(elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger(
				WithLevel(elem.level),
				WithMode(ModeSync, buf, 0),
			)
			elem.logFunc(logger)
			if elem.shouldLog && buf.Len() == 0 {
				t.Error("Expected log to be written, but got nothing")
			}
			if !elem.shouldLog && buf.Len() > 0 {
				t.Error("Expected no log, but got output")
			}
		})
	}
}
func TestWithMode(t *testing.T) {
	t.Run("Async mode", func(t *testing.T) {
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
	t.Run("Sync mode", func(t *testing.T) {
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
}
func TestWithTheme(t *testing.T) {
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
			name:         "Dark theme",
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
			name:         "Light theme",
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
		t.Run(elem.name, func(t *testing.T) {
			testLevel := func(level string, logFunc func(Logger), expectedPrefix string) {
				t.Helper()
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				defer r.Close()
				logger := NewLogger(
					WithLevel(LevelDebug),
					WithMode(ModeSync, w),
					WithTheme(elem.theme),
				)
				logFunc(logger)
				w.Close()
				output, err := io.ReadAll(r)
				if err != nil {
					t.Fatal(err)
				}
				outputStr := string(output)
				if !strings.Contains(outputStr, elem.callerColor) && level == "DEBUG" {
					t.Errorf("%s: expected prefix %q not found in %q", level, elem.callerColor, outputStr)
				}
				if !strings.Contains(outputStr, expectedPrefix) {
					t.Errorf("%s: expected prefix %q not found in %q", level, expectedPrefix, outputStr)
				}
				if !strings.Contains(outputStr, elem.messageColor) {
					t.Errorf("%s: expected message color %q not found", level, elem.messageColor)
				}
				if !strings.Contains(outputStr, elem.reset) {
					t.Errorf("%s: expected message color %q not found", level, elem.reset)
				}
			}
			testLevel("Debug", testDebug, elem.prefixDebug)
			testLevel("Error", testError, elem.prefixError)
			testLevel("Fatal", testFatal, elem.prefixFatal)
			testLevel("Info", testInfo, elem.prefixInfo)
			testLevel("Warn", testWarn, elem.prefixWarn)
		})
	}
}
func TestSetExtractor(t *testing.T) {
	// Дописать
}
func TestSetFormat(t *testing.T) {
	array := []struct {
		name   string
		format TypeFormat
		expect string
	}{
		{"Json", FormatJson, `"message":"test"`},
		{"Text", FormatText, "test"},
	}
	for _, elem := range array {
		t.Run(elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger()
			logger.SetFormat(elem.format)
			logger.SetMode(ModeSync, buf, 0)
			logger.Info("test")
			output := buf.String()
			if !strings.Contains(output, elem.expect) {
				t.Errorf("Expected output to contain %q, got %q", elem.expect, output)
			}
		})
	}
}
func TestSetLevel(t *testing.T) {
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
		t.Run(elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			logger := NewLogger()
			logger.SetLevel(elem.level)
			logger.SetMode(ModeSync, buf, 0)
			elem.logFunc(logger)
			if elem.shouldLog && buf.Len() == 0 {
				t.Error("Expected log to be written, but got nothing")
			}
			if !elem.shouldLog && buf.Len() > 0 {
				t.Error("Expected no log, but got output")
			}
		})
	}
}
func TestSetMode(t *testing.T) {
	t.Run("Async mode", func(t *testing.T) {
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
	t.Run("Sync mode", func(t *testing.T) {
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
func TestSetTheme(t *testing.T) {
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
			name:         "Dark theme",
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
			name:         "Light theme",
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
		t.Run(elem.name, func(t *testing.T) {
			testLevel := func(level string, logFunc func(Logger), expectedPrefix string) {
				t.Helper()
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				defer r.Close()
				logger := NewLogger()
				logger.SetLevel(LevelDebug)
				logger.SetMode(ModeSync, w)
				logger.SetTheme(elem.theme)
				logFunc(logger)
				w.Close()
				output, err := io.ReadAll(r)
				if err != nil {
					t.Fatal(err)
				}
				outputStr := string(output)
				if !strings.Contains(outputStr, elem.callerColor) && level == "DEBUG" {
					t.Errorf("%s: expected prefix %q not found in %q", level, elem.callerColor, outputStr)
				}
				if !strings.Contains(outputStr, expectedPrefix) {
					t.Errorf("%s: expected prefix %q not found in %q", level, expectedPrefix, outputStr)
				}
				if !strings.Contains(outputStr, elem.messageColor) {
					t.Errorf("%s: expected message color %q not found", level, elem.messageColor)
				}
				if !strings.Contains(outputStr, elem.reset) {
					t.Errorf("%s: expected message color %q not found", level, elem.reset)
				}
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
		{"connection refused", []byte("dial: connection refused"), true},
		{"timeout", []byte("i/o timeout"), true},
		{"normal message", []byte("user logged in"), false},
		{"empty", []byte{}, false},
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
func testDebug(l Logger) {
	l.Debug("message")
}
func testDebugWithContext(l Logger) {
	l.DebugWithContext(context.Background(), "message")
}
func testError(l Logger) {
	l.Error("message")
}
func testErrorWithContext(l Logger) {
	l.ErrorWithContext(context.Background(), "message")
}
func testFatal(l Logger) {
	oldExit := osExit
	osExit = func(int) {}
	defer func() { osExit = oldExit }()
	l.Fatal("message")
}
func testFatalWithContext(l Logger) {
	oldExit := osExit
	osExit = func(int) {}
	defer func() { osExit = oldExit }()
	l.FatalWithContext(context.Background(), "message")
}
func testInfo(l Logger) {
	l.Info("message")
}
func testInfoWithContext(l Logger) {
	l.InfoWithContext(context.Background(), "message")
}
func testWarn(l Logger) {
	l.Warn("message")
}
func testWarnWithContext(l Logger) {
	l.WarnWithContext(context.Background(), "message")
}
