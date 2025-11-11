package sugarzero_test

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/bigboss2063/sugarzero"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// setupTest creates a fresh logger for each test with isolated state
func setupTest(t *testing.T, level string) (context.Context, *bytes.Buffer) {
	t.Helper()

	// Reset global logger state before each test
	sugarzero.Reset()

	var buf bytes.Buffer
	ctx, err := sugarzero.New(context.Background(), level, &buf)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	// Cleanup after test
	t.Cleanup(func() {
		sugarzero.Reset()
	})

	return ctx, &buf
}

func TestMain(m *testing.M) {
	// No global setup needed anymore
	os.Exit(m.Run())
}

type logEntry struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

func readLogEntry(t *testing.T, buf *bytes.Buffer, index ...int) map[string]any {
	t.Helper()

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	if len(lines) == 0 {
		t.Fatal("expected log output")
	}

	target := len(lines) - 1
	if len(index) > 0 {
		target = index[0]
		if target < 0 {
			target = len(lines) + target
		}
		if target < 0 || target >= len(lines) {
			t.Fatalf("requested log index %d out of range (total %d)", index[0], len(lines))
		}
	}

	var entry map[string]any
	if err := json.Unmarshal([]byte(lines[target]), &entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}

	return entry
}

func TestLoggerWithContext(t *testing.T) {
	ctx, testWriter := setupTest(t, "debug")

	// 添加字段到上下文
	ctx = sugarzero.WithFields(ctx,
		"request_id", "req-123",
		"user_id", 456,
	)

	sugarzero.Infof(ctx, "User %s logged in", "john")

	entry := readLogEntry(t, testWriter)

	if strings.ToUpper(entry["level"].(string)) != "INFO" {
		t.Fatalf("expected INFO level, got %s", entry["level"])
	}

	if entry["message"].(string) != "User john logged in" {
		t.Fatalf("unexpected message: %s", entry["message"])
	}

	if entry["request_id"].(string) != "req-123" {
		t.Fatalf("expected request_id=req-123, got %v", entry["request_id"])
	}

	if int(entry["user_id"].(float64)) != 456 {
		t.Fatalf("expected user_id=456, got %v", entry["user_id"])
	}

	testWriter.Reset()
	sugarzero.Infoln(ctx, "Request", "completed")

	entry = readLogEntry(t, testWriter)

	if strings.ToUpper(entry["level"].(string)) != "INFO" {
		t.Fatalf("expected INFO level, got %s", entry["level"])
	}

	if !strings.Contains(entry["message"].(string), "Request") {
		t.Fatalf("unexpected message: %s", entry["message"])
	}
}

func TestPackageLoggerWarnsWhenContextMissingLogger(t *testing.T) {
	_, testWriter := setupTest(t, "debug")

	sugarzero.Info(context.Background(), "no logger in context")

	entry := readLogEntry(t, testWriter, 0)

	if strings.ToUpper(entry["level"].(string)) != "WARN" {
		t.Fatalf("expected WARN level, got %s", entry["level"])
	}

	if entry["message"].(string) != "context does not contain a logger, using fallback logger" {
		t.Fatalf("unexpected warning message: %s", entry["message"])
	}

	if got := sugarzero.GetLogLevel(context.Background()); got != "debug" {
		t.Fatalf("logger level changed unexpectedly: %s", got)
	}
}

