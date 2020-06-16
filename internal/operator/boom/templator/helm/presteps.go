package helm

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/templator"
	"github.com/caos/orbos/internal/utils/yaml"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

type TemplatorPreSteps interface {
	templator.HelmApplication
	HelmPreApplySteps(mntr.Monitor, *v1beta2.ToolsetSpec) ([]interface{}, error)
}

func (h *Helm) preApplySteps(app interface{}, spec *v1beta2.ToolsetSpec) error {

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
			return errors.Wrapf(err, "Error while processing pre-steps for application %s", pre.GetName().String())
		}

		resultfilepath := h.GetResultsFilePath(pre.GetName(), h.overlay, h.templatorDirectoryPath)
		y := yaml.New(resultfilepath)
		for i, resource := range resources {
			value, isString := resource.(string)

			if isString {
				err := y.AddStringObject(value)
				if err != nil {
					return errors.Wrapf(err, "Error while adding element %d to result-file", i)
				}
			} else {
				err = y.AddStruct(resource)
				if err != nil {
					return errors.Wrapf(err, "Error while adding element %d to result-file", i)
				}
			}
		}
	}
	return nil
}
