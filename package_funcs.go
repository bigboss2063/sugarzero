package loggerv2

import "context"

// Debug logs a debug message using the logger from context.
// If no logger is found in context, the message is silently dropped.
func Debug(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Debug(ctx, args...)
	}
}

// Debugf logs a formatted debug message using the logger from context.
func Debugf(ctx context.Context, format string, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Debugf(ctx, format, args...)
	}
}

// Debugln logs a debug message using the logger from context.
func Debugln(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Debugln(ctx, args...)
	}
}

// Info logs an info message using the logger from context.
func Info(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Info(ctx, args...)
	}
}

// Infof logs a formatted info message using the logger from context.
func Infof(ctx context.Context, format string, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Infof(ctx, format, args...)
	}
}

// Infoln logs an info message using the logger from context.
func Infoln(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Infoln(ctx, args...)
	}
}

// Warn logs a warning message using the logger from context.
func Warn(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Warn(ctx, args...)
	}
}

// Warnf logs a formatted warning message using the logger from context.
func Warnf(ctx context.Context, format string, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Warnf(ctx, format, args...)
	}
}

// Warnln logs a warning message using the logger from context.
func Warnln(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Warnln(ctx, args...)
	}
}

// Error logs an error message using the logger from context.
func Error(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Error(ctx, args...)
	}
}

// Errorf logs a formatted error message using the logger from context.
func Errorf(ctx context.Context, format string, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Errorf(ctx, format, args...)
	}
}

// Errorln logs an error message using the logger from context.
func Errorln(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Errorln(ctx, args...)
	}
}

// Fatal logs a fatal message using the logger from context.
func Fatal(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Fatal(ctx, args...)
	}
}

// Fatalf logs a formatted fatal message using the logger from context.
func Fatalf(ctx context.Context, format string, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Fatalf(ctx, format, args...)
	}
}

// Fatalln logs a fatal message using the logger from context.
func Fatalln(ctx context.Context, args ...any) {
	if logger := getLoggerFromContext(ctx); logger != nil {
		logger.Fatalln(ctx, args...)
	}
}

// getLoggerFromContext retrieves the Logger from context.
func getLoggerFromContext(ctx context.Context) Logger {
	if ctx == nil {
		return nil
	}
	if logger, ok := ctx.Value(loggerKey).(Logger); ok && logger != nil {
		return logger
	}
	return nil
}
