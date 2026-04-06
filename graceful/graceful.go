package graceful

import (
	"context"
	"errors"
	"os/signal"
	"sync"
	"syscall"
)

var ErrContextClosed = errors.New("context closed")

type AbortableFunc func(context.Context)

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

func New() Graceful {
	return NewWithContext(context.Background())
}

func NewWithContext(ctx context.Context) Graceful {
	notifyCtx, stop := signal.NotifyContext(ctx, syscall.SIGINT, syscall.SIGTERM)
	return &internalGraceful{
		ctx:  notifyCtx,
		stop: stop,
		wg:   new(sync.WaitGroup),
	}
}

func (g *internalGraceful) Wait() {
	g.wg.Wait()
}

func (g *internalGraceful) Add(afs ...AbortableFunc) error {
	if g.ctx.Err() != nil {
		return ErrContextClosed
	}

	for _, af := range afs {
		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			af(g.ctx)
		}()
	}
	return nil
}

func (g *internalGraceful) AddWithCancel(afs ...AbortableFunc) (context.CancelFunc, error) {
	if g.ctx.Err() != nil {
		return nil, ErrContextClosed
	}

	ctx, cancel := context.WithCancel(g.ctx)
	for _, af := range afs {
		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			af(ctx)
		}()
	}
	return cancel, nil
}
