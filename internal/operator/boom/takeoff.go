package boom

import (
	"fmt"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/caos/orbos/v5/internal/operator/boom/app"
	gconfig "github.com/caos/orbos/v5/internal/operator/boom/application/applications/monitoring/config"
	gitcrdconfig "github.com/caos/orbos/v5/internal/operator/boom/gitcrd/config"
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/git"
	"github.com/caos/orbos/v5/pkg/orb"
)

func Metrics(monitor mntr.Monitor) {
	metricsport := "2112"

	http.Handle("/metrics", promhttp.Handler())
	address := strings.Join([]string{":", metricsport}, "")
	go func() {
		if err := http.ListenAndServe(address, nil); err != nil {
			panic(fmt.Errorf("error while serving metrics endpoint: %w", err))
		}

		monitor.WithFields(map[string]interface{}{
			"port":     metricsport,
			"endpoint": "/metrics",
		}).Info("Started metrics")
	}()
}

func Takeoff(monitor mntr.Monitor, toolsDirectoryPath string, orbpath string, ensureClient, queryClient *git.Client) (func(), func()) {
	gitcrdMonitor := monitor.WithField("type", "gitcrd")

	gconfig.DashboardsDirectoryPath = filepath.Join(toolsDirectoryPath, "dashboards")

	appStruct := app.New(monitor, toolsDirectoryPath)
	currentStruct := app.New(monitor, toolsDirectoryPath)

	return task(
			monitor,
			orbpath,
			gitConf(gitcrdMonitor.WithField("task", "ensure"), ensureClient, toolsDirectoryPath),
			appStruct.ReadSpecs,
			appStruct.Reconcile),
		task(
			monitor,
			orbpath,
			gitConf(gitcrdMonitor.WithField("task", "query"), queryClient, toolsDirectoryPath),
			currentStruct.ReadSpecs,
			currentStruct.WriteBackCurrentState)
}

func gitConf(monitor mntr.Monitor, client *git.Client, toolsDirectoryPath string) gitcrdconfig.Config {
	return gitcrdconfig.Config{
		Monitor:          monitor,
		CrdDirectoryPath: filepath.Join(toolsDirectoryPath, "crd"),
		Git:              client,
	}
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
			monitor.Error(fmt.Errorf("unable to start supervised crd: %w", err))
		}

		goErr = do()
		recMonitor := monitor.WithFields(map[string]interface{}{
			"took": time.Since(started),
		})
		recMonitor.Error(goErr)
		recMonitor.Info("Reconciling iteration done")
	}
}
