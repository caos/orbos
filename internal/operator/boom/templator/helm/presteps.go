package helm

import (
	"fmt"

	"github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/templator"
	"github.com/caos/orbos/v5/internal/utils/yaml"
	"github.com/caos/orbos/v5/mntr"
)

type TemplatorPreSteps interface {
	templator.HelmApplication
	HelmPreApplySteps(mntr.Monitor, *latest.ToolsetSpec) ([]interface{}, error)
}

func (h *Helm) preApplySteps(app interface{}, spec *latest.ToolsetSpec) error {

	pre, ok := app.(TemplatorPreSteps)
	if ok {

		logFields := map[string]interface{}{
			"application": pre.GetName().String(),
			"overlay":     h.overlay,
		}

		monitor := h.monitor.WithFields(logFields)
		monitor.Debug("Pre-steps")
		resources, err := pre.HelmPreApplySteps(monitor, spec)
		if err != nil {
			return fmt.Errorf("error while processing pre-steps for application %s: %w", pre.GetName().String(), err)
		}

		resultfilepath := h.GetResultsFilePath(pre.GetName(), h.overlay, h.templatorDirectoryPath)
		y := yaml.New(resultfilepath)
		for i, resource := range resources {
			value, isString := resource.(string)

			if isString {
				err := y.AddStringObject(value)
				if err != nil {
					return fmt.Errorf("error while adding element %d to result-file: %w", i, err)
				}
			} else {
				err = y.AddStruct(resource)
				if err != nil {
					return fmt.Errorf("error while adding element %d to result-file: %w", i, err)
				}
			}
		}
	}
	return nil
}
