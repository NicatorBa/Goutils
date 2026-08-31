package graceful

import (
	"context"
	"os/signal"
	"runtime/debug"
	"sync"
	"syscall"

	"github.com/NicatorBa/Goutils/graceful/logging"
)

type Runner interface {
	Run(context.Context)
}

type RunnerFunc func(context.Context)

func (r RunnerFunc) Run(ctx context.Context) {
	r(ctx)
}

type Initializer interface {
	Initialize() error
}

type Options struct {
	Ctx    context.Context
	Logger logging.Logger
}

type OptionFunc func(*Options) error

func WithContext(ctx context.Context) OptionFunc {
	return func(opts *Options) error {
		opts.Ctx = ctx
		return nil
	}
}

func WithLogger(logger logging.Logger) OptionFunc {
	return func(opts *Options) error {
		opts.Logger = logger
		return nil
	}
}

type Graceful interface {
	Wait()
	Add(...Runner) error
	AddWithCancel(...Runner) (context.CancelFunc, error)
}

type graceful struct {
	ctx  context.Context
	stop context.CancelFunc
	wg   *sync.WaitGroup
}

func New(opts ...OptionFunc) (Graceful, error) {
	options := Options{
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
	return &graceful{
		ctx:  notifyCtx,
		stop: stop,
		wg:   new(sync.WaitGroup),
	}, nil
}

func (g *graceful) Wait() {
	defer g.stop()
	g.wg.Wait()
}

func (g *graceful) Add(rs ...Runner) error {
	if g.ctx.Err() != nil {
		return ErrContextClosed
	}

	err := g.initialize(rs)
	if err != nil {
		return err
	}

	g.add(g.ctx, rs)
	return nil
}

func (g *graceful) AddWithCancel(rs ...Runner) (context.CancelFunc, error) {
	if g.ctx.Err() != nil {
		return nil, ErrContextClosed
	}

	err := g.initialize(rs)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(g.ctx)
	g.add(ctx, rs)
	return cancel, nil
}

func (g *graceful) initialize(rs []Runner) (err error) {
	defer recoverPanicIntoError(&err)

	for _, r := range rs {
		if init, ok := r.(Initializer); ok {
			if err := init.Initialize(); err != nil {
				return err
			}
		}
	}
	return nil
}

func (g *graceful) add(ctx context.Context, rs []Runner) {
	for _, r := range rs {
		g.wg.Go(func() {
			defer g.recoverPanic()
			r.Run(ctx)
		})
	}
}

func (g *graceful) recoverPanic() {
	if r := recover(); r != nil {
		err := &PanicError{Value: r, Stack: debug.Stack()}
		logging.From(g.ctx).Error().Err(err).Msgf("runner panicked: %s", err.Stack)
	}
}

func recoverPanicIntoError(err *error) {
	if r := recover(); r != nil {
		*err = &PanicError{Value: r, Stack: debug.Stack()}
	}
}
