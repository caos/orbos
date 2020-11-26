package bundle

import (
	"github.com/caos/orbos/internal/operator/boom/application/types"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/internal/utils/kubectl"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func deleteWithCurrentState(monitor mntr.Monitor, currentResourceList []*clientgo.Resource, app types.Application) func(resultFilePath, namespace string) error {
	logFields := map[string]interface{}{
		"application": app.GetName().String,
		"command":     "delete",
	}
	delMonitor := monitor.WithFields(logFields)

	resultFunc := func(resultFilePath, namespace string) error {

		for _, resource := range currentResourceList {
			if err := clientgo.DeleteResource(resource); err != nil {
				err := errors.Wrap(err, "Failed to delete resource")
				delMonitor.Error(err)
				return err
			}
		}
		return nil
	}

	return resultFunc
}

func delete(monitor mntr.Monitor) func(resultFilePath, namespace string) error {

	logFields := map[string]interface{}{
		"command": "delete",
	}
	delMonitor := monitor.WithFields(logFields)

	resultFunc := func(resultFilePath, namespace string) error {
		kubectlCmd := kubectl.New("delete").AddParameter("-f", resultFilePath).AddFlag("--ignore-not-found")
		if namespace != "" {
			kubectlCmd.AddParameter("-n", namespace)
		}
		err := helper.Run(delMonitor, kubectlCmd.Build())
		return errors.Wrapf(err, "Failed to delete with file %s", resultFilePath)

	}
	return resultFunc
}
