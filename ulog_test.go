// ULOG (Universal Logger Observability Gateway)
// Copyright (C) 2026 Mikhail Dadaev
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program. If not, see <https://www.gnu.org/licenses/>.

package ulog

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Публичные функции
func TestMain(m *testing.M) {
	loadEnv(".env")
	os.Exit(m.Run())
}
func Test_Telemetry(t *testing.T) {
	buf := &bytes.Buffer{}
	telemetry := NewTelemetry(
		WithFormat(FormatText),
		WithMode(ModeSync, buf),
	)
	defer telemetry.Close()
	if telemetry == nil {
		t.Fatal("NewTelemetry returned nil")
	}
	telemetry.Info(DataLog, String("message", "test info text"))
	if !strings.Contains(buf.String(), "test info text") {
		t.Errorf("Expected 'test info text', got %q", buf.String())
	}
}
func Test_Telemetry_Caller(t *testing.T) {
	buf := &bytes.Buffer{}
	telemetry := NewTelemetry(
		WithLevel(LevelDebug),
		WithFormat(FormatJson),
		WithMode(ModeSync, buf),
	)
	defer telemetry.Close()
	line := 0
	_, _, line, _ = runtime.Caller(0)
	telemetry.Debug(DataLog, String("msg", "test"))
	output := buf.String()
	expected := fmt.Sprintf(`"caller":"ulog_test.go:%d"`, line+1)
	if !strings.Contains(output, expected) {
		t.Errorf("expected %s, got %s", expected, output)
	}
}
func Test_Telemetry_Close(t *testing.T) {
	t.Run("Async", func(t *testing.T) {
		buf := &bytes.Buffer{}
		telemetry := NewTelemetry(WithMode(ModeAsync, buf, 100))
		telemetry.Info(DataLog, String("message", "test info text"))
		err := telemetry.Close()
		if err != nil {
			t.Errorf("Close() returned error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "test info text") {
			t.Error("Message not written after Close")
		}
	})
	t.Run("Sync", func(t *testing.T) {
		buf := &bytes.Buffer{}
		telemetry := NewTelemetry(WithMode(ModeSync, buf))
		telemetry.Info(DataLog, String("message", "test info text"))
		err := telemetry.Close()
		if err != nil {
			t.Errorf("Close() returned error: %v", err)
		}
		if buf.Len() == 0 {
			t.Error("Logger stopped working after Close in sync mode")
		}
	})
}
func Test_Telemetry_GetTime(t *testing.T) {
	var buf1 bytes.Buffer
	getTime(&buf1, time.Date(2026, 1, 15, 12, 0, 0, 0, time.FixedZone("CET", 3600)))
	var buf2 bytes.Buffer
	getTime(&buf2, time.Date(2026, 1, 15, 12, 0, 1, 0, time.FixedZone("CEST", 7200)))
	if !strings.Contains(buf1.String(), "+01:00") {
		t.Errorf("first call: expected +01:00, got %s", buf1.String())
	}
	if !strings.Contains(buf2.String(), "+02:00") {
		t.Errorf("second call: expected +02:00 (TZ changed), got %s", buf2.String())
	}
}
func Test_Telemetry_Extractor(t *testing.T) {
	tests := []struct {
		name      string
		keys      []string
		context   context.Context
		want      map[string]string
		shouldAdd bool
	}{
		{
			name:      "NullContext",
			keys:      []string{"test_empty"},
			context:   context.Background(),
			want:      map[string]string{},
			shouldAdd: false,
		},
		{
			name:      "NullKeys",
			keys:      nil,
			context:   context.WithValue(context.Background(), "trace_id", "5B8EFFF7-9803-8103-D269-B633813FC700"),
			want:      map[string]string{},
			shouldAdd: false,
		},
		{
			name:      "Bool",
			keys:      []string{"test_bool"},
			context:   context.WithValue(context.Background(), "test_bool", true),
			want:      map[string]string{"test_bool": "true"},
			shouldAdd: true,
		},
		{
			name:      "Duration",
			keys:      []string{"test_duration"},
			context:   context.WithValue(context.Background(), "test_duration", 5*time.Second),
			want:      map[string]string{"test_duration": "5s"},
			shouldAdd: true,
		},
		{
			name:      "Float64",
			keys:      []string{"test_float64"},
			context:   context.WithValue(context.Background(), "test_float64", 3.14159),
			want:      map[string]string{"test_float64": "3.14159"},
			shouldAdd: true,
		},
		{
			name:      "Int",
			keys:      []string{"test_int"},
			context:   context.WithValue(context.Background(), "test_int", int(12345)),
			want:      map[string]string{"test_int": "12345"},
			shouldAdd: true,
		},
		{
			name:      "Int64",
			keys:      []string{"test_int64"},
			context:   context.WithValue(context.Background(), "test_int64", int64(12345)),
			want:      map[string]string{"test_int64": "12345"},
			shouldAdd: true,
		},
		{
			name:      "String",
			keys:      []string{"test_string"},
			context:   context.WithValue(context.Background(), "test_string", "abc-123"),
			want:      map[string]string{"test_string": "abc-123"},
			shouldAdd: true,
		},
		{
			name:      "Time",
			keys:      []string{"test_time"},
			context:   context.WithValue(context.Background(), "test_time", time.Date(2026, 4, 10, 12, 0, 0, 0, time.UTC)),
			want:      map[string]string{"test_time": "2026-04-10T12:00:00.000000+00:00"},
			shouldAdd: true,
		},
		{
			name:      "Multiple",
			keys:      []string{"text", "user_id"},
			context:   context.WithValue(context.WithValue(context.Background(), "text", "test"), "user_id", int64(12345)),
			want:      map[string]string{"text": "test", "user_id": "12345"},
			shouldAdd: true,
		},
	}
	for _, elem := range tests {
		t.Run("WithExtractor/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry(
				WithExtractor(elem.keys...),
				WithFormat(FormatJson),
				WithMode(ModeSync, buf),
			)
			defer telemetry.Close()
			telemetry.InfoWithContext(elem.context, DataLog, String("message", "test info text"))
			output := buf.String()
			checkExtractor(t, elem, output)
		})
		t.Run("SetExtractor/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry()
			defer telemetry.Close()
			telemetry.SetExtractor(elem.keys...)
			telemetry.SetFormat(FormatJson)
			telemetry.SetMode(ModeSync, buf)
			telemetry.InfoWithContext(elem.context, DataLog, String("message", "test info text"))
			output := buf.String()
			checkExtractor(t, elem, output)
		})
	}
}
func Test_Telemetry_Field(t *testing.T) {
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
func Test_Telemetry_Format(t *testing.T) {
	array := []struct {
		name   string
		format TypeFormat
		expect string
	}{
		{"Json", FormatJson, `"message":"test info text"`},
		{"Text", FormatText, `message="test info text"`},
	}
	for _, elem := range array {
		t.Run("WithFormat/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry(
				WithFormat(elem.format),
				WithMode(ModeSync, buf, 0),
			)
			defer telemetry.Close()
			telemetry.Info(DataLog, String("message", "test info text"))
			output := buf.String()
			if !strings.Contains(output, elem.expect) {
				t.Errorf("Expected output to contain %q, got %q", elem.expect, output)
			}
		})
		t.Run("SetFormat/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry()
			defer telemetry.Close()
			telemetry.SetFormat(elem.format)
			telemetry.SetMode(ModeSync, buf, 0)
			telemetry.Info(DataLog, String("message", "test info text"))
			output := buf.String()
			if !strings.Contains(output, elem.expect) {
				t.Errorf("Expected output to contain %q, got %q", elem.expect, output)
			}
		})
	}
}
func Test_Telemetry_Level(t *testing.T) {
	array := []struct {
		name         string
		level        TypeLevel
		functionTest func(Telemetry)
		responseBool bool
	}{
		{"Debug->Debug", LevelDebug, testDebug, true},
		{"Debug->Info", LevelDebug, testInfo, true},
		{"Debug->Warn", LevelDebug, testWarn, true},
		{"Debug->Error", LevelDebug, testError, true},
		{"Debug->Fatal", LevelDebug, testFatal, true},
		{"Error->Debug", LevelError, testDebug, false},
		{"Error->Info", LevelError, testInfo, false},
		{"Error->Warn", LevelError, testWarn, false},
		{"Error->Error", LevelError, testError, true},
		{"Error->Fatal", LevelError, testFatal, true},
		{"Fatal->Debug", LevelFatal, testDebug, false},
		{"Fatal->Info", LevelFatal, testInfo, false},
		{"Fatal->Warn", LevelFatal, testWarn, false},
		{"Fatal->Error", LevelFatal, testError, false},
		{"Fatal->Fatal", LevelFatal, testFatal, true},
		{"Info->Debug", LevelInfo, testDebug, false},
		{"Info->Info", LevelInfo, testInfo, true},
		{"Info->Warn", LevelInfo, testWarn, true},
		{"Info->Error", LevelInfo, testError, true},
		{"Info->Fatal", LevelInfo, testFatal, true},
		{"Warn->Debug", LevelWarn, testDebug, false},
		{"Warn->Info", LevelWarn, testInfo, false},
		{"Warn->Warn", LevelWarn, testWarn, true},
		{"Warn->Error", LevelWarn, testError, true},
		{"Warn->Fatal", LevelWarn, testFatal, true},
	}
	for _, elem := range array {
		t.Run("WithLevel/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry(
				WithLevel(elem.level),
				WithMode(ModeSync, buf, 0),
			)
			defer telemetry.Close()
			elem.functionTest(telemetry)
			if elem.responseBool && buf.Len() == 0 {
				t.Error("Expected log to be written, but got nothing")
			}
			if !elem.responseBool && buf.Len() > 0 {
				t.Error("Expected no log, but got output")
			}
		})
		t.Run("SetLevel/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry()
			telemetry.SetLevel(elem.level)
			telemetry.SetMode(ModeSync, buf, 0)
			elem.functionTest(telemetry)
			if elem.responseBool && buf.Len() == 0 {
				t.Error("Expected log to be written, but got nothing")
			}
			if !elem.responseBool && buf.Len() > 0 {
				t.Error("Expected no log, but got output")
			}
		})
	}
}
func Test_Telemetry_Method(t *testing.T) {
	array := []struct {
		name         string
		functionTest func(Telemetry)
		level        TypeLevel
		responseBool bool
	}{
		{"Debug", testDebug, LevelDebug, true},
		{"DebugWithContext", testDebugWithContext, LevelDebug, true},
		{"Error", testError, LevelError, true},
		{"ErrorWithContext", testErrorWithContext, LevelError, true},
		{"Fatal", testFatal, LevelFatal, true},
		{"FatalWithContext", testFatalWithContext, LevelFatal, true},
		{"Info", testInfo, LevelInfo, true},
		{"InfoWithContext", testInfoWithContext, LevelInfo, true},
		{"Warn", testWarn, LevelWarn, true},
		{"WarnWithContext", testWarnWithContext, LevelWarn, true},
	}
	for _, elem := range array {
		t.Run(elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry(
				WithMode(ModeSync, buf),
				WithLevel(elem.level),
			)
			defer telemetry.Close()
			elem.functionTest(telemetry)
			output := buf.String()
			if elem.responseBool && !strings.Contains(output, "message") {
				t.Errorf("Expected message not found in output: %q", output)
			}
		})
	}
}
func Test_Telemetry_Mode(t *testing.T) {
	t.Run("WithMode/Async", func(t *testing.T) {
		writerBuf := &bytes.Buffer{}
		telemetry := NewTelemetry(
			WithMode(ModeAsync, writerBuf, 1000),
		)
		telemetry.Info(DataLog, String("message", "test info text"))
		telemetry.Close()
		if writerBuf.Len() == 0 {
			t.Error("Async mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test info text") {
			t.Error("Async mode: expected message not found")
		}
	})
	t.Run("SetMode/Async", func(t *testing.T) {
		writerBuf := &bytes.Buffer{}
		telemetry := NewTelemetry()
		telemetry.SetMode(ModeAsync, writerBuf, 1000)
		telemetry.Info(DataLog, String("message", "test info text"))
		telemetry.Close()
		if writerBuf.Len() == 0 {
			t.Error("Async mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test info text") {
			t.Error("Async mode: expected message not found")
		}
	})
	t.Run("WithMode/Sync", func(t *testing.T) {
		writerBuf := &bytes.Buffer{}
		telemetry := NewTelemetry(
			WithMode(ModeSync, writerBuf),
		)
		telemetry.Info(DataLog, String("message", "test info text"))
		telemetry.Close()
		if writerBuf.Len() == 0 {
			t.Error("Sync mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test info text") {
			t.Error("Sync mode: expected message not found")
		}
	})
	t.Run("SetMode/Sync", func(t *testing.T) {
		writerBuf := &bytes.Buffer{}
		telemetry := NewTelemetry()
		telemetry.SetMode(ModeSync, writerBuf)
		telemetry.Info(DataLog, String("message", "test info text"))
		telemetry.Close()
		if writerBuf.Len() == 0 {
			t.Error("Sync mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test info text") {
			t.Error("Sync mode: expected message not found")
		}
	})
}
func Test_Telemetry_Theme(t *testing.T) {
	array := []struct {
		name        string
		theme       TypeTheme
		callerColor string
		dataColor   string
		prefixDebug string
		prefixError string
		prefixFatal string
		prefixInfo  string
		prefixWarn  string
		reset       string
	}{
		{
			name:        "Dark",
			theme:       ThemeDark,
			callerColor: colorDarkBlue,
			dataColor:   colorDarkWhite,
			prefixDebug: colorDarkCyan + "[DEBUG]",
			prefixError: colorDarkRed + "[ERROR]",
			prefixFatal: colorDarkPurple + "[FATAL]",
			prefixInfo:  colorDarkGreen + "[INFO]",
			prefixWarn:  colorDarkYellow + "[WARN]",
			reset:       colorReset,
		},
		{
			name:        "Light",
			theme:       ThemeLight,
			callerColor: colorLightBlue,
			dataColor:   colorLightBlack,
			prefixDebug: colorLightCyan + "[DEBUG]",
			prefixError: colorLightRed + "[ERROR]",
			prefixFatal: colorLightPurple + "[FATAL]",
			prefixInfo:  colorLightGreen + "[INFO]",
			prefixWarn:  colorLightYellow + "[WARN]",
			reset:       colorReset,
		},
	}
	for _, elem := range array {
		t.Run("WithTheme/"+elem.name, func(t *testing.T) {
			testLevel := func(level string, functionTest func(Telemetry), expectedPrefix string) {
				buf := &bytes.Buffer{}
				telemetry := NewTelemetry(
					WithFormat(FormatText),
					WithLevel(LevelDebug),
					WithMode(ModeSync, buf),
					WithTheme(elem.theme),
				)
				defer telemetry.Close()
				functionTest(telemetry)
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
			testLevel := func(level string, functionTest func(Telemetry), expectedPrefix string) {
				buf := &bytes.Buffer{}
				telemetry := NewTelemetry()
				defer telemetry.Close()
				telemetry.SetFormat(FormatText)
				telemetry.SetLevel(LevelDebug)
				telemetry.SetMode(ModeSync, buf)
				telemetry.SetTheme(elem.theme)
				functionTest(telemetry)
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
func Test_TelemetryLog(t *testing.T) {
	buf := &bytes.Buffer{}
	telemetry := NewTelemetry(
		WithFormat(FormatText),
		WithMode(ModeSync, buf),
	)
	defer telemetry.Close()
	telemetryLog := NewTelemetryLog(LevelInfo, telemetry)
	telemetryLog.Print("test text")
	if !strings.Contains(buf.String(), "test text") {
		t.Errorf("Expected 'test text', got %q", buf.String())
	}
}
func Test_TelemetryLog_Ignore(t *testing.T) {
	array := []struct {
		name     string
		message  string
		expected bool
	}{
		{"EOF", "read: EOF", true},
		{"TLS handshake", "TLS handshake error", true},
		{"Connection refused", "dial: connection refused", true},
		{"Timeout", "i/o timeout", true},
		{"Broken pipe", "broken pipe", true},
		{"Empty", "", true},
		{"Normal message", "user logged in", false},
	}
	for _, elem := range array {
		t.Run(elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry(
				WithFormat(FormatText),
				WithMode(ModeSync, buf),
			)
			defer telemetry.Close()
			telemetryLog := NewTelemetryLog(LevelError, telemetry)
			telemetryLog.Print(elem.message)
			output := buf.String()
			if elem.expected {
				if output != "" {
					t.Errorf("Expected log to be ignored, but got output: %q", output)
				}
			} else {
				if output == "" {
					t.Error("Expected log to be written, but got nothing")
				}
				if !strings.Contains(output, elem.message) {
					t.Errorf("Expected message %q not found in output: %q", elem.message, output)
				}
			}
		})
	}
}
func Test_Sink(t *testing.T) {
	buf1 := &bytes.Buffer{}
	buf2 := &bytes.Buffer{}
	buf3 := &bytes.Buffer{}
	buf4 := &bytes.Buffer{}
	data := `{"message":"test"}`
	fields := []Field{
		String("message", "test"),
	}
	tee := NewTeeSink(buf1, buf2)
	if tee.Len() != 2 {
		t.Errorf("Len() = %d, want 2", tee.Len())
	}
	tee.Write([]byte(data))
	if buf1.String() != data {
		t.Errorf("buf1: expected %q, got %q", data, buf1.String())
	}
	if buf2.String() != data {
		t.Errorf("buf2: expected %q, got %q", data, buf2.String())
	}
	buf1.Reset()
	buf2.Reset()
	tee.Add(buf3)
	if tee.Len() != 3 {
		t.Errorf("After Add, Len() = %d, want 3", tee.Len())
	}
	tee.Remove(1)
	if tee.Len() != 2 {
		t.Errorf("After Remove, Len() = %d, want 2", tee.Len())
	}
	tee.Replace(0, buf4)
	if tee.Len() != 2 {
		t.Errorf("After Replace, Len() = %d, want 2", tee.Len())
	}
	attributes := writeAttributes{
		typeData:   DataLog,
		typeFormat: FormatJson,
		typeLevel:  LevelInfo,
	}
	tee.WriteWithAttributes(attributes, fields)
	if buf3.String() == "" {
		t.Error("buf3: expected content, got empty")
	}
	if buf4.String() == "" {
		t.Error("buf4: expected content, got empty")
	}
	err := tee.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
	}
}
func Test_Sink_MixedWriters(t *testing.T) {
	var (
		mutex        sync.Mutex
		httpRequests int
		httpBody     []byte
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		mutex.Lock()
		httpRequests++
		httpBody = body
		mutex.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkHttp := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpDisabledCircuit(),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sinkHttp.Close()
	var buf bytes.Buffer
	tee := NewTeeSink(sinkHttp, &buf)
	defer tee.Close()
	attrs := writeAttributes{
		typeData:   DataLog,
		typeFormat: FormatJson,
		typeLevel:  LevelInfo,
	}
	fields := []Field{
		String("message", "test"),
		Int("count", 42),
	}
	n, err := tee.WriteWithAttributes(attrs, fields)
	if err != nil {
		t.Fatalf("WriteWithAttributes failed: %v", err)
	}
	if n == 0 {
		t.Error("expected non-zero bytes written")
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mutex.Lock()
		count := httpRequests
		mutex.Unlock()
		if count >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	mutex.Lock()
	reqCount := httpRequests
	capturedBody := httpBody
	mutex.Unlock()
	if reqCount != 1 {
		t.Errorf("SinkWriter: expected 1 HTTP request, got %d", reqCount)
	}
	var httpJSON map[string]any
	if err := json.Unmarshal(capturedBody, &httpJSON); err != nil {
		t.Errorf("SinkWriter: invalid JSON: %v", err)
	}
	if httpJSON["message"] != "test" {
		t.Errorf("SinkWriter: expected message 'test', got %v", httpJSON["message"])
	}
	if httpJSON["count"] != float64(42) {
		t.Errorf("SinkWriter: expected count 42, got %v", httpJSON["count"])
	}
	output := buf.String()
	if output == "" {
		t.Fatal("io.Writer: expected content, got empty")
	}
	var bufJSON map[string]any
	if err := json.Unmarshal([]byte(output), &bufJSON); err != nil {
		t.Errorf("io.Writer: invalid JSON: %v", err)
	}
	if bufJSON["message"] != "test" {
		t.Errorf("io.Writer: expected message 'test', got %v", bufJSON["message"])
	}
	if bufJSON["count"] != float64(42) {
		t.Errorf("io.Writer: expected count 42, got %v", bufJSON["count"])
	}
	if bufJSON["type"] != "log" {
		t.Errorf("io.Writer: expected type 'log', got %v", bufJSON["type"])
	}
	if bufJSON["level"] != "info" {
		t.Errorf("io.Writer: expected level 'info', got %v", bufJSON["level"])
	}
}
func Test_SinkFactory_Discord(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		var data DiscordData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Errorf("failed to decode JSON: %v", err)
		}
		if data.UserName != "ULog Bot" {
			t.Errorf("expected username 'ULog Bot', got '%s'", data.UserName)
		}
		if data.Content != "test" {
			t.Errorf("expected content 'test', got '%s'", data.Content)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	sinkDiscord := NewSinkDiscord(server.URL, "ULog Bot", "")
	fields := []Field{
		String("message", "test"),
	}
	_, err := sinkDiscord.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkDiscord.Close()
}
func Test_SinkFactory_Kafka(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/vnd.kafka.json.v2+json" {
			t.Errorf("expected Content-Type 'application/vnd.kafka.json.v2+json', got '%s'", r.Header.Get("Content-Type"))
		}
		var records struct {
			Records []KafkaData `json:"records"`
		}
		json.NewDecoder(r.Body).Decode(&records)
		if len(records.Records) != 1 {
			t.Fatal("expected 1 record")
		}
		record := records.Records[0]
		var value map[string]interface{}
		json.Unmarshal(record.Value, &value)
		expectedFields := map[string]interface{}{
			"service":  "test-service",
			"message":  "test",
			"trace_id": "5b8efff798038103d269b633813fc700",
			"node_id":  "node-01",
			"count":    float64(42),
			"duration": "5s",
			"_level":   "ERROR",
			"_type":    "LOG",
		}
		for key, want := range expectedFields {
			if got := value[key]; got != want {
				t.Errorf("%s: expected '%v', got '%v'", key, want, got)
			}
		}
		if record.Key != "5b8efff798038103d269b633813fc700" {
			t.Errorf("key: expected '5b8efff798038103d269b633813fc700', got '%s'", record.Key)
		}
		if record.Timestamp.IsZero() {
			t.Error("timestamp is zero")
		}
		if _, ok := value["_timestamp"]; !ok {
			t.Error("_timestamp missing")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkKafka := NewSinkKafka(server.URL + "/topics/test-topic")
	fields := []Field{
		String("service", "test-service"),
		String("message", "test"),
		String("trace_id", "5B8EFFF7-9803-8103-D269-B633813FC700"),
		String("node_id", "node-01"),
		Int("count", 42),
		Duration("duration", 5*time.Second),
	}
	_, err := sinkKafka.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkKafka.Close()
}
func Test_SinkFactory_Loki(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		if r.Method != http.MethodPost {
			t.Errorf("wrong method: %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("wrong Content-Type: %s", r.Header.Get("Content-Type"))
		}
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read body: %v", err)
			return
		}
		t.Logf("Body (raw): %s", string(rawBody))
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, rawBody, "", "  "); err == nil {
			t.Logf("Body (pretty):\n%s", prettyJSON.String())
		}
		var data LokiData
		if err := json.Unmarshal(rawBody, &data); err != nil {
			t.Errorf("failed to decode JSON: %v", err)
			return
		}
		resource := data.ResourceLogs[0].Resource
		if len(resource.Attributes) != 3 {
			t.Errorf("expected 3 resource attributes, got %d", len(resource.Attributes))
			return
		}
		expectedResourceAttrs := map[string]string{
			"service.name":                "test-service",
			"service.namespace":           "payments",
			"deployment.environment.name": "production",
		}
		foundAttrs := make(map[string]string)
		for _, a := range resource.Attributes {
			foundAttrs[a.Key] = a.Value.StringValue
		}
		for key, want := range expectedResourceAttrs {
			if got, ok := foundAttrs[key]; !ok {
				t.Errorf("resource attribute %q not found", key)
			} else if got != want {
				t.Errorf("resource attribute %q: expected %q, got %q", key, want, got)
			}
		}
		lr := data.ResourceLogs[0].ScopeLogs[0].LogRecords[0]
		if lr.Body.StringValue == nil || *lr.Body.StringValue != "test" {
			t.Errorf("wrong message: %v", lr.Body.StringValue)
		}
		if lr.SeverityText != "ERROR" {
			t.Errorf("wrong severity: %s", lr.SeverityText)
		}
		for _, a := range lr.Attributes {
			if a.Key == "service" || a.Key == "namespace" || a.Key == "environment" {
				t.Errorf("%s should NOT be duplicated in LogRecord.Attributes", a.Key)
			}
		}
		foundUserID := false
		foundTraceID := false
		for _, a := range lr.Attributes {
			if a.Key == "user_id" && a.Value.StringValue == "019687278c7e800087cbbdba4f634d9f" {
				foundUserID = true
			}
			if a.Key == "trace_id" && a.Value.StringValue == "5b8efff798038103d269b633813fc700" {
				foundTraceID = true
			}
		}
		if !foundUserID {
			t.Error("user_id attribute not found")
		}
		if !foundTraceID {
			t.Error("trace_id attribute not found")
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"partialSuccess":{}}`))
	}))
	defer server.Close()
	sinkLoki := NewSinkLoki(
		server.URL,
		WithHttpDisabledBatch(),
	)
	defer sinkLoki.Close()
	fields := []Field{
		String("service", "test-service"),
		String("namespace", "payments"),
		String("environment", "production"),
		String("message", "test"),
		String("user_id", "019687278c7e800087cbbdba4f634d9f"),
		String("trace_id", "5B8EFFF7-9803-8103-D269-B633813FC700"),
	}
	_, err := sinkLoki.WriteWithAttributes(
		writeAttributes{typeData: DataLog, typeLevel: LevelError},
		fields,
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	wg.Wait()
}
func Test_SinkFactory_LokiCloud(t *testing.T) {
	token := os.Getenv("GRAFANA_CLOUD_TOKEN")
	if token == "" {
		t.Skip("GRAFANA_CLOUD_TOKEN not set — skipping integration test")
	}
	sinkLoki := NewSinkLoki(
		"https://otlp-gateway-prod-eu-north-0.grafana.net/otlp/v1/logs",
		WithHttpHeader(
			"Authorization",
			"Basic "+token,
		),
	)
	defer sinkLoki.Close()
	telemetry := NewTelemetry(
		WithMode(ModeSync, sinkLoki),
		WithFormat(FormatJson),
	)
	defer telemetry.Close()
	telemetry.Error(DataLog,
		String("service", "test-service"),
		String("message", "test"),
		String("user_id", "user-12345"),
		String("trace_id", "5B8EFFF7-9803-8103-D269-B633813FC700"),
	)
}
func Test_SinkFactory_Prometheus(t *testing.T) {
	tests := []struct {
		name         string
		metricName   string
		expectedTemp int
		expectedMono bool
	}{
		{
			name:         "counter",
			metricName:   "http_requests_total",
			expectedTemp: 2,
			expectedMono: true,
		},
		{
			name:         "histogram",
			metricName:   "http_request_duration_seconds",
			expectedTemp: 2,
		},
		{
			name:         "gauge",
			metricName:   "cpu_usage",
			expectedTemp: 0,
			expectedMono: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			wg.Add(1)
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				defer wg.Done()
				if r.Method != http.MethodPost {
					t.Errorf("wrong method: %s", r.Method)
				}
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("wrong Content-Type: %s", r.Header.Get("Content-Type"))
				}
				rawBody, err := io.ReadAll(r.Body)
				if err != nil {
					t.Errorf("failed to read body: %v", err)
					return
				}
				t.Logf("Body (raw): %s", string(rawBody))
				var prettyJSON bytes.Buffer
				if err := json.Indent(&prettyJSON, rawBody, "", "  "); err == nil {
					t.Logf("Body (pretty):\n%s", prettyJSON.String())
				}
				var data PrometheusData
				if err := json.Unmarshal(rawBody, &data); err != nil {
					t.Errorf("failed to decode JSON: %v", err)
					return
				}
				if len(data.ResourceMetrics) != 1 {
					t.Errorf("expected 1 resourceMetrics, got %d", len(data.ResourceMetrics))
					return
				}
				rm := data.ResourceMetrics[0]
				if len(rm.Resource.Attributes) != 3 {
					t.Errorf("expected 3 resource attributes, got %d", len(rm.Resource.Attributes))
					return
				}
				expectedResourceAttrs := map[string]string{
					"service.name":                "test-service",
					"service.namespace":           "payments",
					"deployment.environment.name": "production",
				}
				foundAttrs := make(map[string]string)
				for _, attr := range rm.Resource.Attributes {
					foundAttrs[attr.Key] = attr.Value.StringValue
				}
				for key, want := range expectedResourceAttrs {
					if got, ok := foundAttrs[key]; !ok {
						t.Errorf("resource attribute %q not found", key)
					} else if got != want {
						t.Errorf("resource attribute %q: expected %q, got %q", key, want, got)
					}
				}
				if len(rm.ScopeMetrics) != 1 {
					t.Errorf("expected 1 scopeMetrics, got %d", len(rm.ScopeMetrics))
					return
				}
				sm := rm.ScopeMetrics[0]
				if sm.Scope.Name != "ulog" {
					t.Errorf("expected scope.name=ulog, got %s", sm.Scope.Name)
				}
				if sm.Scope.Version != Version {
					t.Errorf("expected scope.version=%s, got %s", Version, sm.Scope.Version)
				}
				if len(sm.Metrics) != 1 {
					t.Errorf("expected 1 metric, got %d", len(sm.Metrics))
					return
				}
				metric := sm.Metrics[0]
				if metric.Name != tt.metricName {
					t.Errorf("wrong metric name: expected %q, got %q", tt.metricName, metric.Name)
				}
				switch tt.name {
				case "counter":
					if metric.Sum == nil {
						t.Fatalf("expected Sum (Counter), got nil")
					}
					if metric.Gauge != nil {
						t.Errorf("Counter should not have Gauge")
					}
					if metric.Histogram != nil {
						t.Errorf("Counter should not have Histogram")
					}
					if metric.Sum.AggregationTemporality != tt.expectedTemp {
						t.Errorf("expected temporality %d, got %d",
							tt.expectedTemp, metric.Sum.AggregationTemporality)
					}
					if metric.Sum.IsMonotonic != tt.expectedMono {
						t.Errorf("expected monotonic %v, got %v",
							tt.expectedMono, metric.Sum.IsMonotonic)
					}
					if len(metric.Sum.DataPoints) != 1 {
						t.Fatalf("expected 1 dataPoint, got %d", len(metric.Sum.DataPoints))
					}
					dp := metric.Sum.DataPoints[0]
					if dp.AsDouble != 42.0 {
						t.Errorf("wrong value: %f", dp.AsDouble)
					}
					if dp.TimeUnixNano == "" {
						t.Error("timeUnixNano is empty")
					}
					for _, a := range dp.Attributes {
						if a.Key == "service" || a.Key == "namespace" || a.Key == "environment" {
							t.Errorf("%s should NOT be duplicated in DataPoint.Attributes", a.Key)
						}
					}
				case "gauge":
					if metric.Gauge == nil {
						t.Fatalf("expected Gauge, got nil")
					}
					if metric.Sum != nil {
						t.Errorf("Gauge should not have Sum")
					}
					if metric.Histogram != nil {
						t.Errorf("Gauge should not have Histogram")
					}
					if len(metric.Gauge.DataPoints) != 1 {
						t.Fatalf("expected 1 dataPoint, got %d", len(metric.Gauge.DataPoints))
					}
					dp := metric.Gauge.DataPoints[0]
					if dp.AsDouble != 42.0 {
						t.Errorf("wrong value: %f", dp.AsDouble)
					}
					if dp.TimeUnixNano == "" {
						t.Error("timeUnixNano is empty")
					}
					for _, a := range dp.Attributes {
						if a.Key == "service" || a.Key == "namespace" || a.Key == "environment" {
							t.Errorf("%s should NOT be duplicated in DataPoint.Attributes", a.Key)
						}
					}
				case "histogram":
					if metric.Histogram == nil {
						t.Fatalf("expected Histogram, got nil")
					}
					if metric.Sum != nil {
						t.Errorf("Histogram should not have Sum")
					}
					if metric.Gauge != nil {
						t.Errorf("Histogram should not have Gauge")
					}
					if metric.Histogram.AggregationTemporality != tt.expectedTemp {
						t.Errorf("expected temporality %d, got %d", tt.expectedTemp, metric.Histogram.AggregationTemporality)
					}
					if len(metric.Histogram.DataPoints) != 1 {
						t.Fatalf("expected 1 dataPoint, got %d", len(metric.Histogram.DataPoints))
					}
					dp := metric.Histogram.DataPoints[0]
					if dp.Count != 150 {
						t.Errorf("expected count 150, got %d", dp.Count)
					}
					if dp.Sum == nil || *dp.Sum != 12.5 {
						t.Errorf("expected sum 12.5, got %v", dp.Sum)
					}
					if len(dp.BucketCounts) != 4 {
						t.Errorf("expected 4 bucket counts, got %d", len(dp.BucketCounts))
					}
					if len(dp.ExplicitBounds) != 4 {
						t.Errorf("expected 4 explicit bounds, got %d", len(dp.ExplicitBounds))
					}
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"partialSuccess":{}}`))
			}))
			defer server.Close()
			sinkPrometheus := NewSinkPrometheus(
				server.URL,
				WithHttpDisabledBatch(),
			)
			defer sinkPrometheus.Close()
			fields := []Field{
				String("service", "test-service"),
				String("namespace", "payments"),
				String("environment", "production"),
				String("name", tt.metricName),
			}
			switch tt.name {
			case "counter":
				fields = append(fields, String("type", "counter"), Float64("value", 42.0))
			case "gauge":
				fields = append(fields, String("type", "gauge"), Float64("value", 42.0))
			case "histogram":
				fields = append(fields,
					String("type", "histogram"),
					Int64("count", 150),
					Float64("sum", 12.5),
					Ints64("bucket_counts", []int64{10, 40, 70, 30}),
					Floats64("explicit_bounds", []float64{0.1, 0.5, 1.0, 5.0}),
				)
			}
			_, err := sinkPrometheus.WriteWithAttributes(
				writeAttributes{typeData: DataMetric, typeLevel: LevelError},
				fields,
			)
			if err != nil {
				t.Fatalf("error: %v", err)
			}
			wg.Wait()
		})
	}
}
func Test_SinkFactory_PrometheusCloud(t *testing.T) {
	token := os.Getenv("GRAFANA_CLOUD_TOKEN")
	if token == "" {
		t.Skip("GRAFANA_CLOUD_TOKEN not set — skipping integration test")
	}
	t.Run("Counter", func(t *testing.T) {
		sinkPrometheus := NewSinkPrometheus(
			"https://otlp-gateway-prod-eu-north-0.grafana.net/otlp/v1/metrics",
			WithHttpHeader("Authorization", "Basic "+token),
		)
		defer sinkPrometheus.Close()
		telemetry := NewTelemetry(
			WithMode(ModeSync, sinkPrometheus),
			WithFormat(FormatJson),
		)
		defer telemetry.Close()
		telemetry.Error(DataMetric,
			String("service", "test-service"),
			String("namespace", "payments"),
			String("environment", "production"),
			String("name", "http_requests_total"),
			String("type", "counter"),
			Float64("value", 42.0),
		)
	})
	t.Run("Gauge", func(t *testing.T) {
		sinkPrometheus := NewSinkPrometheus(
			"https://otlp-gateway-prod-eu-north-0.grafana.net/otlp/v1/metrics",
			WithHttpHeader("Authorization", "Basic "+token),
		)
		defer sinkPrometheus.Close()
		telemetry := NewTelemetry(
			WithMode(ModeSync, sinkPrometheus),
			WithFormat(FormatJson),
		)
		defer telemetry.Close()
		telemetry.Error(DataMetric,
			String("service", "test-service"),
			String("namespace", "payments"),
			String("environment", "production"),
			String("name", "cpu_usage_ratio"),
			String("type", "gauge"),
			Float64("value", 0.75),
		)
	})
	t.Run("Histogram", func(t *testing.T) {
		sinkPrometheus := NewSinkPrometheus(
			"https://otlp-gateway-prod-eu-north-0.grafana.net/otlp/v1/metrics",
			WithHttpHeader("Authorization", "Basic "+token),
		)
		defer sinkPrometheus.Close()
		telemetry := NewTelemetry(
			WithMode(ModeSync, sinkPrometheus),
			WithFormat(FormatJson),
		)
		defer telemetry.Close()
		telemetry.Error(DataMetric,
			String("service", "test-service"),
			String("namespace", "payments"),
			String("environment", "production"),
			String("name", "http_request_duration_seconds"),
			String("type", "histogram"),
			Int64("count", 150),
			Float64("sum", 12.5),
			Ints64("bucket_counts", []int64{10, 40, 70, 30}),
			Floats64("explicit_bounds", []float64{0.1, 0.5, 1.0, 5.0}),
		)
	})
}
func Test_SinkFactory_Slack(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var data SlackData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Errorf("failed to decode JSON: %v", err)
		}
		if data.Channel != "#alerts" {
			t.Errorf("expected channel '#alerts', got '%s'", data.Channel)
		}
		if data.UserName != "ULog" {
			t.Errorf("expected username 'ULog', got '%s'", data.UserName)
		}
		if data.Text != "test" {
			t.Errorf("expected text 'test', got '%s'", data.Text)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkSlack := NewSinkSlack(server.URL, "ULog", ":robot:", "", "#alerts")
	fields := []Field{
		String("message", "test"),
	}
	_, err := sinkSlack.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkSlack.Close()
}
func Test_SinkFactory_Telegram(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		var data TelegramData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Errorf("failed to decode JSON: %v", err)
		}
		if data.ChatID != "chat-123" {
			t.Errorf("expected chat_id 'chat-123', got '%s'", data.ChatID)
		}
		if data.Text != "test" {
			t.Errorf("expected text 'test', got '%s'", data.Text)
		}
		if data.ParseMode != "HTML" {
			t.Errorf("expected parse_mode 'HTML', got '%s'", data.ParseMode)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkTelegram := NewSinkTelegram(server.URL, "chat-123")
	fields := []Field{
		String("message", "test"),
	}
	_, err := sinkTelegram.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkTelegram.Close()
}
func Test_SinkFactory_Tempo(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(1)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer wg.Done()
		if r.Method != http.MethodPost {
			t.Errorf("wrong method: %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("wrong Content-Type: %s", r.Header.Get("Content-Type"))
		}
		rawBody, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read body: %v", err)
			return
		}
		t.Logf("Body (raw): %s", string(rawBody))
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, rawBody, "", "  "); err == nil {
			t.Logf("Body (pretty):\n%s", prettyJSON.String())
		}
		var data TempoData
		if err := json.Unmarshal(rawBody, &data); err != nil {
			t.Errorf("failed to decode JSON: %v", err)
			return
		}
		resource := data.ResourceSpans[0].Resource
		if len(resource.Attributes) != 3 {
			t.Errorf("expected 3 resource attributes, got %d", len(resource.Attributes))
			return
		}
		expectedResourceAttrs := map[string]string{
			"service.name":                "test-service",
			"service.namespace":           "payments",
			"deployment.environment.name": "production",
		}
		foundAttrs := make(map[string]string)
		for _, a := range resource.Attributes {
			foundAttrs[a.Key] = a.Value.StringValue
		}
		for key, want := range expectedResourceAttrs {
			if got, ok := foundAttrs[key]; !ok {
				t.Errorf("resource attribute %q not found", key)
			} else if got != want {
				t.Errorf("resource attribute %q: expected %q, got %q", key, want, got)
			}
		}
		span := data.ResourceSpans[0].ScopeSpans[0].Spans[0]
		if span.Name != "test" {
			t.Errorf("wrong name: %s", span.Name)
		}
		if span.SpanID != "eee19b7ec3c1b100" {
			t.Errorf("wrong span_id: %s", span.SpanID)
		}
		if span.TraceID != "5b8efff798038103d269b633813fc700" {
			t.Errorf("wrong trace_id: %s", span.TraceID)
		}
		for _, a := range span.Attributes {
			if a.Key == "service" || a.Key == "namespace" || a.Key == "environment" {
				t.Errorf("%s should NOT be duplicated in Span.Attributes", a.Key)
			}
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkTempo := NewSinkTempo(
		server.URL,
		WithHttpDisabledBatch(),
	)
	defer sinkTempo.Close()
	fields := []Field{
		String("service", "test-service"),
		String("namespace", "payments"),
		String("environment", "production"),
		String("trace_id", "5B8EFFF7-9803-8103-D269-B633813FC700"),
		String("span_id", "EEE19B7E-C3C1-B100"),
		String("name", "test"),
		Int64("duration", 100),
	}
	_, err := sinkTempo.WriteWithAttributes(
		writeAttributes{typeData: DataTrace, typeLevel: LevelError},
		fields,
	)
	if err != nil {
		t.Fatalf("error: %v", err)
	}
	wg.Wait()
}
func Test_SinkFactory_TempoCloud(t *testing.T) {
	token := os.Getenv("GRAFANA_CLOUD_TOKEN")
	if token == "" {
		t.Skip("GRAFANA_CLOUD_TOKEN not set — skipping integration test")
	}
	sinkTempo := NewSinkTempo(
		"https://otlp-gateway-prod-eu-north-0.grafana.net/otlp/v1/traces",
		WithHttpHeader(
			"Authorization",
			"Basic "+token,
		),
	)
	defer sinkTempo.Close()
	telemetry := NewTelemetry(
		WithMode(ModeSync, sinkTempo),
		WithFormat(FormatJson),
	)
	defer telemetry.Close()
	telemetry.Error(DataTrace,
		String("service", "test-service"),
		String("namespace", "payments"),
		String("environment", "production"),
		String("trace_id", "5B8EFFF7-9803-8103-D269-B633813FC700"),
		String("span_id", "EEE19B7EC3C1B100"),
		String("name", "test"),
		Int64("duration", 150),
	)
}
func Test_SinkFactory_Wechat(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}
		var data WechatData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Errorf("failed to decode JSON: %v", err)
		}
		if data.MsgType != "markdown" {
			t.Errorf("expected msgtype 'markdown', got '%s'", data.MsgType)
		}
		if data.Content != "test" {
			t.Errorf("expected content 'test', got '%s'", data.Content)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkWechat := NewSinkWechat(server.URL)
	fields := []Field{
		String("message", "test"),
	}
	_, err := sinkWechat.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkWechat.Close()
}
func Test_SinkFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	sinkFile, err := NewSinkFile(logFile)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer sinkFile.Close()
	data := []byte("test\n")
	n, err := sinkFile.Write(data)
	if err != nil {
		t.Fatalf("Write failed: %v", err)
	}
	if n != len(data) {
		t.Errorf("Expected %d bytes written, got %d", len(data), n)
	}
	content, err := os.ReadFile(logFile)
	if err != nil {
		t.Fatalf("Failed to read file: %v", err)
	}
	if string(content) != string(data) {
		t.Errorf("Expected %q, got %q", data, content)
	}
}
func Test_SinkFile_CleanupByAge(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	sinkFile, err := NewSinkFile(logFile,
		WithFileMaxAge(1),
		WithFileMaxSize(1),
	)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer sinkFile.Close()
	data := make([]byte, 1024)
	for i := 0; i < 3; i++ {
		_, err := sinkFile.Write(data)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)
	files, err := filepath.Glob(filepath.Join(tmpDir, "test*.log*"))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(files) == 0 {
		t.Error("No log files created")
	}
}
func Test_SinkFile_CleanupByCount(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	sinkFile, err := NewSinkFile(logFile,
		WithFileMaxSize(1),
		WithFileMaxBackups(2),
	)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer sinkFile.Close()
	data := make([]byte, 1024)
	for i := 0; i < 5; i++ {
		_, err := sinkFile.Write(data)
		if err != nil {
			t.Fatalf("Write failed: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	time.Sleep(500 * time.Millisecond)
	files, err := filepath.Glob(filepath.Join(tmpDir, "test*.log*"))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(files) > 3 {
		t.Errorf("Expected max 3 files, got %d", len(files))
	}
}
func Test_SinkFile_Concurrent(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	sinkFile, err := NewSinkFile(logFile,
		WithFileMaxSize(1),
		WithFileMaxBackups(5),
	)
	if err != nil {
		t.Fatalf("NewSinkFile failed: %v", err)
	}
	defer sinkFile.Close()
	const (
		goroutines   = 50
		perGoroutine = 200
	)
	var wg sync.WaitGroup
	var writeErrors atomic.Int64
	var bytesWritten atomic.Int64
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				line := fmt.Sprintf("goroutine=%d iteration=%d\n", id, j)
				n, err := sinkFile.Write([]byte(line))
				if err != nil {
					writeErrors.Add(1)
					continue
				}
				bytesWritten.Add(int64(n))
			}
		}(i)
	}
	wg.Wait()
	if n := writeErrors.Load(); n != 0 {
		t.Errorf("unexpected write errors: %d", n)
	}
	if err := sinkFile.Close(); err != nil {
		t.Fatalf("Close failed: %v", err)
	}
	totalBytes := int64(0)
	files, err := filepath.Glob(filepath.Join(tmpDir, "test*.log*"))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	for _, f := range files {
		info, err := os.Stat(f)
		if err != nil {
			continue
		}
		totalBytes += info.Size()
	}
	if totalBytes == 0 {
		t.Error("no bytes written to any log file")
	}
	if bytesWritten.Load() == 0 {
		t.Error("sinkFile.Write returned 0 bytes for all calls")
	}
}
func Test_SinkFile_Rotate(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	sinkFile, err := NewSinkFile(logFile,
		WithFileMaxBackups(3),
		WithFileMaxSize(1),
	)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer sinkFile.Close()
	data := make([]byte, 1024*1024)
	for i := range data {
		data[i] = 'A'
	}
	for i := 0; i < 2; i++ {
		_, err := sinkFile.Write(data)
		if err != nil {
			t.Fatalf("Write failed at iteration %d: %v", i, err)
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err := sinkFile.Close(); err != nil {
		t.Errorf("Sync error: %v", err)
	}
	time.Sleep(1 * time.Second)
	files, err := filepath.Glob(filepath.Join(tmpDir, "test*.log*"))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(files) == 0 {
		t.Error("No log files created")
	}
}
func Test_SinkFile_WriteWithAttributes(t *testing.T) {
	t.Run("Json", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")
		sinkFile, err := NewSinkFile(logFile)
		if err != nil {
			t.Fatalf("NewSinkFile failed: %v", err)
		}
		defer sinkFile.Close()
		attributes := writeAttributes{
			typeData:   DataLog,
			typeFormat: FormatJson,
			typeLevel:  LevelInfo,
		}
		fields := []Field{
			String("message", "test message"),
			Int("count", 42),
		}
		n, err := sinkFile.WriteWithAttributes(attributes, fields)
		if err != nil {
			t.Fatalf("WriteWithAttributes failed: %v", err)
		}
		if n == 0 {
			t.Error("expected non-zero bytes written")
		}
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		output := string(content)
		var record map[string]any
		if err := json.Unmarshal([]byte(strings.TrimSpace(output)), &record); err != nil {
			t.Fatalf("output is not valid JSON: %v\noutput: %q", err, output)
		}
		if record["level"] != "info" {
			t.Errorf("level: expected 'info', got %v", record["level"])
		}
		if record["type"] != "log" {
			t.Errorf("type: expected 'log', got %v", record["type"])
		}
		if record["message"] != "test message" {
			t.Errorf("message: expected 'test message', got %v", record["message"])
		}
		if record["count"] != float64(42) {
			t.Errorf("count: expected 42, got %v", record["count"])
		}
		if _, ok := record["timestamp"]; !ok {
			t.Error("timestamp missing")
		}
	})
	t.Run("Text", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")
		sinkFile, err := NewSinkFile(logFile)
		if err != nil {
			t.Fatalf("NewSinkFile failed: %v", err)
		}
		defer sinkFile.Close()
		attributes := writeAttributes{
			typeData:   DataLog,
			typeFormat: FormatText,
			typeLevel:  LevelInfo,
			theme:      themeDark,
		}
		fields := []Field{
			String("message", "test message"),
			Int("count", 42),
		}
		_, err = sinkFile.WriteWithAttributes(attributes, fields)
		if err != nil {
			t.Fatalf("WriteWithAttributes failed: %v", err)
		}
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		output := string(content)
		if !strings.Contains(output, "message=\"test message\"") {
			t.Errorf("message not found in output: %q", output)
		}
		if !strings.Contains(output, "count=42") {
			t.Errorf("count not found in output: %q", output)
		}
		if !strings.Contains(output, `type="log"`) {
			t.Errorf("type not found in output: %q", output)
		}
		if !strings.Contains(output, "[INFO]") {
			t.Errorf("level prefix not found in output: %q", output)
		}
	})
	t.Run("UnsupportedFormat", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")
		sinkFile, err := NewSinkFile(logFile)
		if err != nil {
			t.Fatalf("NewSinkFile failed: %v", err)
		}
		defer sinkFile.Close()
		attributes := writeAttributes{
			typeData:   DataLog,
			typeFormat: TypeFormat(999),
			typeLevel:  LevelInfo,
		}
		fields := []Field{String("message", "test")}
		_, err = sinkFile.WriteWithAttributes(attributes, fields)
		if err == nil {
			t.Error("expected error for unsupported format")
		}
		if !strings.Contains(err.Error(), "unsupported format") {
			t.Errorf("expected 'unsupported format' error, got: %v", err)
		}
	})
	t.Run("Rotation", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")
		sinkFile, err := NewSinkFile(logFile,
			WithFileMaxSize(1),
			WithFileMaxBackups(3),
		)
		if err != nil {
			t.Fatalf("NewSinkFile failed: %v", err)
		}
		defer sinkFile.Close()
		attributes := writeAttributes{
			typeData:   DataLog,
			typeFormat: FormatJson,
			typeLevel:  LevelInfo,
		}
		bigMessage := strings.Repeat("A", 512*1024)
		fields := []Field{String("message", bigMessage)}
		for i := 0; i < 2; i++ {
			_, err := sinkFile.WriteWithAttributes(attributes, fields)
			if err != nil {
				t.Fatalf("WriteWithAttributes #%d failed: %v", i, err)
			}
		}
		files, err := filepath.Glob(filepath.Join(tmpDir, "test*.log*"))
		if err != nil {
			t.Fatalf("Glob failed: %v", err)
		}
		if len(files) < 2 {
			t.Errorf("expected at least 2 files after rotation, got %d: %v", len(files), files)
		}
	})
	t.Run("ViaTelemetry", func(t *testing.T) {
		tmpDir := t.TempDir()
		logFile := filepath.Join(tmpDir, "test.log")
		sinkFile, err := NewSinkFile(logFile)
		if err != nil {
			t.Fatalf("NewSinkFile failed: %v", err)
		}
		defer sinkFile.Close()
		telemetry := NewTelemetry(
			WithMode(ModeSync, sinkFile),
			WithFormat(FormatJson),
			WithLevel(LevelDebug),
		)
		defer telemetry.Close()
		telemetry.Info(DataLog,
			String("message", "hello from telemetry"),
			Int("user_id", 12345),
		)
		content, err := os.ReadFile(logFile)
		if err != nil {
			t.Fatalf("ReadFile failed: %v", err)
		}
		output := strings.TrimSpace(string(content))
		var record map[string]any
		if err := json.Unmarshal([]byte(output), &record); err != nil {
			t.Fatalf("output is not valid JSON: %v\noutput: %q", err, output)
		}
		if record["message"] != "hello from telemetry" {
			t.Errorf("message: expected 'hello from telemetry', got %v", record["message"])
		}
		if record["user_id"] != float64(12345) {
			t.Errorf("user_id: expected 12345, got %v", record["user_id"])
		}
		if record["level"] != "info" {
			t.Errorf("level: expected 'info', got %v", record["level"])
		}
	})
}
func Test_SinkHttp(t *testing.T) {
	var mutex sync.Mutex
	var receivedBody []byte
	var receivedMethod string
	var receivedHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()
		receivedMethod = r.Method
		receivedHeaders = r.Header.Clone()
		body, _ := io.ReadAll(r.Body)
		receivedBody = body
		r.Body.Close()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkHttp := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpFilterLevel(LevelDebug),
		WithHttpHeader("Content-Type", "application/json"),
		WithHttpMethod("POST"),
	)
	attributes := writeAttributes{
		typeData:  DataLog,
		typeLevel: LevelInfo,
	}
	fields := []Field{String("message", "test")}
	_, err := sinkHttp.WriteWithAttributes(attributes, fields)
	if err != nil {
		t.Fatalf("WriteWithAttributes failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	mutex.Lock()
	defer mutex.Unlock()
	if receivedMethod != "POST" {
		t.Errorf("Expected POST, got %s", receivedMethod)
	}
	if receivedHeaders.Get("Content-Type") != "application/json" {
		t.Errorf("Expected Content-Type application/json, got %s", receivedHeaders.Get("Content-Type"))
	}
	var body map[string]interface{}
	if err := json.Unmarshal(receivedBody, &body); err != nil {
		t.Errorf("Failed to decode body: %v", err)
	}
	if body["message"] != "test" {
		t.Errorf("Expected message 'test', got %v", body["message"])
	}
}
func Test_SinkHttp_Batch(t *testing.T) {
	var (
		mutex    sync.Mutex
		requests [][]byte
	)
	batchInterval := 10 * time.Second
	batchSize := 3
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		r.Body.Close()
		mutex.Lock()
		requests = append(requests, body)
		mutex.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkHttp := NewSinkHttp(server.URL,
		WithHttpBatch(batchSize, batchInterval),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sinkHttp.Close()
	attributes := writeAttributes{
		typeData:  DataLog,
		typeLevel: LevelInfo,
	}
	for i := 0; i < batchSize; i++ {
		fields := []Field{
			String("message", fmt.Sprintf("test-%d", i)),
			Int("count", i),
		}
		_, err := sinkHttp.WriteWithAttributes(attributes, fields)
		if err != nil {
			t.Fatalf("WriteWithAttributes failed: %v", err)
		}
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		mutex.Lock()
		count := len(requests)
		mutex.Unlock()
		if count >= 1 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	mutex.Lock()
	requestCount := len(requests)
	capturedRequests := make([][]byte, len(requests))
	copy(capturedRequests, requests)
	mutex.Unlock()
	if requestCount != 1 {
		t.Fatalf("expected 1 batch request, got %d", requestCount)
	}
	body := capturedRequests[0]
	lines := bytes.Split(bytes.TrimSpace(body), []byte{'\n'})
	if len(lines) != batchSize {
		t.Fatalf("expected %d records in batch, got %d: %s", batchSize, len(lines), body)
	}
	if bytes.Contains(body, []byte("\n\n")) {
		t.Errorf("body contains empty lines (not NDJSON): %q", body)
	}
	expectedCounts := map[int]bool{0: false, 1: false, 2: false}
	expectedMessages := map[string]bool{
		"test-0": false,
		"test-1": false,
		"test-2": false,
	}
	for _, line := range lines {
		var record map[string]any
		if err := json.Unmarshal(line, &record); err != nil {
			t.Fatalf("failed to unmarshal record %q: %v", line, err)
		}
		countFloat, ok := record["count"].(float64)
		if !ok {
			t.Errorf("record missing 'count': %v", record)
			continue
		}
		msg, ok := record["message"].(string)
		if !ok {
			t.Errorf("record missing 'message': %v", record)
		} else if _, exists := expectedMessages[msg]; !exists {
			t.Errorf("unexpected message %q in record: %v", msg, record)
		} else {
			expectedMessages[msg] = true
		}
		countInt := int(countFloat)
		if _, exists := expectedCounts[countInt]; !exists {
			t.Errorf("unexpected count %d in record: %v", countInt, record)
			continue
		}
		expectedCounts[countInt] = true
	}
	for count, found := range expectedCounts {
		if !found {
			t.Errorf("count %d not found in batch", count)
		}
	}
	for msg, found := range expectedMessages {
		if !found {
			t.Errorf("message %q not found in batch", msg)
		}
	}
}
func Test_SinkHttp_Circuit(t *testing.T) {
	newServer := func(behavior func(count int) int) (*httptest.Server, *atomic.Int32) {
		var count atomic.Int32
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := int(count.Add(1))
			if behavior(c) >= 400 {
				w.WriteHeader(behavior(c))
			} else {
				w.WriteHeader(http.StatusOK)
			}
		})), &count
	}
	t.Run("Disabled", func(t *testing.T) {
		srv, cnt := newServer(func(int) int { return 500 })
		defer srv.Close()
		sinkHttp := NewSinkHttp(srv.URL, WithHttpDisabledBatch(), WithHttpDisabledCircuit())
		attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
		fields := []Field{String("msg", "test")}
		for i := 0; i < 10; i++ {
			sinkHttp.WriteWithAttributes(attrs, fields)
		}
		if cnt.Load() != 10 {
			t.Errorf("requests = %d, want 10", cnt.Load())
		}
	})
	t.Run("Close", func(t *testing.T) {
		srv, _ := newServer(func(c int) int {
			if c <= 2 {
				return 500
			}
			return 200
		})
		defer srv.Close()
		sinkHttp := NewSinkHttp(srv.URL, WithHttpDisabledBatch(), WithHttpCircuitBreaker(2, 50*time.Millisecond))
		attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
		fields := []Field{String("msg", "test")}
		sinkHttp.WriteWithAttributes(attrs, fields)
		sinkHttp.WriteWithAttributes(attrs, fields)
		time.Sleep(120 * time.Millisecond)
		sinkHttp.WriteWithAttributes(attrs, fields)
		if sinkHttp.circuitState.Load() != circuitStateClosed {
			t.Error("should be Closed")
		}
	})
	t.Run("HalfOpen", func(t *testing.T) {
		srv, _ := newServer(func(c int) int {
			return 500
		})
		defer srv.Close()
		sinkHttp := NewSinkHttp(srv.URL, WithHttpDisabledBatch(), WithHttpCircuitBreaker(2, 50*time.Millisecond))
		attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
		fields := []Field{String("msg", "test")}
		sinkHttp.WriteWithAttributes(attrs, fields)
		sinkHttp.WriteWithAttributes(attrs, fields)
		time.Sleep(120 * time.Millisecond)
		sinkHttp.WriteWithAttributes(attrs, fields)
		if sinkHttp.circuitState.Load() != circuitStateOpen {
			t.Error("should stay Open after HalfOpen failure")
		}
	})
	t.Run("Open", func(t *testing.T) {
		srv, cnt := newServer(func(int) int { return 500 })
		defer srv.Close()
		sinkHttp := NewSinkHttp(srv.URL, WithHttpDisabledBatch(), WithHttpCircuitBreaker(2, time.Hour))
		attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
		fields := []Field{String("msg", "test")}
		sinkHttp.WriteWithAttributes(attrs, fields)
		sinkHttp.WriteWithAttributes(attrs, fields)
		if sinkHttp.circuitState.Load() != circuitStateOpen {
			t.Error("should be Open")
		}
		_, err := sinkHttp.WriteWithAttributes(attrs, fields)
		if err == nil || err.Error() != "circuit breaker is open" {
			t.Error("should be blocked")
		}
		if cnt.Load() != 2 {
			t.Errorf("requests = %d, want 2", cnt.Load())
		}
	})
	t.Run("Reset", func(t *testing.T) {
		srv, _ := newServer(func(c int) int {
			if c == 1 {
				return 500
			}
			return 200
		})
		defer srv.Close()
		sinkHttp := NewSinkHttp(srv.URL, WithHttpDisabledBatch(), WithHttpCircuitBreaker(5, time.Hour))
		attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
		fields := []Field{String("msg", "test")}
		sinkHttp.WriteWithAttributes(attrs, fields)
		if sinkHttp.circuitFailures.Load() != 1 {
			t.Error("failures should be 1")
		}
		sinkHttp.WriteWithAttributes(attrs, fields)
		if sinkHttp.circuitFailures.Load() != 0 {
			t.Error("failures should be 0")
		}
		if sinkHttp.circuitState.Load() != circuitStateClosed {
			t.Error("state should be Closed")
		}
	})
	t.Run("HalfOpenRecovery", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := requestCount.Add(1)
			if c <= 2 {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer srv.Close()
		sink := NewSinkHttp(srv.URL,
			WithHttpDisabledBatch(),
			WithHttpCircuitBreaker(2, 50*time.Millisecond),
		)
		defer sink.Close()
		attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
		fields := []Field{String("msg", "test")}
		sink.WriteWithAttributes(attrs, fields)
		sink.WriteWithAttributes(attrs, fields)
		if sink.circuitState.Load() != circuitStateOpen {
			t.Fatalf("after 2 failures: state = %d, want Open (%d)",
				sink.circuitState.Load(), circuitStateOpen)
		}
		time.Sleep(120 * time.Millisecond)
		_, err := sink.WriteWithAttributes(attrs, fields)
		if err != nil {
			t.Fatalf("probe request failed: %v", err)
		}
		if sink.circuitState.Load() != circuitStateClosed {
			t.Errorf("after successful probe: state = %d, want Closed (%d)",
				sink.circuitState.Load(), circuitStateClosed)
		}
		if got := requestCount.Load(); got != 3 {
			t.Errorf("request count = %d, want 3 (2 failures + 1 probe)", got)
		}
	})
	t.Run("HalfOpenOnlyOneProbe", func(t *testing.T) {
		var requestCount atomic.Int32
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestCount.Add(1)
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()
		sink := NewSinkHttp(srv.URL,
			WithHttpDisabledBatch(),
			WithHttpCircuitBreaker(2, 50*time.Millisecond),
		)
		defer sink.Close()
		attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
		fields := []Field{String("msg", "test")}
		sink.WriteWithAttributes(attrs, fields)
		sink.WriteWithAttributes(attrs, fields)
		if sink.circuitState.Load() != circuitStateOpen {
			t.Fatalf("should be Open, got %d", sink.circuitState.Load())
		}
		time.Sleep(120 * time.Millisecond)
		var wg sync.WaitGroup
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				sink.WriteWithAttributes(attrs, fields)
			}()
		}
		wg.Wait()
		if got := requestCount.Load(); got != 3 {
			t.Errorf("request count = %d, want 3 (2 failures + exactly 1 probe)", got)
		}
		if sink.circuitState.Load() != circuitStateOpen {
			t.Errorf("after failed probe: state = %d, want Open (%d)",
				sink.circuitState.Load(), circuitStateOpen)
		}
	})
}
func Test_SinkHttp_CloseDuringWrite(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(1 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkHttp := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpDisabledCircuit(),
		WithHttpFilterLevel(LevelDebug),
	)
	const (
		goroutines   = 50
		perGoroutine = 50
	)
	var wg sync.WaitGroup
	var writeErrors atomic.Int64
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				attrs := writeAttributes{
					typeData:  DataLog,
					typeLevel: LevelInfo,
				}
				fields := []Field{
					String("message", fmt.Sprintf("test-%d-%d", id, j)),
					Int("id", id),
					Int("j", j),
				}
				_, err := sinkHttp.WriteWithAttributes(attrs, fields)
				if err != nil {
					writeErrors.Add(1)
				}
			}
		}(i)
	}
	time.Sleep(10 * time.Millisecond)
	if err := sinkHttp.Close(); err != nil {
		t.Errorf("Close failed: %v", err)
	}
	wg.Wait()
	t.Logf("write errors after Close: %d", writeErrors.Load())
}
func Test_SinkHttp_Deduplication(t *testing.T) {
	t.Run("Basic", func(t *testing.T) {
		var mutex sync.Mutex
		var requestCount int
		deduplication := 1 * time.Second
		shortDelay := 10 * time.Millisecond
		mediumDelay := 100 * time.Millisecond
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			requestCount++
			mutex.Unlock()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		sinkHttp := NewSinkHttp(server.URL,
			WithHttpDedupWindow(deduplication),
			WithHttpDisabledBatch(),
			WithHttpFilterLevel(LevelDebug),
		)
		defer sinkHttp.Close()
		attributes := writeAttributes{
			typeData:  DataLog,
			typeLevel: LevelInfo,
		}
		fields := []Field{String("message", "test")}
		sinkHttp.WriteWithAttributes(attributes, fields)
		time.Sleep(shortDelay)
		sinkHttp.WriteWithAttributes(attributes, fields)
		time.Sleep(mediumDelay)
		mutex.Lock()
		count := requestCount
		mutex.Unlock()
		if count != 1 {
			t.Errorf("Expected 1 request (deduplication), got %d", count)
		}
	})
	t.Run("MaxSize", func(t *testing.T) {
		var mutex sync.Mutex
		var requestCount int
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			requestCount++
			mutex.Unlock()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		sinkHttp := NewSinkHttp(server.URL,
			WithHttpDedupWindow(1*time.Second),
			WithHttpDedupMaxSize(2),
			WithHttpDisabledBatch(),
			WithHttpFilterLevel(LevelDebug),
		)
		defer sinkHttp.Close()
		attributes := writeAttributes{
			typeData:  DataLog,
			typeLevel: LevelInfo,
		}
		for i := 0; i < 5; i++ {
			fields := []Field{String("message", fmt.Sprintf("test-%d", i))}
			sinkHttp.WriteWithAttributes(attributes, fields)
		}
		time.Sleep(50 * time.Millisecond)
		mutex.Lock()
		count := requestCount
		mutex.Unlock()
		if count != 5 {
			t.Errorf("Expected 5 requests (all unique), got %d", count)
		}
	})
}
func Test_SinkHttp_RateLimit(t *testing.T) {
	var mutex sync.Mutex
	attempt := 0
	backoff := 100 * time.Millisecond
	retryMax := 2
	retryAfterSeconds := 1
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		attempt++
		mutex.Unlock()
		if attempt == 1 {
			w.Header().Set("Retry-After", strconv.Itoa(retryAfterSeconds))
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	start := time.Now()
	sinkHttp := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpFilterLevel(LevelDebug),
		WithHttpRetry(retryMax, backoff),
	)
	attributes := writeAttributes{
		typeData:  DataLog,
		typeLevel: LevelInfo,
	}
	fields := []Field{String("message", "test")}
	_, err := sinkHttp.WriteWithAttributes(attributes, fields)
	if err != nil {
		t.Fatalf("WriteWithAttributes failed: %v", err)
	}
	time.Sleep(100 * time.Millisecond)
	elapsed := time.Since(start)
	expectedDuration := time.Duration(retryAfterSeconds) * time.Second
	mutex.Lock()
	finalAttempt := attempt
	mutex.Unlock()
	if finalAttempt != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempt)
	}
	if elapsed < expectedDuration {
		t.Errorf("Expected to wait at least 1 second due to Retry-After, got %v", elapsed)
	}
}
func Test_SinkHttp_Retry(t *testing.T) {
	attempt := 0
	backoff := 10 * time.Millisecond
	retry := 3
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt < retry {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkHttp := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpFilterLevel(LevelDebug),
		WithHttpRetry(retry, backoff),
	)
	attributes := writeAttributes{
		typeData:  DataLog,
		typeLevel: LevelInfo,
	}
	fields := []Field{String("message", "test")}
	sinkHttp.WriteWithAttributes(attributes, fields)
	if attempt != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempt)
	}
}
func Test_SinkHttp_Sampling(t *testing.T) {
	var mutex sync.Mutex
	var requestCount int
	counts := 100
	rate := int32(10)
	expected := counts / int(rate)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requestCount++
		mutex.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkHttp := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpFilterLevel(LevelDebug),
		WithHttpSampleRate(rate),
	)
	defer sinkHttp.Close()
	attributes := writeAttributes{
		typeData:  DataLog,
		typeLevel: LevelInfo,
	}
	fields := []Field{String("message", "test")}
	for i := 0; i < counts; i++ {
		sinkHttp.WriteWithAttributes(attributes, fields)
	}
	time.Sleep(100 * time.Millisecond)
	mutex.Lock()
	count := requestCount
	mutex.Unlock()
	if count != expected {
		t.Errorf("Expected exactly %d requests (deterministic sampling), got %d",
			expected, count)
	}
}
func Test_Stress_Sink_TimeConcurrent(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			buf := &bytes.Buffer{}
			for j := 0; j < 1000; j++ {
				buf.Reset()
				getTime(buf, time.Now())
			}
		}()
	}
	wg.Wait()
}
func Test_Stress_SinkHttp_ConcurrentWrite(t *testing.T) {
	var (
		mutex        sync.Mutex
		requestCount int
	)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mutex.Lock()
		requestCount++
		mutex.Unlock()
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkHttp := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpDisabledCircuit(),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sinkHttp.Close()
	attributes := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
	const (
		goroutines   = 100
		perGoroutine = 100
	)
	var wg sync.WaitGroup
	wg.Add(goroutines)
	var writeErrors atomic.Int64
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				fields := []Field{
					String("message", fmt.Sprintf("test-%d-%d", id, j)),
					Int("id", id),
					Int("j", j),
				}
				if _, err := sinkHttp.WriteWithAttributes(attributes, fields); err != nil {
					writeErrors.Add(1)
				}
			}
		}(i)
	}
	wg.Wait()
	if n := writeErrors.Load(); n != 0 {
		t.Errorf("unexpected write errors: %d", n)
	}
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		mutex.Lock()
		count := requestCount
		mutex.Unlock()
		if count >= goroutines*perGoroutine {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	mutex.Lock()
	finalCount := requestCount
	mutex.Unlock()
	if finalCount != goroutines*perGoroutine {
		t.Errorf("expected %d requests, got %d", goroutines*perGoroutine, finalCount)
	}
}
func Test_Stress_SinkHttp_BatchTickerStress(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sinkHttp := NewSinkHttp(server.URL,
		WithHttpBatch(10, 100*time.Millisecond),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sinkHttp.Close()
	for i := 0; i < 100; i++ {
		WithHttpBatch(10, time.Duration(50+i)*time.Millisecond)(sinkHttp)
	}
	attributes := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
	fields := []Field{String("message", "test")}
	_, err := sinkHttp.WriteWithAttributes(attributes, fields)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(300 * time.Millisecond)
	sinkHttp.batchMutex.Lock()
	ticker := sinkHttp.batchTicker
	sinkHttp.batchMutex.Unlock()
	if ticker == nil {
		t.Error("ticker is nil after stress")
	}
}
func Test_Stress_SinkHttp_BatchToggle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sink := NewSinkHttp(server.URL,
		WithHttpBatch(10, 50*time.Millisecond),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sink.Close()
	attrs := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
	fields := []Field{String("message", "test")}
	sink.WriteWithAttributes(attrs, fields)
	WithHttpDisabledBatch()(sink)
	sink.WriteWithAttributes(attrs, fields)
	WithHttpBatch(5, 50*time.Millisecond)(sink)
	for i := 0; i < 5; i++ {
		sink.WriteWithAttributes(attrs, fields)
	}
	time.Sleep(200 * time.Millisecond)
	sink.batchMutex.Lock()
	ticker := sink.batchTicker
	sink.batchMutex.Unlock()
	if ticker == nil {
		t.Error("ticker should be set")
	}
}
func Test_Stress_SinkHttp_CircuitStuckProbe(t *testing.T) {
	var requestCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	sink := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpCircuitBreaker(2, 50*time.Millisecond),
	)
	defer sink.Close()
	attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
	fields := []Field{String("msg", "test")}
	sink.WriteWithAttributes(attrs, fields)
	sink.WriteWithAttributes(attrs, fields)
	if sink.circuitState.Load() != circuitStateOpen {
		t.Fatalf("should be Open, got %d", sink.circuitState.Load())
	}
	time.Sleep(60 * time.Millisecond)
	go sink.WriteWithAttributes(attrs, fields)
	time.Sleep(20 * time.Millisecond)
	time.Sleep(100 * time.Millisecond)
	if sink.circuitHalfOpenProbeInFlight.Load() {
		t.Error("probe should be reset after timeout")
	}
}
func Test_Stress_SinkHttp_CircuitToggle(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()
	sink := NewSinkHttp(server.URL,
		WithHttpDisabledBatch(),
		WithHttpCircuitBreaker(2, 50*time.Millisecond),
	)
	defer sink.Close()
	attrs := writeAttributes{typeLevel: LevelError, typeData: DataLog}
	fields := []Field{String("msg", "test")}
	sink.WriteWithAttributes(attrs, fields)
	sink.WriteWithAttributes(attrs, fields)
	if sink.circuitState.Load() != circuitStateOpen {
		t.Fatal("should be Open")
	}
	WithHttpDisabledCircuit()(sink)
	if sink.circuitState.Load() != circuitStateClosed {
		t.Error("state should be Closed after disable")
	}
	if sink.circuitFailures.Load() != 0 {
		t.Error("failures should be 0")
	}
	_, err := sink.WriteWithAttributes(attrs, fields)
	if err != nil && err.Error() == "circuit breaker is open" {
		t.Error("circuit breaker should be disabled")
	}
}
func Test_Stress_SinkHttp_Deduplication_EvictionStress(t *testing.T) {
	t.Run("SizeEviction", func(t *testing.T) {
		var (
			mutex        sync.Mutex
			requestCount int
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			requestCount++
			mutex.Unlock()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		const (
			maxSize    = 100
			totalWrite = 1200
		)
		sink := NewSinkHttp(server.URL,
			WithHttpDedupWindow(1*time.Hour),
			WithHttpDedupMaxSize(maxSize),
			WithHttpDisabledBatch(),
			WithHttpDisabledCircuit(),
			WithHttpFilterLevel(LevelDebug),
		)
		defer sink.Close()
		attrs := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
		for i := 0; i < totalWrite; i++ {
			fields := []Field{String("message", fmt.Sprintf("unique-%d", i))}
			sink.WriteWithAttributes(attrs, fields)
		}
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			if sink.dedupCacheCount.Load() <= maxSize {
				break
			}
			time.Sleep(50 * time.Millisecond)
		}
		cacheSize := sink.dedupCacheCount.Load()
		if cacheSize > maxSize {
			t.Errorf("dedupCacheCount = %d, want <= %d after size-based eviction",
				cacheSize, maxSize)
		}
		if cacheSize == 0 {
			t.Error("dedupCacheCount = 0 — eviction removed everything (unexpected)")
		}
	})
	t.Run("TTLEviction", func(t *testing.T) {
		var (
			mutex        sync.Mutex
			requestCount int
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			requestCount++
			mutex.Unlock()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		sink := NewSinkHttp(server.URL,
			WithHttpDedupWindow(100*time.Millisecond),
			WithHttpDedupMaxSize(10),
			WithHttpDisabledBatch(),
			WithHttpDisabledCircuit(),
			WithHttpFilterLevel(LevelDebug),
		)
		defer sink.Close()
		attrs := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
		for i := 0; i < 50; i++ {
			fields := []Field{String("message", fmt.Sprintf("unique-%d", i))}
			sink.WriteWithAttributes(attrs, fields)
		}
		deadline := time.Now().Add(2 * time.Second)
		for time.Now().Before(deadline) {
			if sink.dedupCacheCount.Load() == 0 {
				break
			}
			time.Sleep(10 * time.Millisecond)
		}
		cacheSize := sink.dedupCacheCount.Load()
		if cacheSize != 0 {
			t.Errorf("dedupCacheCount = %d, want 0 after TTL eviction", cacheSize)
		}
	})
	t.Run("UniqueInWindow", func(t *testing.T) {
		var (
			mutex        sync.Mutex
			requestCount int
		)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mutex.Lock()
			requestCount++
			mutex.Unlock()
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()
		sink := NewSinkHttp(server.URL,
			WithHttpDedupWindow(500*time.Millisecond),
			WithHttpDisabledBatch(),
			WithHttpDisabledCircuit(),
			WithHttpFilterLevel(LevelDebug),
		)
		defer sink.Close()
		attrs := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
		for i := 0; i < 100; i++ {
			fields := []Field{String("message", fmt.Sprintf("unique-%d", i))}
			sink.WriteWithAttributes(attrs, fields)
		}
		time.Sleep(100 * time.Millisecond)
		mutex.Lock()
		count := requestCount
		mutex.Unlock()
		if count != 100 {
			t.Errorf("expected 100 requests, got %d", count)
		}
		cacheSize := sink.dedupCacheCount.Load()
		if cacheSize != 100 {
			t.Errorf("dedupCacheCount = %d, want 100 (TTL not yet triggered)", cacheSize)
		}
	})
}
func Test_Stress_SinkHttp_LongRun(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long run test in short mode")
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sink := NewSinkHttp(server.URL,
		WithHttpBatch(10, 50*time.Millisecond),
		WithHttpDedupWindow(100*time.Millisecond),
		WithHttpCircuitBreaker(10, 100*time.Millisecond),
		WithHttpFilterLevel(LevelDebug),
	)
	defer sink.Close()
	attrs := writeAttributes{typeData: DataLog, typeLevel: LevelInfo}
	deadline := time.Now().Add(5 * time.Second)
	var count atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for time.Now().Before(deadline) {
				fields := []Field{
					String("message", "test-"+strconv.Itoa(id)),
					Int("id", id),
				}
				sink.WriteWithAttributes(attrs, fields)
				count.Add(1)
			}
		}(i)
	}
	wg.Wait()
	t.Logf("wrote %d messages in 5 seconds", count.Load())
	if count.Load() == 0 {
		t.Error("expected messages to be written")
	}
	if sink.closed {
		t.Error("sink should not be closed")
	}
}
func Test_Stress_TeeSink_CloseIdempotent(t *testing.T) {
	var closeCount atomic.Int32
	fake := &fakeCloser{closeCount: &closeCount}
	tee := NewTeeSink(fake)

	err1 := tee.Close()
	err2 := tee.Close()

	if err1 != nil {
		t.Errorf("first Close: %v", err1)
	}
	if err2 != nil {
		t.Errorf("second Close: %v", err2)
	}
	if closeCount.Load() != 1 {
		t.Errorf("Close called %d times, want 1", closeCount.Load())
	}
}

