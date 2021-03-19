package app

import (
	"strings"

	"github.com/caos/orbos/internal/operator/boom/api"
	crdconfig "github.com/caos/orbos/internal/operator/boom/crd/config"
	"github.com/caos/orbos/pkg/tree"

	bundleconfig "github.com/caos/orbos/internal/operator/boom/bundle/config"
	"github.com/caos/orbos/internal/operator/boom/crd"
	"github.com/caos/orbos/internal/operator/boom/current"
	"github.com/caos/orbos/internal/operator/boom/gitcrd"
	gitcrdconfig "github.com/caos/orbos/internal/operator/boom/gitcrd/config"
	"github.com/caos/orbos/internal/operator/boom/metrics"
	"github.com/caos/orbos/internal/utils/clientgo"

	"github.com/caos/orbos/internal/operator/boom/bundle/bundles"
	"github.com/caos/orbos/internal/operator/boom/templator/helm"
	"github.com/caos/orbos/mntr"
)

type App struct {
	ToolsDirectoryPath string
	GitCrd             *gitcrd.GitCrd
	Crds               map[string]*crd.Crd
	monitor            mntr.Monitor
}

func New(monitor mntr.Monitor, toolsDirectoryPath string) *App {

	app := &App{
		ToolsDirectoryPath: toolsDirectoryPath,
		monitor:            monitor,
	}

	app.Crds = make(map[string]*crd.Crd, 0)

	return app
}

func (a *App) CleanUp() error {

	a.monitor.Info("Cleanup")

	a.GitCrd.CleanUp()

	if err := a.GitCrd.GetStatus(); err != nil {
		return err
	}

	for _, c := range a.Crds {
		c.CleanUp()

		if err := c.GetStatus(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) ReadSpecs(gitCrdConf *gitcrdconfig.Config, repoURL string, repoKey []byte) error {
	c := gitcrd.New(gitCrdConf)
	if err := c.Clone(repoURL, repoKey); err != nil {
		return err
	}

	bundleConf := &bundleconfig.Config{
		Orb:               strings.TrimSuffix(strings.TrimPrefix(repoURL, "git@"), ".git"),
		BundleName:        bundles.Caos,
		BaseDirectoryPath: a.ToolsDirectoryPath,
		Templator:         helm.GetName(),
	}

	c.SetBundle(bundleConf)
	if err := c.GetStatus(); err != nil {
		return err
	}

	a.GitCrd = c
	return nil
}

func (a *App) getCurrent(monitor mntr.Monitor) ([]*clientgo.Resource, error) {

	monitor.Info("Started determining current state")

	resourceInfoList, err := clientgo.GetGroupVersionsResources(monitor, []string{})
	if err != nil {
		monitor.Error(err)
		metrics.FailedReadingCurrentState()
		return nil, err
	}
	metrics.SuccessfulReadingCurrentState()

	return current.Get(a.monitor, resourceInfoList), nil
}

func (a *App) Reconcile() error {
	monitor := a.monitor.WithFields(map[string]interface{}{
		"action": "reconciling",
	})
	monitor.Info("Started reconciling of GitCRDs")

	a.GitCrd.SetBackStatus()

	currentResourceList, err := a.getCurrent(monitor)
	if err != nil {
		return err
	}

	a.GitCrd.Reconcile(currentResourceList)
	if err := a.GitCrd.GetStatus(); err != nil {
		return err
	}
	return nil
}

func (a *App) WriteBackCurrentState() error {

	monitor := a.monitor.WithFields(map[string]interface{}{
		"action": "current",
	})
	monitor.Info("Started writeback of currentstate of GitCRDs")

	a.GitCrd.SetBackStatus()

	currentResourceList, err := a.getCurrent(monitor)
	if err != nil {
		return err
	}

	a.GitCrd.WriteBackCurrentState(currentResourceList)
	if err := a.GitCrd.GetStatus(); err != nil {
		metrics.FailedWritingCurrentState(a.GitCrd.GetRepoURL())
		return err
	}
	metrics.SuccessfulWriteCurrentState(a.GitCrd.GetRepoURL())
	return nil
}

func (a *App) ReconcileCrd(namespacedName string, toolsetCRD *tree.Tree) error {
	a.monitor.WithFields(map[string]interface{}{
		"name": namespacedName,
	}).Info("Started reconciling of CRD")

	var err error
	managedcrd, ok := a.Crds[namespacedName]
	if !ok {
		crdConf := &crdconfig.Config{
			Monitor: a.monitor,
		}

		managedcrd = crd.New(crdConf)

		bundleConf := &bundleconfig.Config{
			Monitor:           a.monitor,
			CrdName:           namespacedName,
			BundleName:        bundles.Caos,
			BaseDirectoryPath: a.ToolsDirectoryPath,
			Templator:         helm.GetName(),
		}
		managedcrd.SetBundle(bundleConf)

		if err := managedcrd.GetStatus(); err != nil {
			return err
		}

		a.Crds[namespacedName] = managedcrd
	}

	currentResourceList, err := a.getCurrent(a.monitor)
	if err != nil {
		return err
	}

	desired, _, _, _, _, _, err := api.ParseToolset(toolsetCRD)
	if err != nil {
		return err
	}

	managedcrd.Reconcile(currentResourceList, desired, false)
	return managedcrd.GetStatus()
}
