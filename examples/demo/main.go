package main

import (
	"context"

	"github.com/bigboss2063/loggerv2"
)

func main() {
	ctx, err := loggerv2.New(context.Background(), "debug")
	if err != nil {
		panic(err)
	}

	ctx = loggerv2.WithFields(ctx,
		"service", "test-app",
		"version", "2.0.0",
	)

	loggerv2.Debug(ctx, "This is a DEBUG message")
	loggerv2.Info(ctx, "This is an INFO message")
	loggerv2.Warn(ctx, "This is a WARN message")
	loggerv2.Error(ctx, "This is an ERROR message")

	ctx = loggerv2.WithFields(ctx, "user_id", 12345, "action", "login")
	loggerv2.Infof(ctx, "User %s logged in successfully", "Alice")
}