func Test_Stress_TeeSink_ReplaceDoesNotClose(t *testing.T) {
	var closeCount atomic.Int32
	fake := &fakeCloser{closeCount: &closeCount}
	tee := NewTeeSink(fake)
	defer tee.Close()

	tee.Replace(0, &bytes.Buffer{})

	if closeCount.Load() != 0 {
		t.Errorf("Replace should not close old writer, closed %d times", closeCount.Load())
	}
}

type fakeCloser struct {
	closeCount *atomic.Int32
}

func (f *fakeCloser) Write(p []byte) (int, error) { return len(p), nil }
func (f *fakeCloser) Close() error {
	f.closeCount.Add(1)
	return nil
}

// Приватные функции
func loadEnv(path string) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		os.Setenv(key, value)
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "ulog: error reading .env: %v\n", err)
	}
}
func checkExtractor(t *testing.T, elem struct {
	name      string
	keys      []string
	context   context.Context
	want      map[string]string
	shouldAdd bool
}, output string) {
	t.Helper()
	if elem.shouldAdd {
		for key, value := range elem.want {
			if !strings.Contains(output, key) {
				t.Errorf("extractor with keys %v: expected field %q not found in output: %s",
					elem.keys, key, output)
			}
			if !strings.Contains(output, value) {
				t.Errorf("extractor with keys %v: expected value %q for key %q not found in output: %s",
					elem.keys, value, key, output)
			}
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
	name        string
	theme       TypeTheme
	callerColor string
	dataColor   string
	prefixDebug string
	prefixError string
	prefixFatal string
	prefixInfo  string
	prefixWarn  string
	reset       string
}, output string) {
	t.Helper()
	if !strings.Contains(output, elem.callerColor) && level == "DEBUG" {
		t.Errorf("%s: expected prefix %q not found in %q", level, elem.callerColor, output)
	}
	if !strings.Contains(output, expectedPrefix) {
		t.Errorf("%s: expected prefix %q not found in %q", level, expectedPrefix, output)
	}
	if !strings.Contains(output, elem.dataColor) {
		t.Errorf("%s: expected data color %q not found", level, elem.dataColor)
	}
	if !strings.Contains(output, elem.reset) {
		t.Errorf("%s: expected data color %q not found", level, elem.reset)
	}
}
func testDebug(telemetry Telemetry) {
	telemetry.Debug(DataLog, String("message", "test debug text"))
}
func testDebugWithContext(telemetry Telemetry) {
	telemetry.DebugWithContext(context.Background(), DataLog, String("message", "test debug text"))
}
func testError(telemetry Telemetry) {
	telemetry.Error(DataLog, String("message", "test error text"))
}
func testErrorWithContext(telemetry Telemetry) {
	telemetry.ErrorWithContext(context.Background(), DataLog, String("message", "test error text"))
}
func testFatal(telemetry Telemetry) {
	oldExit := osExit
	osExit = func(int) {}
	defer func() { osExit = oldExit }()
	telemetry.Fatal(DataLog, String("message", "test fatal text"))
}
func testFatalWithContext(telemetry Telemetry) {
	oldExit := osExit
	osExit = func(int) {}
	defer func() { osExit = oldExit }()
	telemetry.FatalWithContext(context.Background(), DataLog, String("message", "test fatal text"))
}
func testInfo(telemetry Telemetry) {
	telemetry.Info(DataLog, String("message", "test info text"))
}
func testInfoWithContext(telemetry Telemetry) {
	telemetry.InfoWithContext(context.Background(), DataLog, String("message", "test info text"))
}
func testWarn(telemetry Telemetry) {
	telemetry.Warn(DataLog, String("message", "test warn text"))
}
func testWarnWithContext(telemetry Telemetry) {
	telemetry.WarnWithContext(context.Background(), DataLog, String("message", "test warn text"))
}
