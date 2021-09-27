package helper

import (
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/internal/operator/boom/templator"
	"github.com/caos/orbos/v5/internal/operator/boom/templator/helm"
	"github.com/caos/orbos/v5/internal/operator/boom/templator/yaml"
	"github.com/caos/orbos/v5/mntr"
)

func NewTemplator(monitor mntr.Monitor, overlay string, baseDirectoryPath string, templatorName name.Templator) templator.Templator {
	switch templatorName {
	case helm.GetName():
		return helm.New(monitor, overlay, baseDirectoryPath)
	case yaml.GetName():
		return yaml.New(monitor, overlay, baseDirectoryPath)
	}

	return nil
}
