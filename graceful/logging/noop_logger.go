package logging

type NoopLogger struct{}

func (n *NoopLogger) Info() Event { return noopEvent{} }

func (n *NoopLogger) Warn() Event { return noopEvent{} }

func (n *NoopLogger) Error() Event { return noopEvent{} }

type noopEvent struct{}

func (e noopEvent) Err(_ error) Event { return e }

func (e noopEvent) Msg(_ string) {}

func (e noopEvent) Msgf(_ string, _ ...any) {}
