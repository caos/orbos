package bundle

import (
	"path/filepath"

	"github.com/caos/orbos/internal/operator/boom/application/types"
	"github.com/caos/orbos/internal/operator/boom/desired"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/mntr"
)

func applyWithCurrentState(monitor mntr.Monitor, currentResourceList []*clientgo.Resource, app types.Application, force bool) func(resultFilePath, namespace string) error {

	logFields := map[string]interface{}{
		"command": "apply",
	}
	applyMonitor := monitor.WithFields(logFields)

	resultFunc := func(resultFilePath, namespace string) error {
		applyFunc := apply(monitor, app, force)

		desiredResources, err := desired.Get(monitor, resultFilePath, namespace, app.GetName())
		if err != nil {
			return err
		}

		if err := applyFunc(resultFilePath, namespace); err != nil {
			return err
		}
		deleteResources := make([]*clientgo.Resource, 0)
		for _, currentResource := range currentResourceList {
			found := false
			for _, desiredResource := range desiredResources {
				apiVersion := filepath.Join(currentResource.Group, currentResource.Version)
				if desiredResource.ApiVersion == apiVersion &&
					desiredResource.Kind == currentResource.Kind &&
					desiredResource.Metadata.Name == currentResource.Name &&
					(currentResource.Namespace == "" || desiredResource.Metadata.Namespace == currentResource.Namespace) {
					found = true
					break
				}
			}
			if found == false {
				otherAPI := false
				for _, desiredResource := range desiredResources {
					apiVersion := filepath.Join(currentResource.Group, currentResource.Version)
					if apiVersion != desiredResource.ApiVersion &&
						currentResource.Kind == desiredResource.Kind &&
						currentResource.Name == desiredResource.Metadata.Name &&
						(currentResource.Namespace == "" || desiredResource.Metadata.Namespace == currentResource.Namespace) {
						otherAPI = true
						break
					}
				}

				if !otherAPI {
					deleteResources = append(deleteResources, currentResource)
				}
			}
		}
		applyMonitor.Debug("Resources to delete calculated")

		if deleteResources != nil && len(deleteResources) > 0 {
			for _, deleteResource := range deleteResources {
				if err := clientgo.DeleteResource(deleteResource); err != nil {
					return err
				}
			}
		}

		return nil
	}

	return resultFunc
}

func apply(monitor mntr.Monitor, app types.Application, force bool) func(resultFilePath, namespace string) error {

	logFields := map[string]interface{}{
		"command": "apply",
	}
	applyMonitor := monitor.WithFields(logFields)

	resultFunc := func(resultFilePath, namespace string) error {
		return desired.Apply(applyMonitor, resultFilePath, namespace, app.GetName(), force)
	}

	return resultFunc
}
