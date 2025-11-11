package sugarzero

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/rs/zerolog"
)

type ctxKey struct{ name string }

const (
	// callerSkipFramePublic is the skip frame count for public log methods (Debug, Info, etc.)
	callerSkipFramePublic = 5
	// callerSkipFrameInternal is the skip frame count when logging internal warnings
	callerSkipFrameInternal = 3
)

var (
	loggerKey        = ctxKey{name: "logger"}
	fieldsKey        = ctxKey{name: "fields"}
	configureZerolog sync.Once
	globalLogger     *ZeroLogger
)

// ZeroLogger wraps zerolog and satisfies the Logger interface.
type ZeroLogger struct {
	mu     sync.RWMutex
	logger zerolog.Logger
	level  zerolog.Level
}

// Reset resets the global logger state. This is intended for testing purposes only.
// ! Notice: DO NOT use this in production code.
func Reset() {
	globalLogger = nil
	configureZerolog = sync.Once{}
}

// New creates a zerolog-backed Logger, stores it as the global default, and
// injects it into the returned context. When writers is empty, os.Stdout is used.
// ! Notice: This function should be called only once during application initialization.
func New(ctx context.Context, level string, writers ...io.Writer) (context.Context, error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if globalLogger != nil {
		return context.WithValue(ctx, loggerKey, globalLogger), nil
	}

	lvl, err := parseLevel(level)
	if err != nil {
		return ctx, err
	}

	writer := selectWriter(writers...)

	configureZerolog.Do(func() {
		// Configure zerolog to use "position" as caller field name and uppercase level
		zerolog.CallerFieldName = "position"
		zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
			return fmt.Sprintf("%s:%d", file, line)
		}
		zerolog.LevelFieldMarshalFunc = func(l zerolog.Level) string {
			return strings.ToUpper(l.String())
		}

		// Create logger with native Caller() for position
		base := zerolog.New(writer).
			Level(lvl).
			With().
			Timestamp().
			Caller().
			Logger()

		globalLogger = &ZeroLogger{
			logger: base,
			level:  lvl,
		}
	})

	if globalLogger == nil {
		return ctx, fmt.Errorf("sugarzero: logger not initialized")
	}

	return context.WithValue(ctx, loggerKey, globalLogger), nil
}

// WithFields merges the provided fields into the context so they are emitted
// on the next log call. Fields should be provided as alternating key-value pairs.
// Example: WithFields(ctx, "user_id", 123, "action", "login")
func WithFields(ctx context.Context, keyvals ...any) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(keyvals) == 0 {
		return ctx
	}

	// keyvals must be in pairs
	if len(keyvals)%2 != 0 {
		// Ignore the last odd value
		keyvals = keyvals[:len(keyvals)-1]
	}

	flat := make([]any, 0, len(keyvals))
	for i := 0; i < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok || key == "" {
			// Skip non-string keys
			continue
		}
		flat = append(flat, key, keyvals[i+1])
	}

	if len(flat) == 0 {
		return ctx
	}

	if existing, ok := ctx.Value(fieldsKey).([]any); ok && len(existing) > 0 {
		merged := make([]any, 0, len(existing)+len(flat))
		merged = append(merged, existing...)
		flat = append(merged, flat...)
	}

	return context.WithValue(ctx, fieldsKey, flat)
}

// WithField is a convenience wrapper to add a single field to the context.
func WithField(ctx context.Context, key string, value any) context.Context {
	if key == "" {
		return ctx
	}
	return WithFields(ctx, key, value)
}

// FieldsFromContext exposes the currently attached fields.
func FieldsFromContext(ctx context.Context) map[string]any {
	flat := flattenedFieldsFromContext(ctx)
	if len(flat) == 0 {
		return nil
	}

	fields := make(map[string]any, len(flat)/2)
	for i := 0; i+1 < len(flat); i += 2 {
		key, _ := flat[i].(string)
		fields[key] = flat[i+1]
	}
	return fields
}

