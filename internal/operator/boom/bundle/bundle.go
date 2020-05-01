package bundle

import (
	"sync"

	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application"
	"github.com/caos/orbos/internal/operator/boom/bundle/bundles"
	"github.com/caos/orbos/internal/operator/boom/bundle/config"
	"github.com/caos/orbos/internal/operator/boom/current"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/boom/templator"
	"github.com/caos/orbos/internal/operator/boom/templator/helm"
	helperTemp "github.com/caos/orbos/internal/operator/boom/templator/helper"
	"github.com/caos/orbos/internal/operator/boom/templator/yaml"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

var (
	Testmode bool = false
)

type Bundle struct {
	baseDirectoryPath string
	crdName           string
	Applications      map[name.Application]application.Application
	HelmTemplator     templator.Templator
	YamlTemplator     templator.Templator
	monitor           mntr.Monitor
}

func New(conf *config.Config) *Bundle {
	apps := make(map[name.Application]application.Application, 0)
	helmTemplator := helperTemp.NewTemplator(conf.Monitor, conf.CrdName, conf.BaseDirectoryPath, helm.GetName())
	yamlTemplator := helperTemp.NewTemplator(conf.Monitor, conf.CrdName, conf.BaseDirectoryPath, yaml.GetName())

	b := &Bundle{
		crdName:           conf.CrdName,
		baseDirectoryPath: conf.BaseDirectoryPath,
		monitor:           conf.Monitor,
		HelmTemplator:     helmTemplator,
		YamlTemplator:     yamlTemplator,
		Applications:      apps,
	}
	return b
}

func (b *Bundle) CleanUp() error {

	err := b.HelmTemplator.CleanUp()
	if err != nil {
		return err
	}

	return b.YamlTemplator.CleanUp()
}

func (b *Bundle) GetApplications() map[name.Application]application.Application {
	return b.Applications
}

func (b *Bundle) AddApplicationsByBundleName(name name.Bundle) error {

	names := bundles.Get(name)
	if names == nil {
		return errors.Errorf("No bundle known with name %s", name)
	}

	bnew := b
	for _, name := range names {
		if err := bnew.AddApplicationByName(name); err != nil {
			return err
		}
	}
	return nil
}

func (b *Bundle) AddApplicationByName(appName name.Application) error {

	app := application.New(b.monitor, appName)
	return b.AddApplication(app)
}

func (b *Bundle) AddApplication(app application.Application) error {

	if _, found := b.Applications[app.GetName()]; found {
		return errors.New("Application already in bundle")
	}

	b.Applications[app.GetName()] = app
	return nil
}

func (b *Bundle) Reconcile(currentResourceList []*clientgo.Resource, spec *v1beta1.ToolsetSpec) error {

	applicationCount := 0
	// go through list of application until every application is reconciled
	// and this orderNumber by orderNumber (default is 1)
	for orderNumber := 0; applicationCount < len(b.Applications); orderNumber++ {
		var wg sync.WaitGroup
		errList := make(map[name.Application]chan error, len(b.Applications))
		for appName := range b.Applications {
			//if application has the same orderNumber as currently iterating the reconcile the application
			if application.GetOrderNumber(appName) == orderNumber {
				wg.Add(1)
				errChan := make(chan error)
				go b.ReconcileApplication(currentResourceList, appName, spec, &wg, errChan)
				applicationCount++
				errList[appName] = errChan
			}
		}
		for appName, errChan := range errList {
			if err := <-errChan; err != nil {
				return errors.Wrapf(err, "Error while reconciling application %s", appName.String())
			}
		}
		wg.Wait()
	}

	return nil
}

func (b *Bundle) ReconcileApplication(currentResourceList []*clientgo.Resource, appName name.Application, spec *v1beta1.ToolsetSpec, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	logFields := map[string]interface{}{
		"application": appName,
		"action":      "reconciling",
	}
	monitor := b.monitor.WithFields(logFields)

	app, found := b.Applications[appName]
	if !found {
		err := errors.New("Application not found")
		monitor.Error(err)
		errChan <- err
		return
	}
	monitor.Info("Start")

	deploy := app.Deploy(spec)
	currentApplicationResourceList := current.FilterForApplication(appName, currentResourceList)

	var resultFunc func(string, string) error
	if Testmode {
		resultFunc = func(resultFilePath, namespace string) error {
			return nil
		}
	} else {
		if deploy {
			resultFunc = applyWithCurrentState(monitor, currentApplicationResourceList, app, spec.ForceApply)
		} else {
			resultFunc = deleteWithCurrentState(monitor, currentApplicationResourceList, app)
		}
	}

	_, usedHelm := app.(application.HelmApplication)
	if usedHelm {
		err := b.HelmTemplator.Template(app, spec, resultFunc)
		if err != nil {
			errChan <- err
			return
		}
	}
	_, usedYaml := app.(application.YAMLApplication)
	if usedYaml {
		err := b.YamlTemplator.Template(app, spec, resultFunc)
		if err != nil {
			errChan <- err
			return
		}
	}

	monitor.Info("Done")
	errChan <- nil
}
