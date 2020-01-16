package cron

import (
	"github.com/caos/orbiter/internal/operator"
	"github.com/caos/orbiter/internal/watcher"
	"github.com/caos/orbiter/logging"
	"github.com/robfig/cron"
)

func New(logger logging.Logger, pattern string) operator.Watcher {
	return watcher.Func(func(changes chan<- struct{}) error {
		cr := cron.New()
		if err := cr.AddFunc(pattern, func() {
			logger.Debug("Cron triggered")
			changes <- struct{}{}
		}); err != nil {
			return err
		}
		go cr.Run()
		return nil
	})
}
