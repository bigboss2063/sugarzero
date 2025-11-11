package main

import (
	"context"
	"fmt"
	"time"

	"github.com/bigboss2063/sugarzero"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

func main() {
	// 1. 初始化 OpenTelemetry（此处使用最小化配置）
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName("sugarzero-otel-example"),
		)),
	)
	defer func() { _ = tp.Shutdown(context.Background()) }()

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.TraceContext{})

	// 2. 初始化 sugarzero 并注入 logger
	ctx, err := sugarzero.New(context.Background(), "info")
	if err != nil {
		panic(err)
	}

	ctx = sugarzero.WithFields(ctx,
		"env", "demo",
		"component", "order-service",
	)

	// 3. 模拟多次请求，日志会自动包含 trace_id/span_id
	for i := 1; i <= 2; i++ {
		processOrder(ctx, fmt.Sprintf("order-%03d", i))
	}
}

func processOrder(ctx context.Context, orderID string) {
	tracer := otel.Tracer("example/otel")

	ctx, span := tracer.Start(ctx, "processOrder")
	defer span.End()

	ctx = sugarzero.WithTracing(ctx) // 建议：显式同步 context，方便其他组件复用
	ctx = sugarzero.WithField(ctx, "order_id", orderID)

	sugarzero.Info(ctx, "order received")

	if err := chargePayment(ctx); err != nil {
		span.RecordError(err)
		span.SetAttributes(semconv.ExceptionStacktrace(err.Error()))
		sugarzero.Errorf(ctx, "order failed: %v", err)
		return
	}

	if err := dispatchPackage(ctx); err != nil {
		span.RecordError(err)
		sugarzero.Errorf(ctx, "dispatch failed: %v", err)
		return
	}

	sugarzero.Info(ctx, "order completed")
}

func chargePayment(ctx context.Context) error {
	tracer := otel.Tracer("example/otel")

	ctx, span := tracer.Start(ctx, "chargePayment")
	defer span.End()

	ctx = sugarzero.WithTracing(ctx)
	sugarzero.Debug(ctx, "charging payment gateway")

	time.Sleep(50 * time.Millisecond)
	return nil
}

func dispatchPackage(ctx context.Context) error {
	tracer := otel.Tracer("example/otel")

	ctx, span := tracer.Start(ctx, "dispatchPackage")
	defer span.End()

	ctx = sugarzero.WithTracing(ctx)
	sugarzero.Debug(ctx, "sending package to logistics system")

	time.Sleep(75 * time.Millisecond)
	return nil
}