func (l *ZeroLogger) Debug(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.DebugLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Debugf(ctx context.Context, format string, args ...any) {
	l.writef(ctx, zerolog.DebugLevel, callerSkipFramePublic, format, args...)
}

func (l *ZeroLogger) Debugln(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.DebugLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Info(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.InfoLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Infof(ctx context.Context, format string, args ...any) {
	l.writef(ctx, zerolog.InfoLevel, callerSkipFramePublic, format, args...)
}

func (l *ZeroLogger) Infoln(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.InfoLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Warn(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.WarnLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Warnf(ctx context.Context, format string, args ...any) {
	l.writef(ctx, zerolog.WarnLevel, callerSkipFramePublic, format, args...)
}

func (l *ZeroLogger) Warnln(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.WarnLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Error(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.ErrorLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Errorf(ctx context.Context, format string, args ...any) {
	l.writef(ctx, zerolog.ErrorLevel, callerSkipFramePublic, format, args...)
}

func (l *ZeroLogger) Errorln(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.ErrorLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Fatal(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.FatalLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) Fatalf(ctx context.Context, format string, args ...any) {
	l.writef(ctx, zerolog.FatalLevel, callerSkipFramePublic, format, args...)
}

func (l *ZeroLogger) Fatalln(ctx context.Context, args ...any) {
	l.writeArgs(ctx, zerolog.FatalLevel, callerSkipFramePublic, args...)
}

func (l *ZeroLogger) SetLogLevel(level string) error {
	lvl, err := parseLevel(level)
	if err != nil {
		return err
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	l.level = lvl
	l.logger = l.logger.Level(lvl)

	return nil
}

func (l *ZeroLogger) GetLogLevel() string {
	l.mu.RLock()
	lvl := l.level
	l.mu.RUnlock()
	return lvl.String()
}

func (l *ZeroLogger) writeArgs(ctx context.Context, level zerolog.Level, skipFrame int, args ...any) {
	l.mu.RLock()
	logger := l.logger
	l.mu.RUnlock()

	event := logger.WithLevel(level).CallerSkipFrame(skipFrame)
	if event == nil {
		return
	}

	if fields := flattenedFieldsFromContext(ctx); len(fields) > 0 {
		event.Fields(fields)
	}

	if len(args) == 0 {
		event.Msg("")
		return
	}

	if len(args) == 1 {
		event.Msgf("%v", args[0])
	} else {
		event.Msg(fmt.Sprint(args...))
	}
}

func (l *ZeroLogger) writef(ctx context.Context, level zerolog.Level, skipFrame int, format string, args ...any) {
	l.mu.RLock()
	logger := l.logger
	l.mu.RUnlock()

	event := logger.WithLevel(level).CallerSkipFrame(skipFrame)
	if event == nil {
		return
	}

	if fields := flattenedFieldsFromContext(ctx); len(fields) > 0 {
		event.Fields(fields)
	}

	event.Msgf(format, args...)
}

func (l *ZeroLogger) logMissingLoggerWarning() {
	l.mu.RLock()
	logger := l.logger
	l.mu.RUnlock()

	event := logger.WithLevel(zerolog.WarnLevel).CallerSkipFrame(callerSkipFrameInternal)
	if event == nil {
		return
	}
	event.Msg("context does not contain a logger, using fallback logger")
}

func parseLevel(level string) (zerolog.Level, error) {
	if level == "" {
		return zerolog.InfoLevel, nil
	}
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		return zerolog.InfoLevel, fmt.Errorf("invalid log level %q: %w", level, err)
	}
	return lvl, nil
}

func selectWriter(writers ...io.Writer) io.Writer {
	if len(writers) == 0 {
		return os.Stdout
	}
	if len(writers) == 1 {
		return writers[0]
	}
	return io.MultiWriter(writers...)
}

func flattenedFieldsFromContext(ctx context.Context) []any {
	if ctx == nil {
		return nil
	}
	if fields, ok := ctx.Value(fieldsKey).([]any); ok && len(fields) > 0 {
		return fields
	}
	return nil
}

func loggerFromContextValue(ctx context.Context) *ZeroLogger {
	if ctx == nil {
		return nil
	}
	if ctxLogger, ok := ctx.Value(loggerKey).(*ZeroLogger); ok && ctxLogger != nil {
		return ctxLogger
	}
	return nil
}
