package boom

import (
	"github.com/caos/orbos/internal/operator/boom/app"
	gconfig "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/config"
	gitcrdconfig "github.com/caos/orbos/internal/operator/boom/gitcrd/config"
	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"path/filepath"
	"strings"
	"time"
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

func Takeoff(monitor mntr.Monitor, orb *orb.Orb, basePath string, localMode bool) func() {
	//toolsDirectoryPath := filepath.Join(basePath, "tools")
	crdDirectoryPath := filepath.Join(basePath, "crd")
	gconfig.DashboardsDirectoryPath = filepath.Join(basePath, "dashboards")

	appStruct := app.New(monitor, basePath)
	gitcrdMonitor := monitor.WithFields(map[string]interface{}{"type": "gitcrd"})

	gitcrdConf := &gitcrdconfig.Config{
		Monitor:          gitcrdMonitor,
		CrdDirectoryPath: crdDirectoryPath,
		CrdUrl:           orb.URL,
		PrivateKey:       []byte(orb.Repokey),
		CrdPath:          "boom.yml",
		User:             "Boom",
		Email:            "boom@caos.ch",
	}

	if localMode {
		clientgo.InConfig = false
	}

	if err := appStruct.AddGitCrd(gitcrdConf); err != nil {
		monitor.Error(errors.Wrap(err, "unable to start supervised crd"))
	}

	return func() {
		// TODO: use a function scoped error variable
		started := time.Now()
		goErr := appStruct.ReconcileGitCrds(orb.Masterkey)
		recMonitor := monitor.WithFields(map[string]interface{}{
			"took": time.Since(started),
		})
		if goErr != nil {
			recMonitor.Error(goErr)
		}
		recMonitor.Info("Reconciling iteration done")
	}
}

func TakeOffCurrentState(monitor mntr.Monitor, orb *orb.Orb, toolsDirectoryPath string) func() {
	appStruct := app.New(monitor, toolsDirectoryPath)
	gitcrdMonitor := monitor.WithFields(map[string]interface{}{"type": "gitcrd"})

	gitcrdConf := &gitcrdconfig.Config{
		Monitor:          gitcrdMonitor,
		CrdDirectoryPath: "/tmp/crd",
		CrdUrl:           orb.URL,
		PrivateKey:       []byte(orb.Repokey),
		CrdPath:          "boom.yml",
		User:             "Boom",
		Email:            "boom@caos.ch",
	}

	if err := appStruct.AddGitCrd(gitcrdConf); err != nil {
		monitor.Error(errors.Wrap(err, "unable to start supervised crd"))
	}

	return func() {
		started := time.Now()
		goErr := appStruct.WriteBackCurrentState(orb.Masterkey)
		recMonitor := monitor.WithFields(map[string]interface{}{
			"took": time.Since(started),
		})
		if goErr != nil {
			recMonitor.Error(goErr)
		}
		recMonitor.Info("Current state iteration done")
	}
}
