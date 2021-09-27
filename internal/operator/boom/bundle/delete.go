package bundle

import (
	"fmt"

	"github.com/caos/orbos/v5/internal/operator/boom/application"
	"github.com/caos/orbos/v5/internal/utils/clientgo"
	"github.com/caos/orbos/v5/mntr"
)

func deleteWithCurrentState(monitor mntr.Monitor, currentResourceList []*clientgo.Resource, app application.Application) func(resultFilePath, namespace string) error {

	resultFunc := func(resultFilePath, namespace string) error {

		for _, resource := range currentResourceList {
			if err := clientgo.DeleteResource(monitor, resource); err != nil {
				return fmt.Errorf("failed to delete resource for application %s: %w", app.GetName().String(), err)
			}
		}
		return nil
	}

	return resultFunc
}
