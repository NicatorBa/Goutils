package graceful

import (
	"context"
	"errors"
	"net/http"
	"time"
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
		srv := &http.Server{
			Addr:    options.Addr,
			Handler: handler,
		}

		go func() {
			srv.ListenAndServe()
		}()

		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), options.ShutdownTimeout)
		defer cancel()

		srv.Shutdown(shutdownCtx)
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
