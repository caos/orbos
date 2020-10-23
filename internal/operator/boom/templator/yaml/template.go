package yaml

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	helper2 "github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/internal/utils/yaml"
	"path/filepath"
)

func (y *YAML) Template(appInterface interface{}, spec *v1beta2.ToolsetSpec, resultFunc func(string, string) error) error {
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

	return resultFunc(resultAbsFilePath, "caos-system")
}
