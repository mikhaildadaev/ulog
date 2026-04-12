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
func ExampleNewLogger() {
	buf := &bytes.Buffer{}
	ctx := context.WithValue(context.Background(), "trace_id", "abc-123")
	logger := ulog.NewLogger(
		ulog.WithFormat(ulog.FormatText),
		ulog.WithLevel(ulog.LevelDebug),
		ulog.WithMode(ulog.ModeAsync, buf, 1000),
		ulog.WithTheme(ulog.ThemeDark),
	)
	defer logger.Close()
	output := ""
	logger.Debug("test message")
	logger.Info("test message")
	logger.Warn("test message")
	logger.Error("test message")
	logger.Sync()
	logger.SetExtractor("trace_id")
	logger.SetFormat(ulog.FormatJson)
	logger.SetLevel(ulog.LevelDebug)
	logger.SetMode(ulog.ModeSync, buf)
	logger.DebugWithContext(ctx, "test message")
	logger.InfoWithContext(ctx, "test message")
	logger.WarnWithContext(ctx, "test message")
	logger.ErrorWithContext(ctx, "test message")
	output = formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	// [DEBUG] test message
	// [INFO] test message
	// [WARN] test message
	// [ERROR] test message
	// {"level":"debug","message":"test message","trace_id":"abc-123"}
	// {"level":"info","message":"test message","trace_id":"abc-123"}
	// {"level":"warn","message":"test message","trace_id":"abc-123"}
	// {"level":"error","message":"test message","trace_id":"abc-123"}
}
func ExampleNewLoggerLog() {
	buf := &bytes.Buffer{}
	logger := ulog.NewLogger(
		ulog.WithMode(ulog.ModeSync, buf),
		ulog.WithFormat(ulog.FormatText),
		ulog.WithLevel(ulog.LevelError),
	)
	loggerLog := ulog.NewLoggerLog(ulog.LevelError, logger)
	loggerLog.Print("this will be logged as ERROR")
	loggerLog.Printf("user %s failed to login", "john")
	loggerLog.Println("another error message")
	output := formatOutput(buf.String())
	fmt.Print(output)
	// Output:
	// [ERROR] this will be logged as ERROR
	// [ERROR] user john failed to login
	// [ERROR] another error message
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
