package operator

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/logging"
)

type Watcher interface {
	Watch(fire chan<- struct{}) error
}

type Iterator struct {
	ctx      context.Context
	logger   logging.Logger
	iterate  func()
	watchers []Watcher
	fired    chan struct{}
}

func New(ctx context.Context,
	logger logging.Logger,
	iterate func(),
	watchers []Watcher) *Iterator {
	return &Iterator{
		ctx:      ctx,
		logger:   logger,
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
	//	refspecs := []config.RefSpec{config.RefSpec("+refs/heads/master:refs/remotes/origin/master")}

loop:
	for {
		select {
		case <-i.ctx.Done():
			return
		case <-i.fired:
			if len(i.fired) != 0 {
				i.logger.Info("Skipping iteration")
				continue loop
			}
			started := time.Now()
			i.iterate()
			i.logger.WithFields(map[string]interface{}{
				"took": time.Now().Sub(started),
			}).Info("Iteration done")
		}
	}
}
