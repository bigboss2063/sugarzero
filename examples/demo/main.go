package main

import (
	"context"

	"github.com/bigboss2063/sugarzero"
)

func main() {
	ctx, err := sugarzero.New(context.Background(), "debug")
	if err != nil {
		panic(err)
	}

	ctx = sugarzero.WithFields(ctx,
		"service", "test-app",
		"version", "2.0.0",
	)

	sugarzero.Debug(ctx, "This is a DEBUG message")
	sugarzero.Info(ctx, "This is an INFO message")
	sugarzero.Warn(ctx, "This is a WARN message")
	sugarzero.Error(ctx, "This is an ERROR message")

	ctx = sugarzero.WithFields(ctx, "user_id", 12345, "action", "login")
	sugarzero.Infof(ctx, "User %s logged in successfully", "Alice")
}
