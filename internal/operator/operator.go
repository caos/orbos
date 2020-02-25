package operator

import (
	"context"
	"fmt"
	"runtime/debug"
	"time"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/mntr"
)

type Watcher interface {
	Watch(fire chan<- struct{}) error
}

type Iterator struct {
	ctx      context.Context
	monitor  mntr.Monitor
	iterate  func()
	watchers []Watcher
	fired    chan struct{}
}

func New(ctx context.Context,
	monitor mntr.Monitor,
	iterate func(),
	watchers []Watcher) *Iterator {
	return &Iterator{
		ctx:      ctx,
		monitor:  monitor,
		iterate:  iterate,
		watchers: watchers,
		fired:    make(chan struct{}),
	}
}

func (i *Iterator) Initialize() error {

	for _, watcher := range i.watchers {
		if err := watcher.Watch(i.fired); err != nil {
			return errors.Wrap(err, "starting watcher failed")
		}
	}
	return nil
}

func (i *Iterator) Run() {

loop:
	for {
		select {
		case <-i.ctx.Done():
			return
		case <-i.fired:
			if len(i.fired) != 0 {
				i.monitor.Info("Skipping iteration")
				continue loop
			}
			i.iterateWrapped()
		}
	}
}

func (i *Iterator) iterateWrapped() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println(string(debug.Stack()))
			panic(r)
		}
	}()
	i.monitor.Debug("Starting iteration")
	started := time.Now()
	i.iterate()
	i.monitor.WithFields(map[string]interface{}{
		"took": time.Now().Sub(started),
	}).Info("Iteration done")
}
