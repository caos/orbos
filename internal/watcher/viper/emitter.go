package viper

/*
import (
	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/operator"
	"github.com/caos/orbiter/internal/watcher"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

func New(logger logging.Logger, viper *viper.Viper) operator.Watcher {
	fieldLogger := logger.WithFields(map[string]interface{}{
		"file": viper.ConfigFileUsed(),
	})
	return watcher.Func(func(events chan<- struct{}) error {
		viper.OnConfigChange(func(ev fsnotify.Event) {
			fieldLogger.Debug("Configuration changed")
			events <- struct{}{}
		})
		go viper.WatchConfig()
		return nil
	})
}
*/
