# LoggerV2

基于 [zerolog](https://github.com/rs/zerolog) 的轻量级上下文日志库，封装了包级别的日志函数并提供 `Logger` 接口实现，便于在全局或局部动态调整日志级别。

## 特性

- ✅ **上下文驱动**：`logger` 与字段都通过 `context.Context` 传递
- ✅ **包级函数**：直接调用 `loggerv2.Info(ctx, ...)`，无需显式持有实例
- ✅ **字段累积**：多层 `WithFields/WithField` 自动合并
- ✅ **线程安全**：调整日志级别时加锁，日志写入无数据竞争
- ✅ **可扩展输出**：支持传入任意 `io.Writer`，默认使用 `os.Stdout`

## 安装

```bash
go get github.com/bigboss2063/loggerv2
```

## 快速开始

```go
package main

import (
	"context"
	"github.com/bigboss2063/loggerv2"
)

func main() {
	ctx, logger, err := loggerv2.New(context.Background(), "info")
	if err != nil {
		panic(err)
	}

	ctx = loggerv2.WithFields(ctx,
		"service", "api-gateway",
		"version", "1.0.0",
	)

	loggerv2.Info(ctx, "application started")
	loggerv2.Infof(ctx, "listening on %d", 8080)

	logger.SetLogLevel("debug")
	loggerv2.Debug(ctx, "debug logs are now visible")
}
```

### 字段累积

```go
ctx = loggerv2.WithField(ctx, "service", "order-api")
ctx = loggerv2.WithFields(ctx,
	"request_id", "req-12345",
	"user_id",    789,
)

loggerv2.Info(ctx, "processing request") // 会自动携带所有字段
```

### HTTP 处理示例

```go
func (s *Server) HandleRequest(w http.ResponseWriter, r *http.Request) {
	ctx := loggerv2.WithField(r.Context(), "request_id", uuid.New().String())
	loggerv2.Info(ctx, "request started")

	ctx = loggerv2.WithField(ctx, "user_id", getUserID(r))
	result := s.process(ctx)

	loggerv2.Infof(ctx, "request completed: %s", result)
}
```

## API 概览

### 初始化

```go
ctx, logger, err := loggerv2.New(ctx context.Context, level string, writers ...io.Writer)
```

- `ctx`: 父上下文（可为 `nil`）
- `level`: 日志级别（debug/info/warn/error/fatal）
- `writers`: 可选输出目标（为空时使用 `os.Stdout`）

返回值：
1. 注入了 `Logger` 的新 `context.Context`
2. `Logger` 接口实现（`*ZeroLogger`）
3. 错误（级别无效时）

### 上下文字段

- `WithFields(ctx, keyvals ...any) context.Context`：使用成对的 `key, value` 追加字段
- `WithField(ctx, key string, value any) context.Context`：追加单个字段
- `FieldsFromContext(ctx context.Context) map[string]any`：读取当前字段（返回副本，可安全修改）

### 包级日志函数

所有函数均接受 `context.Context` 为首参，并自动读取其中的 `Logger`：

```
Debug | Debugf | Debugln
Info  | Infof  | Infoln
Warn  | Warnf  | Warnln
Error | Errorf | Errorln
Fatal | Fatalf | Fatalln
```

当上下文中没有注入 logger 时，这些函数会静默忽略调用，避免 panic。

### Logger 接口

```go
type Logger interface {
	Debug(context.Context, ...any)
	Debugf(context.Context, string, ...any)
	...
	SetLogLevel(string)
	GetLogLevel() string
}
```

`ZeroLogger` 是该接口的默认实现：

- `SetLogLevel`：线程安全地动态调整日志级别
- `GetLogLevel`：读取当前级别

## 最佳实践

1. **应用启动时初始化一次**，通过 context 传递
2. **在请求链路中不断追加字段**，便于排查问题
3. **仅在需要时提升日志级别**，其他时间保持低噪声

## 开发与测试

```bash
go test ./...
```

可选基准测试（示例在 `example_test.go`）：

```bash
go test -bench=. -benchmem
```

## 许可

MIT