func TestLoggerWithAdditionalFields(t *testing.T) {
	ctx, testWriter := setupTest(t, "debug")

	// 第一层字段
	ctx = sugarzero.WithField(ctx, "service", "api-gateway")
	sugarzero.Debug(ctx, "Service initialized")

	entry := readLogEntry(t, testWriter)

	if strings.ToUpper(entry["level"].(string)) != "DEBUG" {
		t.Fatalf("expected DEBUG level, got %s", entry["level"])
	}

	if entry["message"].(string) != "Service initialized" {
		t.Fatalf("unexpected message: %s", entry["message"])
	}

	if entry["service"].(string) != "api-gateway" {
		t.Fatalf("expected service=api-gateway, got %v", entry["service"])
	}

	// 添加更多字段（会合并）
	testWriter.Reset()
	ctx = sugarzero.WithFields(ctx,
		"endpoint", "/api/users",
		"method", "GET",
	)
	sugarzero.Info(ctx, "Handling request")

	entry = readLogEntry(t, testWriter)

	if strings.ToUpper(entry["level"].(string)) != "INFO" {
		t.Fatalf("expected INFO level, got %s", entry["level"])
	}

	if entry["message"].(string) != "Handling request" {
		t.Fatalf("unexpected message: %s", entry["message"])
	}

	if entry["service"].(string) != "api-gateway" {
		t.Fatalf("expected service=api-gateway, got %v", entry["service"])
	}

	if entry["endpoint"].(string) != "/api/users" {
		t.Fatalf("expected endpoint=/api/users, got %v", entry["endpoint"])
	}

	if entry["method"].(string) != "GET" {
		t.Fatalf("expected method=GET, got %v", entry["method"])
	}

	// 验证字段提取
	fields := sugarzero.FieldsFromContext(ctx)
	if fields == nil {
		t.Fatal("Expected fields in context")
	}
	if fields["service"] != "api-gateway" {
		t.Errorf("Expected service=api-gateway, got %v", fields["service"])
	}
}

func TestLogLevelChange(t *testing.T) {
	ctx, testWriter := setupTest(t, "info")

	sugarzero.SetLogLevel(ctx, "info")
	if sugarzero.GetLogLevel(ctx) != "info" {
		t.Errorf("Expected level 'info', got '%s'", sugarzero.GetLogLevel(ctx))
	}

	// 测试 info 级别，debug 应该不输出
	sugarzero.Debug(ctx, "This should not appear")
	if strings.TrimSpace(testWriter.String()) != "" {
		t.Errorf("Debug message should not appear at info level")
	}

	// 测试 info 级别，info 应该输出
	testWriter.Reset()
	sugarzero.Info(ctx, "This should appear")

	entry := readLogEntry(t, testWriter)

	if strings.ToUpper(entry["level"].(string)) != "INFO" {
		t.Fatalf("expected INFO level, got %s", entry["level"])
	}

	// 修改为 error 级别
	sugarzero.SetLogLevel(ctx, "error")
	if sugarzero.GetLogLevel(ctx) != "error" {
		t.Errorf("Expected level 'error', got '%s'", sugarzero.GetLogLevel(ctx))
	}

	// 测试 error 级别，info 应该不输出
	testWriter.Reset()
	sugarzero.Info(ctx, "This should not appear at error level")
	if strings.TrimSpace(testWriter.String()) != "" {
		t.Errorf("Info message should not appear at error level")
	}

	// 测试 error 级别，error 应该输出
	testWriter.Reset()
	sugarzero.Error(ctx, "This is an error")

	entry = readLogEntry(t, testWriter)

	if strings.ToUpper(entry["level"].(string)) != "ERROR" {
		t.Fatalf("expected ERROR level, got %s", entry["level"])
	}

	if entry["message"].(string) != "This is an error" {
		t.Fatalf("unexpected message: %s", entry["message"])
	}
}

func TestAllLogLevels(t *testing.T) {
	ctx, testWriter := setupTest(t, "debug")

	tests := []struct {
		name     string
		logFunc  func(context.Context, ...any)
		message  string
		expected string
	}{
		{"Debug", sugarzero.Debug, "debug message", "DEBUG"},
		{"Info", sugarzero.Info, "info message", "INFO"},
		{"Warn", sugarzero.Warn, "warn message", "WARN"},
		{"Error", sugarzero.Error, "error message", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWriter.Reset()
			tt.logFunc(ctx, tt.message)

			entry := readLogEntry(t, testWriter)

			if strings.ToUpper(entry["level"].(string)) != tt.expected {
				t.Fatalf("expected %s level, got %s", tt.expected, entry["level"])
			}

			if entry["message"].(string) != tt.message {
				t.Fatalf("unexpected message: %s", entry["message"])
			}
		})
	}
}

