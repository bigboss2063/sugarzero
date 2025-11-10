package loggerv2

import "context"

type Logger interface {
	Debug(context.Context, ...any)
	Debugf(context.Context, string, ...any)
	Debugln(context.Context, ...any)
	Info(context.Context, ...any)
	Infof(context.Context, string, ...any)
	Infoln(context.Context, ...any)
	Warn(context.Context, ...any)
	Warnf(context.Context, string, ...any)
	Warnln(context.Context, ...any)
	Error(context.Context, ...any)
	Errorf(context.Context, string, ...any)
	Errorln(context.Context, ...any)
	Fatal(context.Context, ...any)
	Fatalf(context.Context, string, ...any)
	Fatalln(context.Context, ...any)
	SetLogLevel(string)
	GetLogLevel() string
}
