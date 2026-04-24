package ulog_test

import (
	"bytes"
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/mikhaildadaev/ulog"
)

// Примеры использования публичных конструкторов
func ExampleNewTelemetry() {
	buf := &bytes.Buffer{}
	ctx := context.Background()
	ctx = context.WithValue(ctx, "node_id", "123-abc")
	ctx = context.WithValue(ctx, "trace_id", "abc-123")
	telemetry := ulog.NewTelemetry(
		ulog.WithFormat(ulog.FormatText),
		ulog.WithLevel(ulog.LevelDebug),
		ulog.WithMode(ulog.ModeAsync, buf, 1000),
		ulog.WithTheme(ulog.ThemeDark),
	)
	defer telemetry.Close()
	telemetry.Debug(ulog.DataLog, ulog.String("message", "test debug text"))
	telemetry.Info(ulog.DataLog, ulog.String("message", "test info text"))
	telemetry.Warn(ulog.DataLog, ulog.String("message", "test warn text"))
	telemetry.Error(ulog.DataLog, ulog.String("message", "test error text"))
	telemetry.Sync()
	telemetry.SetExtractor("node_id", "trace_id")
	telemetry.SetFormat(ulog.FormatJson)
	telemetry.SetLevel(ulog.LevelDebug)
	telemetry.SetMode(ulog.ModeSync, buf)
	telemetry.DebugWithContext(ctx, ulog.DataLog, ulog.String("message", "test debug text"))
	telemetry.InfoWithContext(ctx, ulog.DataLog, ulog.String("message", "test info text"))
	telemetry.WarnWithContext(ctx, ulog.DataLog, ulog.String("message", "test warn text"))
	telemetry.ErrorWithContext(ctx, ulog.DataLog, ulog.String("message", "test error text"))
	telemetry.Sync()
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	// [DEBUG] type="log" message="test debug text"
	// [INFO] type="log" message="test info text"
	// [WARN] type="log" message="test warn text"
	// [ERROR] type="log" message="test error text"
	// {"level":"debug","type":"log","message":"test debug text","node_id":"123-abc","trace_id":"abc-123"}
	// {"level":"info","type":"log","message":"test info text","node_id":"123-abc","trace_id":"abc-123"}
	// {"level":"warn","type":"log","message":"test warn text","node_id":"123-abc","trace_id":"abc-123"}
	// {"level":"error","type":"log","message":"test error text","node_id":"123-abc","trace_id":"abc-123"}
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
			re := regexp.MustCompile(`"(?:time|caller)":"[^"]*",?`)
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
