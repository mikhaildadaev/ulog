package ulog_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/mikhaildadaev/ulog"
)

// Примеры использования публичных конструкторов
func ExampleNewTelemetry() {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	telemetry := ulog.NewTelemetry(
		ulog.WithExtractor("node_id", "trace_id"),
		ulog.WithFormat(ulog.FormatJson),
		ulog.WithLevel(ulog.LevelDebug),
		ulog.WithMode(ulog.ModeAsync, buf, 1000),
		ulog.WithTheme(ulog.ThemeDark),
	)
	defer telemetry.Close()
	telemetry.DebugWithContext(ctx, ulog.DataLog, ulog.String("message", "debug test text"))
	telemetry.ErrorWithContext(ctx, ulog.DataLog, ulog.String("message", "error test text"))
	telemetry.InfoWithContext(ctx, ulog.DataLog, ulog.String("message", "info test text"))
	telemetry.WarnWithContext(ctx, ulog.DataLog, ulog.String("message", "warn test text"))
	telemetry.Sync()
	telemetry.SetExtractor()
	telemetry.SetFormat(ulog.FormatText)
	telemetry.SetLevel(ulog.LevelDebug)
	telemetry.SetMode(ulog.ModeSync, buf)
	telemetry.Debug(ulog.DataLog, ulog.String("message", "debug test text"))
	telemetry.Error(ulog.DataLog, ulog.String("message", "error test text"))
	telemetry.Info(ulog.DataLog, ulog.String("message", "info test text"))
	telemetry.Warn(ulog.DataLog, ulog.String("message", "warn test text"))
	telemetry.Sync()
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	// {"level":"debug","type":"log","message":"debug test text","node_id":"123-abc","trace_id":"abc-123"}
	// {"level":"error","type":"log","message":"error test text","node_id":"123-abc","trace_id":"abc-123"}
	// {"level":"info","type":"log","message":"info test text","node_id":"123-abc","trace_id":"abc-123"}
	// {"level":"warn","type":"log","message":"warn test text","node_id":"123-abc","trace_id":"abc-123"}
	// [DEBUG] type="log" message="debug test text"
	// [ERROR] type="log" message="error test text"
	// [INFO] type="log" message="info test text"
	// [WARN] type="log" message="warn test text"
}
func ExampleNewTelemetryLog() {
	buf := &bytes.Buffer{}
	telemetry := ulog.NewTelemetry(
		ulog.WithMode(ulog.ModeSync, buf),
		ulog.WithFormat(ulog.FormatText),
		ulog.WithLevel(ulog.LevelError),
	)
	telemetryLog := ulog.NewTelemetryLog(ulog.LevelError, telemetry)
	telemetryLog.Print("this will be logged as ERROR")
	telemetryLog.Printf("user %s failed to login", "john")
	telemetryLog.Println("another error message")
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	// [ERROR] type="log" message="this will be logged as ERROR"
	// [ERROR] type="log" message="user john failed to login"
	// [ERROR] type="log" message="another error message"
}
func ExampleTelemetry_data() {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	telemetry := ulog.NewTelemetry(
		ulog.WithExtractor("node_id", "trace_id"),
		ulog.WithMode(ulog.ModeSync, buf),
	)
	defer telemetry.Close()
	// DataLog
	telemetry.Info(ulog.DataLog,
		ulog.String("message", "user login"))
	telemetry.InfoWithContext(ctx, ulog.DataLog,
		ulog.String("message", "user login"))
	// DataMetric
	telemetry.Info(ulog.DataMetric,
		ulog.String("name", "logins"),
		ulog.Float64("value", 1.0),
	)
	telemetry.InfoWithContext(ctx, ulog.DataMetric,
		ulog.String("name", "logins"),
		ulog.Float64("value", 1.0),
	)
	// DataTrace
	telemetry.Info(ulog.DataTrace,
		ulog.String("span_id", "def"),
		ulog.String("name", "login"),
		ulog.Int64("duration", 150),
	)
	telemetry.InfoWithContext(ctx, ulog.DataTrace,
		ulog.String("span_id", "def"),
		ulog.String("name", "login"),
		ulog.Int64("duration", 150),
	)
	telemetry.Sync()
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	//{"level":"info","type":"log","message":"user login"}
	//{"level":"info","type":"log","message":"user login","node_id":"123-abc","trace_id":"abc-123"}
	//{"level":"info","type":"metric","name":"logins","value":1}
	//{"level":"info","type":"metric","name":"logins","value":1,"node_id":"123-abc","trace_id":"abc-123"}
	//{"level":"info","type":"trace","span_id":"def","name":"login","duration":150}
	//{"level":"info","type":"trace","span_id":"def","name":"login","duration":150,"node_id":"123-abc","trace_id":"abc-123"}
}
func ExampleTelemetry_field() {
	buf := &bytes.Buffer{}
	telemetry := ulog.NewTelemetry(
		ulog.WithFormat(ulog.FormatJson),
		ulog.WithMode(ulog.ModeSync, buf),
	)
	defer telemetry.Close()
	telemetry.Info(ulog.DataLog,
		ulog.Bool("bool", true),
		ulog.Bools("bools", []bool{true, false}),
		ulog.Duration("duration", 5*time.Second),
		ulog.Durations("durations", []time.Duration{5 * time.Second, 10 * time.Second}),
		ulog.Error(fmt.Errorf("err")),
		ulog.Errors([]error{fmt.Errorf("err1"), fmt.Errorf("err2")}),
		ulog.Float64("float64", 3.14159),
		ulog.Floats64("floats64", []float64{1.5, 2.5}),
		ulog.Int("int", 42),
		ulog.Ints("ints", []int{10, 20, 30}),
		ulog.Int64("int64", 1234567890),
		ulog.Ints64("ints64", []int64{1234567890, 9876543210}),
		ulog.String("string", "str"),
		ulog.Strings("strings", []string{"str1", "str2", "str3"}),
		ulog.Time("time", time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC)),
		ulog.Times("times", []time.Time{time.Date(2026, 4, 22, 12, 0, 0, 0, time.UTC), time.Date(2025, 4, 22, 12, 0, 0, 0, time.UTC)}),
	)
	telemetry.Sync()
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	//{"level":"info","type":"log","bool":true,"bools":[true,false],"duration":"5s","durations":["5s","10s"],"error":"err","errors":["err1","err2"],"float64":3.14159,"floats64":[1.5,2.5],"int":42,"ints":[10,20,30],"int64":1234567890,"ints64":[1234567890,9876543210],"string":"str","strings":["str1","str2","str3"],"time":"2026-04-22T12:00:00.000000+00:00","times":["2026-04-22T12:00:00.000000+00:00","2025-04-22T12:00:00.000000+00:00"]}
}
func ExampleTelemetry_format() {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	telemetry := ulog.NewTelemetry(
		ulog.WithExtractor("node_id", "trace_id"),
		ulog.WithFormat(ulog.FormatText),
		ulog.WithMode(ulog.ModeSync, buf),
	)
	defer telemetry.Close()
	telemetry.Info(ulog.DataLog, ulog.String("message", "test info text"))
	telemetry.Sync()
	telemetry.SetFormat(ulog.FormatJson)
	telemetry.InfoWithContext(ctx, ulog.DataLog, ulog.String("message", "test info text"))
	telemetry.Sync()
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	// [INFO] type="log" message="test info text"
	// {"level":"info","type":"log","message":"test info text","node_id":"123-abc","trace_id":"abc-123"}
}
func ExampleTelemetry_level() {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	telemetry := ulog.NewTelemetry(
		ulog.WithExtractor("node_id", "trace_id"),
		ulog.WithLevel(ulog.LevelDebug),
		ulog.WithMode(ulog.ModeSync, buf),
	)
	defer telemetry.Close()
	telemetry.DebugWithContext(ctx, ulog.DataLog, ulog.String("message", "debug test text"))
	telemetry.ErrorWithContext(ctx, ulog.DataLog, ulog.String("message", "error test text"))
	telemetry.InfoWithContext(ctx, ulog.DataLog, ulog.String("message", "info test text"))
	telemetry.WarnWithContext(ctx, ulog.DataLog, ulog.String("message", "warn test text"))
	telemetry.Sync()
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	//{"level":"debug","type":"log","message":"debug test text","node_id":"123-abc","trace_id":"abc-123"}
	//{"level":"error","type":"log","message":"error test text","node_id":"123-abc","trace_id":"abc-123"}
	//{"level":"info","type":"log","message":"info test text","node_id":"123-abc","trace_id":"abc-123"}
	//{"level":"warn","type":"log","message":"warn test text","node_id":"123-abc","trace_id":"abc-123"}
}
func ExampleTelemetry_mode() {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	telemetry := ulog.NewTelemetry(
		ulog.WithExtractor("node_id", "trace_id"),
		ulog.WithMode(ulog.ModeAsync, buf, 1000),
	)
	defer telemetry.Close()
	telemetry.Info(ulog.DataLog, ulog.String("message", "test info text"))
	telemetry.Sync()
	telemetry.SetMode(ulog.ModeSync, buf)
	telemetry.InfoWithContext(ctx, ulog.DataLog, ulog.String("message", "test info text"))
	telemetry.Sync()
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	// {"level":"info","type":"log","message":"test info text"}
	// {"level":"info","type":"log","message":"test info text","node_id":"123-abc","trace_id":"abc-123"}
}

