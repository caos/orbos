package immediate

import (
	"github.com/caos/orbos/internal/operator"
	"github.com/caos/orbos/internal/watcher"
	"github.com/caos/orbos/mntr"
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
