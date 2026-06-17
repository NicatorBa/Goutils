package logging

import "github.com/rs/zerolog"

type ZerologLogger struct {
	l zerolog.Logger
}

func NewZerologLoggerFrom(l zerolog.Logger) *ZerologLogger {
	return &ZerologLogger{l: l}
}

func (z *ZerologLogger) Info() Event { return &zerologEvent{e: z.l.Info()} }

func (z *ZerologLogger) Warn() Event { return &zerologEvent{e: z.l.Warn()} }

func (z *ZerologLogger) Error() Event { return &zerologEvent{e: z.l.Error()} }

type zerologEvent struct {
	e *zerolog.Event
}

func (z *zerologEvent) Err(err error) Event {
	z.e = z.e.Err(err)
	return z
}

func (z *zerologEvent) Msg(msg string) {
	z.e.Msg(msg)
}

func (z *zerologEvent) Msgf(format string, args ...any) {
	z.e.Msgf(format, args...)
}
