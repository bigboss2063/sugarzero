package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bigboss2063/loggerv2"
)

func main() {
	// 1. 初始化 logger 并注入到上下文
	ctx, err := loggerv2.New(context.Background(), "debug")
	if err != nil {
		panic(err)
	}

	loggerv2.Infof(context.Background(), "no logger with context, will give an additional warning log")

	// 添加应用级别的字段
	ctx = loggerv2.WithFields(ctx,
		"app", "example-service",
		"version", "1.0.0",
		"env", "development",
	)

	// 2. 使用包级函数输出日志
	loggerv2.Info(ctx, "Application started")

	// 3. 模拟处理请求
	processRequest(ctx, "req-001", "john")
	processRequest(ctx, "req-002", "jane")

	// 4. 动态修改日志级别
	loggerv2.SetLogLevel(ctx, "error")
	loggerv2.Debug(ctx, "This debug message will not be shown")
	loggerv2.Error(ctx, "But errors are still shown")

	loggerv2.Info(ctx, "Application stopping")
}

// processRequest 模拟处理一个请求
func processRequest(ctx context.Context, requestID, username string) {
	// 为当前请求添加字段
	ctx = loggerv2.WithFields(ctx,
		"request_id", requestID,
		"username", username,
	)

	loggerv2.Infof(ctx, "Processing request for user: %s", username)

	// 模拟业务逻辑
	if err := doSomething(ctx); err != nil {
		loggerv2.Errorf(ctx, "Request failed: %v", err)
		return
	}

	loggerv2.Info(ctx, "Request completed successfully")
}

// doSomething 模拟业务逻辑
func doSomething(ctx context.Context) error {
	// 添加更多上下文信息
	ctx = loggerv2.WithField(ctx, "operation", "database_query")

	loggerv2.Debug(ctx, "Starting database operation")

	// 模拟耗时操作
	time.Sleep(100 * time.Millisecond)

	// 模拟错误
	if time.Now().UnixNano()%2 == 0 {
		return fmt.Errorf("simulated error")
	}

	loggerv2.Debug(ctx, "Database operation completed")
	return nil
}
