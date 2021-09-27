package helm

import (
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/templator"
	"github.com/caos/orbos/v5/mntr"
)

type TemplatorMutateValues interface {
	templator.HelmApplication
	HelmMutateValues(mntr.Monitor, *latest.ToolsetSpec, string) error
}

func (h *Helm) mutateValue(app interface{}, spec *latest.ToolsetSpec, valuesAbsFilePath string) error {
	mutate, ok := app.(TemplatorMutateValues)
	if ok {

		logFields := map[string]interface{}{
			"application": mutate.GetName().String(),
			"overlay":     h.overlay,
		}
		mutateMonitor := h.monitor.WithFields(logFields)

		mutateMonitor.Debug("Mutate values")

		if err := mutate.HelmMutateValues(mutateMonitor, spec, valuesAbsFilePath); err != nil {
			return err
		}
	}

	return nil
}
