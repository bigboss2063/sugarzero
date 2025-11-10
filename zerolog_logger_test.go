package loggerv2_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/bigboss2063/loggerv2"
)

var testWriter bytes.Buffer

func TestMain(m *testing.M) {
	if _, err := loggerv2.New(context.Background(), "debug", &testWriter); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type logEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func TestLoggerWithContext(t *testing.T) {
	// 初始�?logger 并注入上下文
	ctx, err := loggerv2.New(context.Background(), "debug")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 添加字段到上下文
	ctx = loggerv2.WithFields(ctx,
		"request_id", "req-123",
		"user_id", 456,
	)

	// 使用包级函数直接输出日志（会包含字段�?	loggerv2.Info(ctx, "Processing request")
	loggerv2.Infof(ctx, "User %s logged in", "john")
	loggerv2.Infoln(ctx, "Request", "completed")
}

func TestLoggerWithoutContext(t *testing.T) {
	// 初始�?logger 并注入上下文
	ctx, err := loggerv2.New(context.Background(), "info")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 使用包级函数输出日志
	loggerv2.Info(ctx, "This works with context from New()")
	loggerv2.Warnf(ctx, "Warning: %s", "using package-level functions")

	// 使用空上下文（没�?logger，日志会被丢弃）
	emptyCtx := context.Background()
	loggerv2.Info(emptyCtx, "This will be silently dropped")
}

func TestPackageLoggerWarnsWhenContextMissingLogger(t *testing.T) {
	testWriter.Reset()

	loggerv2.Info(context.Background(), "no logger in context")

	lines := strings.Split(strings.TrimSpace(testWriter.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("expected log output")
	}

	var entry logEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}

	if strings.ToUpper(entry.Level) != "WARN" {
		t.Fatalf("expected WARN level, got %s", entry.Level)
	}

	if entry.Message != "context does not contain a logger, using fallback logger" {
		t.Fatalf("unexpected warning message: %s", entry.Message)
	}

	if got := loggerv2.GetLogLevel(context.Background()); got != "debug" {
		t.Fatalf("logger level changed unexpectedly: %s", got)
	}
}

func TestLoggerWithAdditionalFields(t *testing.T) {
	ctx, err := loggerv2.New(context.Background(), "debug")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// 第一层字段
	ctx = loggerv2.WithField(ctx, "service", "api-gateway")
	loggerv2.Debug(ctx, "Service initialized")

	// 添加更多字段（会合并）
	ctx = loggerv2.WithFields(ctx,
		"endpoint", "/api/users",
		"method", "GET",
	)
	loggerv2.Info(ctx, "Handling request")

	// 验证字段提取
	fields := loggerv2.FieldsFromContext(ctx)
	if fields == nil {
		t.Fatal("Expected fields in context")
	}
	if fields["service"] != "api-gateway" {
		t.Errorf("Expected service=api-gateway, got %v", fields["service"])
	}
}

func TestLogLevelChange(t *testing.T) {
	ctx, err := loggerv2.New(context.Background(), "info")
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	loggerv2.SetLogLevel(ctx, "info")
	if loggerv2.GetLogLevel(ctx) != "info" {
		t.Errorf("Expected level 'info', got '%s'", loggerv2.GetLogLevel(ctx))
	}

	loggerv2.SetLogLevel(ctx, "error")
	if loggerv2.GetLogLevel(ctx) != "error" {
		t.Errorf("Expected level 'error', got '%s'", loggerv2.GetLogLevel(ctx))
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	ctx, _ := loggerv2.New(context.Background(), "info")
	ctx = loggerv2.WithFields(ctx,
		"request_id", "bench-123",
		"user_id", 789,
	)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loggerv2.Info(ctx, "Benchmark message")
	}
}

func BenchmarkLoggerWithoutPreFormatting(b *testing.B) {
	ctx, _ := loggerv2.New(context.Background(), "info")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		loggerv2.Infof(ctx, "Message number: %d", i)
	}
}
