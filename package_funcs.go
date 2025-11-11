package sugarzero

import "context"

func Debug(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Debug(resolved, args...)
	})
}

func Debugf(ctx context.Context, format string, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Debugf(resolved, format, args...)
	})
}

func Debugln(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Debugln(resolved, args...)
	})
}

func Info(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Info(resolved, args...)
	})
}

func Infof(ctx context.Context, format string, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Infof(resolved, format, args...)
	})
}

func Infoln(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Infoln(resolved, args...)
	})
}

func Warn(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Warn(resolved, args...)
	})
}

func Warnf(ctx context.Context, format string, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Warnf(resolved, format, args...)
	})
}

func Warnln(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Warnln(resolved, args...)
	})
}

func Error(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Error(resolved, args...)
	})
}

func Errorf(ctx context.Context, format string, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Errorf(resolved, format, args...)
	})
}

func Errorln(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Errorln(resolved, args...)
	})
}

func Fatal(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Fatal(resolved, args...)
	})
}

func Fatalf(ctx context.Context, format string, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Fatalf(resolved, format, args...)
	})
}

func Fatalln(ctx context.Context, args ...any) {
	withLogger(ctx, func(logger *ZeroLogger, resolved context.Context) {
		logger.Fatalln(resolved, args...)
	})
}

func withLogger(ctx context.Context, fn func(*ZeroLogger, context.Context)) {
	if ctx == nil {
		ctx = context.Background()
	}
	if logger := loggerFromContextValue(ctx); logger != nil {
		fn(logger, ctx)
		return
	}
	if globalLogger != nil {
		globalLogger.logMissingLoggerWarning()
		// Pass the global logger in the context to allow further context-based logging.
		// So this Warning is only logged once.
		ctx = context.WithValue(ctx, loggerKey, globalLogger)
		fn(globalLogger, ctx)
	}
}

func SetLogLevel(ctx context.Context, level string) {
	if logger := loggerFromContextValue(ctx); logger != nil {
		logger.SetLogLevel(level)
		return
	}
	if globalLogger != nil {
		globalLogger.SetLogLevel(level)
	}
}

func GetLogLevel(ctx context.Context) string {
	if logger := loggerFromContextValue(ctx); logger != nil {
		return logger.GetLogLevel()
	}
	if globalLogger != nil {
		return globalLogger.GetLogLevel()
	}
	return ""
}
