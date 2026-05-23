package logging

import (
	"fmt"
	"log"
	"os"
	"strings"
)

type StdLogger struct {
	info  *log.Logger
	warn  *log.Logger
	error *log.Logger
}

func NewStdLogger() *StdLogger {
	flags := log.LstdFlags | log.Lmsgprefix
	return &StdLogger{
		info:  log.New(os.Stderr, "INFO ", flags),
		warn:  log.New(os.Stderr, "WARN ", flags),
		error: log.New(os.Stderr, "ERROR ", flags),
	}
}

func (s *StdLogger) Info() Event { return newStdEvent(s.info) }

func (s *StdLogger) Warn() Event { return newStdEvent(s.warn) }

func (s *StdLogger) Error() Event { return newStdEvent(s.error) }

type stdEvent struct {
	l      *log.Logger
	fields []field
}

func newStdEvent(l *log.Logger) *stdEvent {
	return &stdEvent{l: l}
}

func (e *stdEvent) Err(err error) Event {
	if err != nil {
		e.fields = append(e.fields, field{"error", err.Error()})
	}
	return e
}

func (e *stdEvent) Msg(msg string) {
	e.write(msg)
}

func (e *stdEvent) Msgf(format string, args ...any) {
	e.write(fmt.Sprintf(format, args...))
}

func (e *stdEvent) write(msg string) {
	line := msg + formatFields(e.fields)
	e.l.Println(line)
}

type field struct {
	key string
	val any
}

func formatFields(fields []field) string {
	var out strings.Builder
	for _, f := range fields {
		fmt.Fprintf(&out, " %s=%v", f.key, f.val)
	}
	return out.String()
}
