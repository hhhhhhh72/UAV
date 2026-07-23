// Package logger provides structured logging with console + daily file output,
// replicating the JS Logger class from the reference project (logger.js).
package logger

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	logDir string
	mu     sync.Mutex
	// level mapping: error=0, warn=1, info=2, debug=3
	currentLevel int
)

// Init configures the logger with console + optional file output.
func Init(env string) {
	level := slog.LevelInfo
	currentLevel = 2
	if env == "development" || env == "dev" {
		level = slog.LevelDebug
		currentLevel = 3
	}
	if d := os.Getenv("LOG_DIR"); d != "" {
		logDir = d
	} else {
		logDir = "./logs"
	}
	if lv := os.Getenv("LOG_LEVEL"); lv != "" {
		switch lv {
		case "error":
			currentLevel = 0
			level = slog.LevelError
		case "warn":
			currentLevel = 1
			level = slog.LevelWarn
		case "info":
			currentLevel = 2
			level = slog.LevelInfo
		case "debug":
			currentLevel = 3
			level = slog.LevelDebug
		}
	}
	// Console handler
	opts := &slog.HandlerOptions{Level: level}
	var handler slog.Handler
	if env == "production" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}
	slog.SetDefault(slog.New(handler))
	// Ensure log directory exists
	if logDir != "" {
		os.MkdirAll(logDir, 0755)
	}
}

// writeFile appends a line to today's log file (YYYY-MM-DD.log).
func writeFile(level, msg string) {
	if logDir == "" {
		return
	}
	mu.Lock()
	defer mu.Unlock()
	dateStr := time.Now().Format("2006-01-02")
	logPath := filepath.Join(logDir, dateStr+".log")
	f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	timestamp := time.Now().Format(time.RFC3339)
	line := fmt.Sprintf("[%s] [%s] %s\n", timestamp, level, msg)
	io.WriteString(f, line)
}

func logf(level string, lvl int, msg string, args ...any) {
	if lvl > currentLevel {
		return
	}
	// Format slog-style key-value pairs: "msg" "key1", val1, "key2", val2 → "msg key1=val1 key2=val2"
	formatted := msg
	for i := 0; i+1 < len(args); i += 2 {
		k := fmt.Sprint(args[i])
		v := fmt.Sprint(args[i+1])
		formatted += " " + k + "=" + v
	}
	// Handle odd number of args (last arg as value with key "?")
	if len(args)%2 != 0 {
		formatted += " ?=" + fmt.Sprint(args[len(args)-1])
	}
	writeFile(level, formatted)
}

// Info logs at info level (level 2).
func Info(msg string, args ...any) {
	slog.Info(msg, args...)
	logf("INFO", 2, msg, args...)
}

// Warn logs at warn level (level 1).
func Warn(msg string, args ...any) {
	slog.Warn(msg, args...)
	logf("WARN", 1, msg, args...)
}

// Error logs at error level (level 0).
func Error(msg string, args ...any) {
	slog.Error(msg, args...)
	logf("ERROR", 0, msg, args...)
}

// Debug logs at debug level (level 3).
func Debug(msg string, args ...any) {
	slog.Debug(msg, args...)
	logf("DEBUG", 3, msg, args...)
}

// Fatal logs at error level and exits.
func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	logf("FATAL", 0, msg, args...)
	os.Exit(1)
}

// RequestLog logs an HTTP request with structured metadata.
func RequestLog(method, path, ip string, statusCode int, duration time.Duration) {
	msg := fmt.Sprintf("%s %s status=%d duration=%dms ip=%s", method, path, statusCode, duration.Milliseconds(), ip)
	slog.Info("request", "method", method, "path", path, "ip", ip, "status", statusCode, "duration_ms", duration.Milliseconds())
	writeFile("INFO", msg)
}

// ErrorWithContext logs an error with request context (like JS errorWithContext).
func ErrorWithContext(err error, method, path, ip, body string) {
	msg := fmt.Sprintf("ERROR %s %s ip=%s err=%v", method, path, ip, err)
	if len(body) > 500 {
		body = body[:500]
	}
	if body != "" {
		msg += " body=" + body
	}
	slog.Error("request_error", "method", method, "path", path, "ip", ip, "error", err)
	writeFile("ERROR", msg)
}
