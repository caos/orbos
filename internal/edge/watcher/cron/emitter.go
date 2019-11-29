package cron

import (
	"github.com/caos/infrop/internal/core/logging"
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/edge/watcher"
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
