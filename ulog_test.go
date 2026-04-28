package ulog

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// Публичные функции
func Test_Telemetry(t *testing.T) {
	buf := &bytes.Buffer{}
	telemetry := NewTelemetry(
		WithFormat(FormatText),
		WithMode(ModeSync, buf),
	)
	if telemetry == nil {
		t.Fatal("NewTelemetry returned nil")
	}
	telemetry.Info(DataLog, String("message", "test info text"))
	telemetry.Sync()
	if !strings.Contains(buf.String(), "test info text") {
		t.Errorf("Expected 'test info text', got %q", buf.String())
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
func Test_Telemetry_Extractor(t *testing.T) {
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
			telemetry := NewTelemetry(
				WithExtractor(elem.keys...),
				WithFormat(FormatJson),
				WithMode(ModeSync, buf),
			)
			telemetry.InfoWithContext(elem.context, DataLog, String("message", "test info text"))
			telemetry.Sync()
			output := buf.String()
			checkExtractor(t, elem, output)
		})
		t.Run("SetExtractor/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry()
			telemetry.SetExtractor(elem.keys...)
			telemetry.SetFormat(FormatJson)
			telemetry.SetMode(ModeSync, buf)
			telemetry.InfoWithContext(elem.context, DataLog, String("message", "test info text"))
			telemetry.Sync()
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
			telemetry.Info(DataLog, String("message", "test info text"))
			telemetry.Sync()
			output := buf.String()
			if !strings.Contains(output, elem.expect) {
				t.Errorf("Expected output to contain %q, got %q", elem.expect, output)
			}
		})
		t.Run("SetFormat/"+elem.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			telemetry := NewTelemetry()
			telemetry.SetFormat(elem.format)
			telemetry.SetMode(ModeSync, buf, 0)
			telemetry.Info(DataLog, String("message", "test info text"))
			telemetry.Sync()
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
			elem.functionTest(telemetry)
			telemetry.Sync()
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
			telemetry.Sync()
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
			elem.functionTest(telemetry)
			telemetry.Sync()
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
		telemetry.Sync()
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
		telemetry.Sync()
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
		telemetry.Sync()
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
		telemetry.Sync()
		if writerBuf.Len() == 0 {
			t.Error("Sync mode: expected output, got nothing")
		}
		if !strings.Contains(writerBuf.String(), "test info text") {
			t.Error("Sync mode: expected message not found")
		}
	})
}
func Test_Telemetry_Sync(t *testing.T) {
	t.Run("Async", func(t *testing.T) {
		buf := &bytes.Buffer{}
		telemetry := NewTelemetry(WithMode(ModeAsync, buf, 1000))
		defer telemetry.Close()
		telemetry.Info(DataLog, String("message", "test info text"))
		err := telemetry.Sync()
		if err != nil {
			t.Errorf("Sync() returned error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "test info text") {
			t.Error("Message not written after Sync")
		}
	})
	t.Run("Sync", func(t *testing.T) {
		buf := &bytes.Buffer{}
		telemetry := NewTelemetry(WithMode(ModeSync, buf))
		defer telemetry.Close()
		telemetry.Info(DataLog, String("message", "test info text"))
		err := telemetry.Sync()
		if err != nil {
			t.Errorf("Sync() returned error: %v", err)
		}
		output := buf.String()
		if !strings.Contains(output, "test info text") {
			t.Error("Message not written after Sync")
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
				functionTest(telemetry)
				telemetry.Sync()
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
				telemetry.SetFormat(FormatText)
				telemetry.SetLevel(LevelDebug)
				telemetry.SetMode(ModeSync, buf)
				telemetry.SetTheme(elem.theme)
				functionTest(telemetry)
				telemetry.Sync()
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
		t.Errorf("WriteWithAttributes: buf1 should be empty (removed), got %q", buf1.String())
	}
	if buf2.String() != data {
		t.Errorf("WriteWithAttributes: buf2 should be empty (removed), got %q", buf2.String())
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
	if buf3.String() != "" {
		t.Errorf("WriteWithAttributes: buf3 is empty")
	}
	if buf4.String() != "" {
		t.Errorf("WriteWithAttributes: buf4 is empty")
	}
	err := tee.Close()
	if err != nil {
		t.Errorf("Close() returned error: %v", err)
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
		if data.Content != "test message" {
			t.Errorf("expected content 'test message', got '%s'", data.Content)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	sinkDiscord := NewSinkDiscord(server.URL, "ULog Bot", "")
	fields := []Field{
		String("message", "test message"),
	}
	_, err := sinkDiscord.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkDiscord.Sync()
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
			"message":  "test kafka message",
			"service":  "test-service",
			"trace_id": "trace-abc-123",
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
		if record.Key != "trace-abc-123" {
			t.Errorf("key: expected 'trace-abc-123', got '%s'", record.Key)
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
	sinkKafka := NewSinkKafka(server.URL, "test-topic")
	fields := []Field{
		String("message", "test kafka message"),
		String("service", "test-service"),
		String("trace_id", "trace-abc-123"),
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
	sinkKafka.Sync()
}
func Test_SinkFactory_Loki(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Error("wrong method")
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Error("wrong Content-Type")
		}
		var data LokiData
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			t.Errorf("failed to decode JSON: %v", err)
		}
		if len(data.Streams) == 0 {
			t.Error("expected at least one stream")
			return
		}
		stream := data.Streams[0]
		if stream.Stream["app"] != "test-app" {
			t.Errorf("wrong app label: got '%s', want 'test-app'", stream.Stream["app"])
		}
		if stream.Stream["level"] != "ERROR" {
			t.Errorf("wrong level label: got '%s', want 'error'", stream.Stream["level"])
		}
		if stream.Stream["trace_id"] != "abc-123" {
			t.Errorf("wrong trace_id label: got '%s', want 'abc-123'", stream.Stream["trace_id"])
		}
		if len(stream.Values) == 0 {
			t.Error("expected at least one value")
			return
		}
		value := stream.Values[0]
		if len(value) != 2 {
			t.Errorf("expected [timestamp, logLine], got %v", value)
			return
		}
		if _, err := strconv.ParseInt(value[0], 10, 64); err != nil {
			t.Errorf("timestamp should be numeric: %s", value[0])
		}
		if !strings.Contains(value[1], "test error message") {
			t.Errorf("log line should contain message: %s", value[1])
		}
		if !strings.Contains(value[1], "user_id") || !strings.Contains(value[1], "12345") {
			t.Errorf("log line should contain extra fields: %s", value[1])
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()
	sinkLoki := NewSinkLoki(server.URL, map[string]string{
		"app": "test-app",
		"env": "test",
	})
	fields := []Field{
		String("message", "test error message"),
		String("user_id", "12345"),
		String("trace_id", "abc-123"),
	}
	_, err := sinkLoki.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkLoki.Sync()
}
func Test_SinkFactory_Prometheus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "text/plain" {
			t.Errorf("expected Content-Type: text/plain, got %s", r.Header.Get("Content-Type"))
		}
		body := make([]byte, 1024)
		n, _ := r.Body.Read(body)
		content := string(body[:n])
		if content == "" {
			t.Error("expected non-empty Prometheus metric")
		}
		expectedPattern := "http_requests_total"
		if len(content) < len(expectedPattern) || content[:len(expectedPattern)] != expectedPattern {
			t.Errorf("expected metric name 'http_requests_total', got '%s'", content)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkPrometheus := NewSinkPrometheus(server.URL)
	fields := []Field{
		String("name", "http_requests_total"),
		String("method", "GET"),
		Float64("value", 42.0),
	}
	_, err := sinkPrometheus.WriteWithAttributes(
		writeAttributes{typeData: DataMetric},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkPrometheus.Sync()
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
		if data.Text != "test message" {
			t.Errorf("expected text 'test message', got '%s'", data.Text)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkSlack := NewSinkSlack(server.URL, "ULog", ":robot:", "", "#alerts")
	fields := []Field{
		String("message", "test message"),
	}
	_, err := sinkSlack.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkSlack.Sync()
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
		if data.ChatID != "test-chat-123" {
			t.Errorf("expected chat_id 'test-chat-123', got '%s'", data.ChatID)
		}
		if data.Text != "test message" {
			t.Errorf("expected text 'test message', got '%s'", data.Text)
		}
		if data.ParseMode != "HTML" {
			t.Errorf("expected parse_mode 'HTML', got '%s'", data.ParseMode)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkTelegram := NewSinkTelegram("test-bot-token", "test-chat-123")
	sinkTelegram.endPoint = server.URL
	fields := []Field{
		String("message", "test message"),
	}
	_, err := sinkTelegram.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkTelegram.Sync()
}
func Test_SinkFactory_Tempo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var trace TempoData
		json.NewDecoder(r.Body).Decode(&trace)
		if trace.Duration != 100 {
			t.Error("wrong duration")
		}
		if trace.Name != "test" {
			t.Error("wrong name")
		}
		if trace.SpanID != "def" {
			t.Error("wrong span_id")
		}
		if trace.TraceID != "abc" {
			t.Error("wrong trace_id")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkTempo := NewSinkTempo(server.URL)
	fields := []Field{
		String("trace_id", "abc"),
		String("span_id", "def"),
		String("name", "test"),
		Int64("duration", 100),
	}
	sinkTempo.WriteWithAttributes(writeAttributes{typeData: DataTrace}, fields)
	sinkTempo.Sync()
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
		if data.Content != "test message" {
			t.Errorf("expected content 'test message', got '%s'", data.Content)
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	sinkWechat := NewSinkWechat(server.URL)
	fields := []Field{
		String("message", "test message"),
	}
	_, err := sinkWechat.WriteWithAttributes(
		writeAttributes{typeLevel: LevelError, typeData: DataLog},
		fields,
	)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	sinkWechat.Sync()
}
func Test_SinkFile(t *testing.T) {
	tmpDir := t.TempDir()
	logFile := filepath.Join(tmpDir, "test.log")
	sinkFile, err := NewSinkFile(logFile)
	if err != nil {
		t.Fatalf("NewFileSink failed: %v", err)
	}
	defer sinkFile.Close()
	data := []byte("test message\n")
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
	time.Sleep(200 * time.Millisecond)
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
	time.Sleep(200 * time.Millisecond)
	files, err := filepath.Glob(filepath.Join(tmpDir, "test*.log*"))
	if err != nil {
		t.Fatalf("Glob failed: %v", err)
	}
	if len(files) > 3 {
		t.Errorf("Expected max 3 files, got %d", len(files))
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
	if err := sinkFile.Sync(); err != nil {
		t.Errorf("Sync error: %v", err)
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
	var mutex sync.Mutex
	var requests [][]byte
	batchInterval := 100 * time.Millisecond
	batchSize := 3
	delay := 200 * time.Millisecond
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
	attributes := writeAttributes{
		typeData:  DataLog,
		typeLevel: LevelInfo,
	}
	for i := 0; i < batchSize; i++ {
		fields := []Field{
			String("message", "test"),
			Int("count", i),
		}
		sinkHttp.WriteWithAttributes(attributes, fields)
	}
	time.Sleep(delay)
	mutex.Lock()
	requestCount := len(requests)
	mutex.Unlock()
	if requestCount != 1 {
		t.Errorf("Expected 1 batch request, got %d", requestCount)
	}
	if requestCount > 0 {
		bodyStr := string(requests[0])
		if !strings.Contains(bodyStr, "0") || !strings.Contains(bodyStr, "1") || !strings.Contains(bodyStr, "2") {
			t.Errorf("Batch should contain all messages: %s", bodyStr)
		}
	}
}
func Test_SinkHttp_Circuit(t *testing.T) {
	newServer := func(behavior func(count int) int) (*httptest.Server, *int32) {
		var count int32
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := int(atomic.AddInt32(&count, 1))
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
		if atomic.LoadInt32(cnt) != 10 {
			t.Errorf("requests = %d, want 10", *cnt)
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
		time.Sleep(60 * time.Millisecond)
		sinkHttp.WriteWithAttributes(attrs, fields)
		if atomic.LoadInt32(&sinkHttp.circuitState) != circuitStateClosed {
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
		time.Sleep(60 * time.Millisecond)
		sinkHttp.WriteWithAttributes(attrs, fields)
		if atomic.LoadInt32(&sinkHttp.circuitState) != circuitStateOpen {
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
		if atomic.LoadInt32(&sinkHttp.circuitState) != circuitStateOpen {
			t.Error("should be Open")
		}
		_, err := sinkHttp.WriteWithAttributes(attrs, fields)
		if err == nil || err.Error() != "circuit breaker is open" {
			t.Error("should be blocked")
		}
		if atomic.LoadInt32(cnt) != 2 {
			t.Errorf("requests = %d, want 2", *cnt)
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
		if atomic.LoadInt32(&sinkHttp.circuitFailures) != 1 {
			t.Error("failures should be 1")
		}
		sinkHttp.WriteWithAttributes(attrs, fields)
		if atomic.LoadInt32(&sinkHttp.circuitFailures) != 0 {
			t.Error("failures should be 0")
		}
	})
}
func Test_SinkHttp_Deduplication(t *testing.T) {
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
}
func Test_SinkHttp_RateLimit(t *testing.T) {
	var mutex sync.Mutex
	attempt := 0
	backoff := 100 * time.Millisecond
	retry := 2
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
		WithHttpRetry(retry, backoff),
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
	if finalAttempt != retry {
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
	if attempt != retry {
		t.Errorf("Expected 3 attempts, got %d", attempt)
	}
}
func Test_SinkHttp_Sampling(t *testing.T) {
	var mutex sync.Mutex
	var requestCount int
	counts := 100
	rate := int32(10)
	expected := counts / int(rate)
	delta := expected / 2
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
	if count < expected-delta || count > expected+delta {
		t.Errorf("Expected ~10 requests, got %d", count)
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