// Вспомогательные функции
func formatOutput(str string) string {
	lines := strings.Split(str, "\n")
	var result []string
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	for _, line := range lines {
		if line == "" {
			continue
		}
		line = ansiRegex.ReplaceAllString(line, "")
		if strings.HasPrefix(line, "{") {
			re := regexp.MustCompile(`"(?:timestamp|caller)":"[^"]*",?`)
			line = re.ReplaceAllString(line, "")
			result = append(result, line)
			continue
		}
		bracketStart := strings.Index(line, "[")
		if bracketStart == -1 {
			result = append(result, line)
			continue
		}
		bracketEnd := strings.Index(line[bracketStart:], "]")
		if bracketEnd == -1 {
			result = append(result, line[bracketStart:])
			continue
		}
		level := line[bracketStart : bracketStart+bracketEnd+1]
		remaining := line[bracketStart+bracketEnd+1:]
		goIdx := strings.Index(remaining, ".go:")
		if goIdx != -1 {
			afterGo := remaining[goIdx+4:]
			spaceIdx := strings.Index(afterGo, " ")
			if spaceIdx != -1 {
				remaining = afterGo[spaceIdx+1:]
			} else {
				remaining = afterGo
			}
		} else {
			remaining = strings.TrimSpace(remaining)
		}
		result = append(result, level+" "+remaining)
	}
	output := strings.Join(result, "\n")
	output = strings.ReplaceAll(output, "}{", "}\n{")
	return output
}