func TestFormattedLogFunctions(t *testing.T) {
	ctx, testWriter := setupTest(t, "debug")

	tests := []struct {
		name     string
		logFunc  func(context.Context, string, ...any)
		format   string
		args     []any
		expected string
		level    string
	}{
		{"Debugf", sugarzero.Debugf, "user %s age %d", []any{"alice", 25}, "user alice age 25", "DEBUG"},
		{"Infof", sugarzero.Infof, "request %d processed", []any{123}, "request 123 processed", "INFO"},
		{"Warnf", sugarzero.Warnf, "warning: %s", []any{"low memory"}, "warning: low memory", "WARN"},
		{"Errorf", sugarzero.Errorf, "error code: %d", []any{500}, "error code: 500", "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWriter.Reset()
			tt.logFunc(ctx, tt.format, tt.args...)

			entry := readLogEntry(t, testWriter)

			if strings.ToUpper(entry["level"].(string)) != tt.level {
				t.Fatalf("expected %s level, got %s", tt.level, entry["level"])
			}

			if entry["message"].(string) != tt.expected {
				t.Fatalf("expected message '%s', got '%s'", tt.expected, entry["message"])
			}
		})
	}
}

func TestLogFunctionsWithLn(t *testing.T) {
	ctx, testWriter := setupTest(t, "debug")

	tests := []struct {
		name    string
		logFunc func(context.Context, ...any)
		args    []any
		level   string
	}{
		{"Debugln", sugarzero.Debugln, []any{"debug", "message"}, "DEBUG"},
		{"Infoln", sugarzero.Infoln, []any{"info", "message"}, "INFO"},
		{"Warnln", sugarzero.Warnln, []any{"warn", "message"}, "WARN"},
		{"Errorln", sugarzero.Errorln, []any{"error", "message"}, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testWriter.Reset()
			tt.logFunc(ctx, tt.args...)

			entry := readLogEntry(t, testWriter)

			if strings.ToUpper(entry["level"].(string)) != tt.level {
				t.Fatalf("expected %s level, got %s", tt.level, entry["level"])
			}

			message := entry["message"].(string)
			for _, arg := range tt.args {
				if !strings.Contains(message, arg.(string)) {
					t.Fatalf("expected message to contain '%v', got '%s'", arg, message)
				}
			}
		})
	}
}

func TestNilContext(t *testing.T) {
	_, testWriter := setupTest(t, "debug")

	sugarzero.Info(nil, "nil context message")

	// 至少应包含警告和 Info
	_ = readLogEntry(t, testWriter, 1)

	entry := readLogEntry(t, testWriter, 0)

	if strings.ToUpper(entry["level"].(string)) != "WARN" {
		t.Fatalf("expected WARN level for missing logger, got %s", entry["level"])
	}
}

