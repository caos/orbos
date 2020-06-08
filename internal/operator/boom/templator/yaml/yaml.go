package yaml

import (
	"path/filepath"

	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/boom/templator"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

const (
	templatorName name.Templator = "yaml"
)

func GetName() name.Templator {
	return templatorName
}

type YAML struct {
	monitor                mntr.Monitor
	overlay                string
	templatorDirectoryPath string
}

func New(monitor mntr.Monitor, overlay, templatorDirectoryPath string) *YAML {
	return &YAML{
		monitor:                monitor,
		overlay:                overlay,
		templatorDirectoryPath: templatorDirectoryPath,
	}
}

func (y *YAML) getResultsFileDirectory(appName name.Application, overlay, basePath string) string {
	return filepath.Join(basePath, appName.String(), overlay, "results")
}

func (y *YAML) GetResultsFilePath(appName name.Application, overlay, basePath string) string {
	return filepath.Join(y.getResultsFileDirectory(appName, overlay, basePath), "results.yaml")
}

func (y *YAML) CleanUp() error {
	return nil
}

func checkTemplatorInterface(templatorInterface interface{}) (templator.YamlApplication, error) {
	app, isTemplator := templatorInterface.(templator.YamlApplication)
	if !isTemplator {
		err := errors.Errorf("YAML templating interface not implemented")
		return nil, err
	}

	return app, nil
}
