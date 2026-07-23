package logger_test

import (
	"errors"
	"testing"
	"time"

	"drone-platform/internal/logger"
)

func TestInitDev(t *testing.T) {
	logger.Init("development")
}

func TestInitProd(t *testing.T) {
	logger.Init("production")
}

func TestInfo(t *testing.T) {
	logger.Init("development")
	logger.Info("test info", "key", "value")
}

func TestWarn(t *testing.T) {
	logger.Init("development")
	logger.Warn("test warn", "key", "value")
}

func TestError(t *testing.T) {
	logger.Init("development")
	logger.Error("test error", "key", "value")
}

func TestDebug(t *testing.T) {
	logger.Init("development")
	logger.Debug("test debug", "key", "value")
}

func TestRequestLog(t *testing.T) {
	logger.Init("development")
	logger.RequestLog("GET", "/api/test", "127.0.0.1", 200, 10*time.Millisecond)
}

func TestErrorWithContext(t *testing.T) {
	logger.Init("development")
	logger.ErrorWithContext(errors.New("boom"), "POST", "/api/test", "127.0.0.1", "body content")
}
