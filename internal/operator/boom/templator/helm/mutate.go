package helm

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/templator"
	"github.com/caos/orbos/mntr"
)

type TemplatorMutate interface {
	templator.HelmApplication
	HelmMutate(mntr.Monitor, *v1beta2.ToolsetSpec, string) error
}

func (h *Helm) mutate(app interface{}, spec *v1beta2.ToolsetSpec) error {

	mutate, ok := app.(TemplatorMutate)
	if ok {

		logFields := map[string]interface{}{
			"application": mutate.GetName().String(),
			"overlay":     h.overlay,
		}
		mutateMonitor := h.monitor.WithFields(logFields)

		mutateMonitor.WithFields(logFields).Debug("Mutate before apply")

		resultfilepath := h.GetResultsFilePath(mutate.GetName(), h.overlay, h.templatorDirectoryPath)

		if err := mutate.HelmMutate(mutateMonitor, spec, resultfilepath); err != nil {
			return err
		}
	}

	return nil
}
