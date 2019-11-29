package immediate

import (
	"github.com/caos/infrop/internal/core/logging"
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/edge/watcher"
)

func New(logger logging.Logger) operator.Watcher {
	return watcher.Func(func(changes chan<- struct{}) error {
		logger.Debug("Immediate triggered")
		go func() {
			changes <- struct{}{}
		}()
		return nil
	})
}
