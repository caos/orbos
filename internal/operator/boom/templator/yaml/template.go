package yaml

import (
	"path/filepath"

	"github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	helper2 "github.com/caos/orbos/v5/internal/utils/helper"
	"github.com/caos/orbos/v5/internal/utils/yaml"
)

func (y *YAML) Template(appInterface interface{}, spec *latest.ToolsetSpec, resultFunc func(string, string) error) error {
	app, err := checkTemplatorInterface(appInterface)
	if err != nil {
		return err
	}

	yamlInterface := app.GetYaml(y.monitor, spec)
	resultfilepath := y.GetResultsFilePath(app.GetName(), y.overlay, y.templatorDirectoryPath)
	resultfiledirectory := y.getResultsFileDirectory(app.GetName(), y.overlay, y.templatorDirectoryPath)

	resultAbsFilePath, err := filepath.Abs(resultfilepath)
	if err != nil {
		return err
	}
	resultAbsFileDirectory, err := filepath.Abs(resultfiledirectory)
	if err != nil {
		return err
	}

	if err := helper2.RecreatePath(resultAbsFileDirectory); err != nil {
		return err
	}

	yamlStr := yamlInterface.(string)
	if err := yaml.New(resultAbsFilePath).AddStringObject(yamlStr); err != nil {
		return err
	}

	return resultFunc(resultAbsFilePath, "")
}
