package graceful

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/NicatorBa/Goutils/graceful/logging"
)

type (
	HttpListenAndServeOptions struct {
		Addr            string
		ShutdownTimeout time.Duration
	}

	HttpListenAndServeOpt func(*HttpListenAndServeOptions) error
)

func HttpListenAndServe(handler http.Handler, opts ...HttpListenAndServeOpt) AbortableFunc {
	options := HttpListenAndServeOptions{
		ShutdownTimeout: 5 * time.Second,
	}
	for _, opt := range opts {
		opt(&options)
	}

	return func(ctx context.Context) {
		logger := logging.From(ctx)
		srv := &http.Server{
			Addr:    options.Addr,
			Handler: handler,
		}

		go func() {
			err := srv.ListenAndServe()
			if err != nil {
				logger.Error().Err(err).Msg("HTTP server error")
			}
		}()

		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), options.ShutdownTimeout)
		defer cancel()

		err := srv.Shutdown(shutdownCtx)
		if err != nil {
			logger.Error().Err(err).Msg("HTTP server shutdown failed")
		}
	}
}

func WithAddr(addr string) HttpListenAndServeOpt {
	return func(opts *HttpListenAndServeOptions) error {
		opts.Addr = addr
		return nil
	}
}

func WithShutdownTimeout(timeout time.Duration) HttpListenAndServeOpt {
	return func(opts *HttpListenAndServeOptions) error {
		if timeout <= 0 {
			return errors.New("timeout value must be greater than 0")
		}

		opts.ShutdownTimeout = timeout
		return nil
	}
}
