package boom

import (
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/caos/orbos/internal/orb"

	"github.com/caos/orbos/internal/operator/boom/app"
	gconfig "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/config"
	gitcrdconfig "github.com/caos/orbos/internal/operator/boom/gitcrd/config"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func Metrics(monitor mntr.Monitor) {
	metricsport := "2112"

	http.Handle("/metrics", promhttp.Handler())
	address := strings.Join([]string{":", metricsport}, "")
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			monitor.Error(errors.Wrap(err, "error while serving metrics endpoint"))
		}

		monitor.WithFields(map[string]interface{}{
			"port":     metricsport,
			"endpoint": "/metrics",
		}).Info("Started metrics")
	}()
}

func Takeoff(monitor mntr.Monitor, toolsDirectoryPath string, localMode bool, orbpath string) (func(), func()) {
	gitcrdMonitor := monitor.WithFields(map[string]interface{}{"type": "gitcrd"})

	gitcrdConf := gitcrdconfig.Config{
		Monitor:          gitcrdMonitor,
		CrdDirectoryPath: filepath.Join(toolsDirectoryPath, "crd"),
		User:             "Boom",
		Email:            "boom@caos.ch",
	}

	if localMode {
		clientgo.InConfig = false
	}

	gconfig.DashboardsDirectoryPath = filepath.Join(toolsDirectoryPath, "dashboards")

	appStruct := app.New(monitor, toolsDirectoryPath)
	currentStruct := app.New(monitor, toolsDirectoryPath)

	return task(monitor, orbpath, gitcrdConf, appStruct.ReadSpecs, appStruct.Reconcile), task(monitor, orbpath, gitcrdConf, currentStruct.ReadSpecs, currentStruct.WriteBackCurrentState)
}

func task(monitor mntr.Monitor, orbpath string, gitcrdConf gitcrdconfig.Config, readSpecs func(gitCrdConf *gitcrdconfig.Config, repoURL string, repoKey []byte) error, do func() error) func() {
	return func() {
		// TODO: use a function scoped error variable
		started := time.Now()

		orbConfig, goErr := orb.ParseOrbConfig(orbpath)
		if goErr != nil {
			monitor.Error(goErr)
			return
		}

		if err := readSpecs(&gitcrdConf, orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			monitor.Error(errors.Wrap(err, "unable to start supervised crd"))
		}

		goErr = do()
		recMonitor := monitor.WithFields(map[string]interface{}{
			"took": time.Since(started),
		})
		recMonitor.Error(goErr)
		recMonitor.Info("Reconciling iteration done")
	}
}
