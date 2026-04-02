package graceful

import (
	"context"
	"os/signal"
	"sync"
	"syscall"
)

type AbortableFunc func(context.Context)

type Graceful interface {
	Wait()
	Add(...AbortableFunc)
	AddWithCancel(...AbortableFunc) context.CancelFunc
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

func (g *internalGraceful) Add(afs ...AbortableFunc) {
	for _, af := range afs {
		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			af(g.ctx)
		}()
	}
}

func (g *internalGraceful) AddWithCancel(afs ...AbortableFunc) context.CancelFunc {
	ctx, cancel := context.WithCancel(g.ctx)
	for _, af := range afs {
		g.wg.Add(1)
		go func() {
			defer g.wg.Done()
			af(ctx)
		}()
	}
	return cancel
}
