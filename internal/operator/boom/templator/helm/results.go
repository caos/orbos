package helm

import (
	"os"

	"github.com/caos/orbos/internal/operator/boom/templator"
	"github.com/pkg/errors"
)

func (h *Helm) deleteResults(app templator.HelmApplication) error {
	resultsFileDirectory := h.getResultsFileDirectory(app.GetName(), h.overlay, h.templatorDirectoryPath)
	if err := os.RemoveAll(resultsFileDirectory); err != nil {
		return errors.Wrapf(err, "Error while deleting result file directory in path %s", resultsFileDirectory)
	}

	if err := os.MkdirAll(resultsFileDirectory, os.ModePerm); err != nil {
		return errors.Wrapf(err, "Error while recreating result file directory in path %s", resultsFileDirectory)
	}

	return nil
}
