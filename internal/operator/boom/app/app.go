package app

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"strings"

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
	GitCrds            []gitcrd.GitCrd
	Crds               map[string]crd.Crd
	monitor            mntr.Monitor
}

func New(monitor mntr.Monitor, toolsDirectoryPath string) *App {

	app := &App{
		ToolsDirectoryPath: toolsDirectoryPath,
		monitor:            monitor,
	}

	app.Crds = make(map[string]crd.Crd, 0)
	app.GitCrds = make([]gitcrd.GitCrd, 0)

	return app
}

func (a *App) CleanUp() error {

	a.monitor.Info("Cleanup")

	for _, g := range a.GitCrds {
		g.CleanUp()

		if err := g.GetStatus(); err != nil {
			return err
		}
	}

	for _, c := range a.Crds {
		c.CleanUp()

		if err := c.GetStatus(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) AddGitCrd(gitCrdConf *gitcrdconfig.Config) error {
	c, err := gitcrd.New(gitCrdConf)
	if err != nil {
		return err
	}

	bundleConf := &bundleconfig.Config{
		Orb:               strings.TrimSuffix(strings.TrimPrefix(gitCrdConf.CrdUrl, "git@"), ".git"),
		BundleName:        bundles.Caos,
		BaseDirectoryPath: a.ToolsDirectoryPath,
		Templator:         helm.GetName(),
	}

	c.SetBundle(bundleConf)
	if err := c.GetStatus(); err != nil {
		return err
	}

	a.GitCrds = append(a.GitCrds, c)
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

func SelfReconcile(monitor mntr.Monitor, kubeconfig *string, version string) error {
	k8sClient := kubernetes.NewK8sClient(monitor, kubeconfig)
	if *kubeconfig == "" {
		err := k8sClient.RefreshLocal()
		if err != nil {
			return err
		}
	}

	if k8sClient.Available() {
		if err := kubernetes.EnsureBoomArtifacts(monitor, k8sClient, version); err != nil {
			monitor.Info("failed to deploy boom into k8s-cluster")
			return err
		}
		monitor.Info("Deployed boom")
	} else {
		monitor.Info("Failed to connect to k8s")
	}

	return nil
}

func (a *App) ReconcileGitCrds(masterkey string) error {
	monitor := a.monitor.WithFields(map[string]interface{}{
		"action": "reconciling",
	})
	monitor.Info("Started reconciling of GitCRDs")

	for _, crdGit := range a.GitCrds {
		crdGit.SetBackStatus()

		currentResourceList, err := a.getCurrent(monitor)
		if err != nil {
			return err
		}

		crdGit.Reconcile(currentResourceList, masterkey)
		if err := crdGit.GetStatus(); err != nil {
			return err
		}
	}
	return nil
}

func (a *App) WriteBackCurrentState(masterkey string) error {

	monitor := a.monitor.WithFields(map[string]interface{}{
		"action": "current",
	})
	monitor.Info("Started writeback of currentstate of GitCRDs")

	for _, crdGit := range a.GitCrds {
		crdGit.SetBackStatus()

		currentResourceList, err := a.getCurrent(monitor)
		if err != nil {
			return err
		}

		crdGit.WriteBackCurrentState(currentResourceList, masterkey)
		if err := crdGit.GetStatus(); err != nil {
			metrics.FailedWritingCurrentState(crdGit.GetRepoURL(), crdGit.GetRepoCRDPath())
			return err
		}
		metrics.SuccessfulWriteCurrentState(crdGit.GetRepoURL(), crdGit.GetRepoCRDPath())
	}
	return nil
}

// func (a *App) ReconcileCrd(version, namespacedName string, getToolsetCRD func(instance runtime.Object) error) error {
// 	a.monitor.WithFields(map[string]interface{}{
// 		"name": namespacedName,
// 	}).Info("Started reconciling of CRD")

// 	var err error
// 	managedcrd, ok := a.Crds[namespacedName]
// 	if !ok {
// 		crdConf := &crdconfig.Config{
// 			Monitor:  a.monitor,
// 			Version: v1beta1.GetVersion(),
// 		}

// 		managedcrd, err = crd.New(crdConf)
// 		if err != nil {
// 			return err
// 		}

// 		bundleConf := &bundleconfig.Config{
// 			Monitor:            a.monitor,
// 			CrdName:           namespacedName,
// 			BundleName:        bundles.Caos,
// 			BaseDirectoryPath: a.ToolsDirectoryPath,
// 			Templator:         helm.GetName(),
// 		}
// 		managedcrd.SetBundle(bundleConf)

// 		if err := managedcrd.GetStatus(); err != nil {
// 			return err
// 		}

// 		a.Crds[namespacedName] = managedcrd
// 	}

// 	managedcrd.ReconcileWithFunc(getToolsetCRD)
// 	return managedcrd.GetStatus()
// }
