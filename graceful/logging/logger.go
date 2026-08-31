package logging

import "context"

type Event interface {
	Err(err error) Event
	Msg(msg string)
	Msgf(format string, args ...any)
}

type Logger interface {
	Info() Event
	Warn() Event
	Error() Event
}

type contextKey struct{}

func With(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

func From(ctx context.Context) Logger {
	if l, ok := ctx.Value(contextKey{}).(Logger); ok {
		return l
	}
	return &NoopLogger{}
}
