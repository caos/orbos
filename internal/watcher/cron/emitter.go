package cron

import (
	"github.com/caos/orbiter/internal/operator"
	"github.com/caos/orbiter/internal/watcher"
	"github.com/caos/orbiter/mntr"
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
