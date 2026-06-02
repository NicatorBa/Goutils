package graceful

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/NicatorBa/Goutils/graceful/logging"
)

type (
	ServerOptions struct {
		Addr            string
		ShutdownTimeout time.Duration
	}

	ServerOptionFunc func(*ServerOptions) error
)

type serverRunner struct {
	handler http.Handler
	opts    []ServerOptionFunc
	options ServerOptions
}

func (s *serverRunner) Initialize() error {
	s.options = ServerOptions{
		ShutdownTimeout: 5 * time.Second,
	}
	for _, opt := range s.opts {
		if err := opt(&s.options); err != nil {
			return err
		}
	}
	return nil
}

func (s *serverRunner) Run(ctx context.Context) {
	logger := logging.From(ctx)
	srv := &http.Server{
		Addr:    s.options.Addr,
		Handler: s.handler,
	}

	go func() {
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			logger.Error().Err(err).Msg("HTTP server error")
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.options.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error().Err(err).Msg("HTTP server shutdown failed")
	}
}

func ListenAndServe(handler http.Handler, opts ...ServerOptionFunc) Runner {
	return &serverRunner{handler: handler, opts: opts}
}

func WithAddr(addr string) ServerOptionFunc {
	return func(opts *ServerOptions) error {
		opts.Addr = addr
		return nil
	}
}

func WithShutdownTimeout(timeout time.Duration) ServerOptionFunc {
	return func(opts *ServerOptions) error {
		if timeout <= 0 {
			return errors.New("timeout value must be greater than 0")
		}

		opts.ShutdownTimeout = timeout
		return nil
	}
}
