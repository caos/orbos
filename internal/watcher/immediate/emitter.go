package immediate

import (
	"github.com/caos/orbiter/internal/operator"
	"github.com/caos/orbiter/internal/watcher"
	"github.com/caos/orbiter/mntr"
)

func New(monitor mntr.Monitor) operator.Watcher {
	return watcher.Func(func(changes chan<- struct{}) error {
		monitor.Debug("Immediate triggered")
		go func() {
			changes <- struct{}{}
		}()
		return nil
	})
}
