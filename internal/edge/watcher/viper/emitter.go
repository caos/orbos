package viper

/*
import (
	"github.com/caos/infrop/internal/core/logging"
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/edge/watcher"
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
