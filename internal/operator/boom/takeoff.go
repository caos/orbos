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
	"strings"
	"time"
)

func Takeoff(monitor mntr.Monitor, orb *orb.Orb, toolsDirectoryPath string, localMode bool, finishedChan chan bool) func() {
	appStruct := app.New(monitor, toolsDirectoryPath)
	gitcrdMonitor := monitor.WithFields(map[string]interface{}{"type": "gitcrd"})

	gitcrdConf := &gitcrdconfig.Config{
		Monitor:          gitcrdMonitor,
		CrdDirectoryPath: "/boom/crd",
		CrdUrl:           orb.URL,
		PrivateKey:       []byte(orb.Repokey),
		CrdPath:          "boom.yml",
		User:             "Boom",
		Email:            "boom@caos.ch",
	}

	if localMode {
		clientgo.InConfig = false
	}

	gconfig.DashboardsDirectoryPath = "/boom/dashboards"

	if err := appStruct.AddGitCrd(gitcrdConf); err != nil {
		monitor.Error(errors.Wrap(err, "unable to start supervised crd"))
	}

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

func TakeOffCurrentState(monitor mntr.Monitor, orb *orb.Orb, toolsDirectoryPath string, finishedChan chan bool) func() {
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