func TestInvalidLogLevel(t *testing.T) {
	sugarzero.Reset()

	var testWriter bytes.Buffer
	_, err := sugarzero.New(context.Background(), "invalid_level", &testWriter)
	if err == nil {
		t.Fatal("expected error for invalid log level")
	}

	if !strings.Contains(err.Error(), "invalid log level") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestEmptyFields(t *testing.T) {
	ctx, testWriter := setupTest(t, "debug")

	// 测试空 key
	ctx = sugarzero.WithField(ctx, "", "value")
	sugarzero.Info(ctx, "message with empty key")

	entry := readLogEntry(t, testWriter)

	// 空 key 应该被忽略
	if _, exists := entry[""]; exists {
		t.Fatal("empty key should not be added to log entry")
	}
}

func TestOddNumberOfFieldsIgnoresLastValue(t *testing.T) {
	ctx, testWriter := setupTest(t, "debug")

	// 奇数个参数，最后一个应该被忽略
	ctx = sugarzero.WithFields(ctx, "key1", "value1", "key2")
	sugarzero.Info(ctx, "message with odd fields")

	entry := readLogEntry(t, testWriter)

	if entry["key1"].(string) != "value1" {
		t.Fatalf("expected key1=value1, got %v", entry["key1"])
	}

	// key2 应该不存在，因为没有对应的值
	if _, exists := entry["key2"]; exists {
		t.Fatal("key2 should not exist without a corresponding value")
	}
}

func TestFieldsMerging(t *testing.T) {
	ctx, testWriter := setupTest(t, "debug")

	// 第一次添加字段
	ctx = sugarzero.WithField(ctx, "field1", "value1")
	// 第二次添加字段
	ctx = sugarzero.WithField(ctx, "field2", "value2")
	// 第三次添加多个字段
	ctx = sugarzero.WithFields(ctx, "field3", "value3", "field4", "value4")

	sugarzero.Info(ctx, "message with merged fields")

	entry := readLogEntry(t, testWriter)

	// 验证所有字段都存在
	expectedFields := map[string]string{
		"field1": "value1",
		"field2": "value2",
		"field3": "value3",
		"field4": "value4",
	}

	for key, expectedValue := range expectedFields {
		if entry[key].(string) != expectedValue {
			t.Fatalf("expected %s=%s, got %v", key, expectedValue, entry[key])
		}
	}
}

func TestMultipleWriters(t *testing.T) {
	sugarzero.Reset()

	var writer1, writer2 bytes.Buffer

	ctx, err := sugarzero.New(context.Background(), "info", &writer1, &writer2)
	if err != nil {
		t.Fatalf("Failed to create logger: %v", err)
	}

	t.Cleanup(func() {
		sugarzero.Reset()
	})

	sugarzero.Info(ctx, "message to multiple writers")

	// 验证两个 writer 都收到了日志
	if writer1.Len() == 0 {
		t.Fatal("writer1 should have received log output")
	}

	if writer2.Len() == 0 {
		t.Fatal("writer2 should have received log output")
	}

	// 验证内容相同
	if writer1.String() != writer2.String() {
		t.Fatal("both writers should have the same content")
	}

	var entry map[string]any
	if err := json.Unmarshal(writer1.Bytes(), &entry); err != nil {
		t.Fatalf("failed to decode log entry: %v", err)
	}

	if entry["message"].(string) != "message to multiple writers" {
		t.Fatalf("unexpected message: %s", entry["message"])
	}
}

func TestGetLogLevelWithNoLogger(t *testing.T) {
	setupTest(t, "debug")

	// 测试在没有 logger 的 context 中获取日志级别
	level := sugarzero.GetLogLevel(context.Background())
	if level == "" {
		t.Fatal("expected non-empty log level from global logger")
	}
}

func TestLoggerAutomaticallyInjectsTraceMetadata(t *testing.T) {
	ctx, testWriter := setupTest(t, "info")

	tp := sdktrace.NewTracerProvider()
	t.Cleanup(func() {
		_ = tp.Shutdown(context.Background())
	})

	tracer := tp.Tracer("test-tracer")
	ctx, span := tracer.Start(ctx, "traceable-operation")
	defer span.End()

	sugarzero.Info(ctx, "message with trace metadata")

	entry := readLogEntry(t, testWriter)
	spanCtx := span.SpanContext()

	traceID, ok := entry["trace_id"].(string)
	if !ok || traceID != spanCtx.TraceID().String() {
		t.Fatalf("expected trace_id %s, got %v", spanCtx.TraceID().String(), entry["trace_id"])
	}

	spanID, ok := entry["span_id"].(string)
	if !ok || spanID != spanCtx.SpanID().String() {
		t.Fatalf("expected span_id %s, got %v", spanCtx.SpanID().String(), entry["span_id"])
	}
}

func TestLoggerOmitsTraceMetadataWithoutSpan(t *testing.T) {
	ctx, testWriter := setupTest(t, "info")

	sugarzero.Info(ctx, "message without trace metadata")

	entry := readLogEntry(t, testWriter)

	if _, ok := entry["trace_id"]; ok {
		t.Fatal("did not expect trace_id in log entry")
	}

	if _, ok := entry["span_id"]; ok {
		t.Fatal("did not expect span_id in log entry")
	}
}

func TestFieldsFromEmptyContext(t *testing.T) {
	ctx := context.Background()
	fields := sugarzero.FieldsFromContext(ctx)
	if fields != nil {
		t.Fatal("expected nil fields from empty context")
	}
}

func BenchmarkLoggerWithFields(b *testing.B) {
	sugarzero.Reset()
	ctx, _ := sugarzero.New(context.Background(), "info")
	ctx = sugarzero.WithFields(ctx,
		"request_id", "bench-123",
		"user_id", 789,
	)

	b.Cleanup(func() {
		sugarzero.Reset()
	})

	for b.Loop() {
		sugarzero.Info(ctx, "Benchmark message")
	}
}

func BenchmarkLoggerWithoutPreFormatting(b *testing.B) {
	sugarzero.Reset()
	ctx, _ := sugarzero.New(context.Background(), "info")

	b.Cleanup(func() {
		sugarzero.Reset()
	})

	for i := 0; b.Loop(); i++ {
		sugarzero.Infof(ctx, "Message number: %d", i)
	}
}
