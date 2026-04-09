package ulog

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// Публичные функции
func TestField(t *testing.T) {
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
	t.Run("Err", func(t *testing.T) {
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
	t.Run("Errs", func(t *testing.T) {
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
		if f.typeValue != FieldInt {
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

// Вспомогательный тип для теста ошибок
func TestConcurrency(t *testing.T) {
	logger := NewLogger()
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.SetLevel(LevelDebug)
			logger.SetTheme(ThemeDark)
		}()
	}
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			logger.Info("test")
			logger.Debug("debug")
		}()
	}
	wg.Wait()
}
func TestGetLoggerLevel(t *testing.T) {
	tests := []struct {
		env      string
		expected TypeLevel
	}{
		{"debug", LevelDebug},
		{"info", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"fatal", LevelFatal},
		{"", LevelInfo},
	}
	for _, tt := range tests {
		t.Setenv("LOG_LEVEL", tt.env)
		if got := getLoggerLevel(); got != tt.expected {
			t.Errorf("LOG_LEVEL=%s: got %d, want %d", tt.env, got, tt.expected)
		}
	}
}
func TestGetLoggerTheme(t *testing.T) {
	t.Run("Dark theme", func(t *testing.T) {
		t.Setenv("TERM_THEME", "dark")
		theme := getLoggerTheme()
		if !strings.HasPrefix(theme.prefixError, colorDarkRed) {
			t.Error("Light theme prefix should start with red color")
		}
		if !strings.HasPrefix(theme.prefixDebug, colorDarkCyan) {
			t.Error("Light theme debug should start with cyan")
		}
		if !strings.HasPrefix(theme.prefixFatal, colorDarkPurple) {
			t.Error("Light theme fatal should start with purple")
		}
		if !strings.HasPrefix(theme.prefixInfo, colorDarkGreen) {
			t.Error("Light theme info should start with green")
		}
		if !strings.HasPrefix(theme.prefixWarn, colorDarkYellow) {
			t.Error("Light theme warn should start with yellow")
		}
	})
	t.Run("Light theme", func(t *testing.T) {
		t.Setenv("TERM_THEME", "light")
		theme := getLoggerTheme()
		if !strings.HasPrefix(theme.prefixError, colorLightRed) {
			t.Error("Light theme prefix should start with red color")
		}
		if !strings.HasPrefix(theme.prefixDebug, colorLightCyan) {
			t.Error("Light theme debug should start with cyan")
		}
		if !strings.HasPrefix(theme.prefixFatal, colorLightPurple) {
			t.Error("Light theme fatal should start with purple")
		}
		if !strings.HasPrefix(theme.prefixInfo, colorLightGreen) {
			t.Error("Light theme info should start with green")
		}
		if !strings.HasPrefix(theme.prefixWarn, colorLightYellow) {
			t.Error("Light theme warn should start with yellow")
		}
	})
}
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

//func TestSetLevel(t *testing.T) {
//	logger := NewLogger()
//	logger.SetLevel(LevelError)
//	if logger.getLevel() != LevelError {
//		t.Errorf("Expected level %d, got %d", LevelError, logger.getLevel())
//	}
//}

//func TestSetTheme(t *testing.T) {
//	logger := NewLogger(WithTheme(ThemeDark))
//	buf := &bytes.Buffer{}
//	logger.SetMode(ModeSync, buf, 0)
//	logger.Info("test message")
//	output := buf.String()
//	if !strings.Contains(output, "\033[92m") {
//		t.Error("Dark theme colors not found in output")
//	}
//	buf.Reset()
//	logger.SetTheme(ThemeLight)
//	logger.Info("test message")
//	if !strings.Contains(output, "\033[92m") {
//		t.Error("Light theme colors not found")
//	}
//}
