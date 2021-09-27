package helm

import (
	"fmt"
	"os"

	"github.com/caos/orbos/v5/internal/operator/boom/templator"
)

func (h *Helm) deleteResults(app templator.HelmApplication) error {
	resultsFileDirectory := h.getResultsFileDirectory(app.GetName(), h.overlay, h.templatorDirectoryPath)
	if err := os.RemoveAll(resultsFileDirectory); err != nil {
		return fmt.Errorf("error while deleting result file directory in path %s: %w", resultsFileDirectory, err)
	}

	if err := os.MkdirAll(resultsFileDirectory, os.ModePerm); err != nil {
		return fmt.Errorf("error while recreating result file directory in path %s: %w", resultsFileDirectory, err)
	}

	return nil
}
