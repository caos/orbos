package cron

import (
	"github.com/caos/orbos/internal/operator"
	"github.com/caos/orbos/internal/watcher"
	"github.com/caos/orbos/mntr"
	"github.com/robfig/cron"
)

func New(monitor mntr.Monitor, pattern string) operator.Watcher {
	return watcher.Func(func(changes chan<- struct{}) error {
		cr := cron.New()
		if err := cr.AddFunc(pattern, func() {
			monitor.Debug("Cron triggered")
			changes <- struct{}{}
		}); err != nil {
			return err
		}
		go cr.Run()
		return nil
	})
}
