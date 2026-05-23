package graceful

import (
	"context"
	"errors"
	"os/signal"
	"sync"
	"syscall"

	"github.com/NicatorBa/Goutils/graceful/logging"
)

var ErrContextClosed = errors.New("context closed")

type AbortableFunc func(context.Context)

type GracefulOptions struct {
	Ctx    context.Context
	Logger logging.Logger
}

type GracefulOpt func(*GracefulOptions) error

func WithContext(ctx context.Context) GracefulOpt {
	return func(opts *GracefulOptions) error {
		opts.Ctx = ctx
		return nil
	}
}

func WithLogger(logger logging.Logger) GracefulOpt {
	return func(opts *GracefulOptions) error {
		opts.Logger = logger
		return nil
	}
}

type Graceful interface {
	Wait()
	Add(...AbortableFunc) error
	AddWithCancel(...AbortableFunc) (context.CancelFunc, error)
}

type internalGraceful struct {
	ctx  context.Context
	stop context.CancelFunc
	wg   *sync.WaitGroup
}

func New(opts ...GracefulOpt) (Graceful, error) {
	options := GracefulOptions{
		Ctx:    context.Background(),
		Logger: &logging.NoopLogger{},
	}
	for _, opt := range opts {
		err := opt(&options)
		if err != nil {
			return nil, err
		}
	}

	logCtx := logging.With(options.Ctx, options.Logger)
	notifyCtx, stop := signal.NotifyContext(logCtx, syscall.SIGINT, syscall.SIGTERM)
	return &internalGraceful{
		ctx:  notifyCtx,
		stop: stop,
		wg:   new(sync.WaitGroup),
	}, nil
}

func (g *internalGraceful) Wait() {
	g.wg.Wait()
}

func (g *internalGraceful) Add(afs ...AbortableFunc) error {
	if g.ctx.Err() != nil {
		return ErrContextClosed
	}

	g.add(g.ctx, afs)
	return nil
}

func (g *internalGraceful) AddWithCancel(afs ...AbortableFunc) (context.CancelFunc, error) {
	if g.ctx.Err() != nil {
		return nil, ErrContextClosed
	}

	ctx, cancel := context.WithCancel(g.ctx)
	g.add(ctx, afs)
	return cancel, nil
}

func (g *internalGraceful) add(ctx context.Context, afs []AbortableFunc) {
	for _, af := range afs {
		g.wg.Go(func() {
			af(ctx)
		})
	}
}
