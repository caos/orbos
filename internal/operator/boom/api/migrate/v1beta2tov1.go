package migrate

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/secret"
)

func V1beta2Tov1(oldToolset *v1beta2.Toolset) (newToolset *latest.Toolset, secrets map[string]*secret.Secret) {

	defer func() {
		secrets = latest.GetSecretsMap(newToolset)
	}()

	newToolset = &latest.Toolset{
		APIVersion: "boom.caos.ch/v1",
		Metadata: &latest.Metadata{
			Name:      oldToolset.Metadata.Name,
			Namespace: oldToolset.Metadata.Namespace,
		},
		Kind: "Toolset",
		Spec: &latest.ToolsetSpec{
			Boom:               oldToolset.Spec.Boom,
			ForceApply:         oldToolset.Spec.ForceApply,
			CurrentStateFolder: oldToolset.Spec.CurrentStateFolder,
			PreApply:           oldToolset.Spec.PreApply,
			PostApply:          oldToolset.Spec.PostApply,
			MetricCollection:   oldToolset.Spec.MetricCollection,
			//			LogCollection:          oldToolset.Spec.LogCollection,
			NodeMetricsExporter:    oldToolset.Spec.NodeMetricsExporter,
			SystemdMetricsExporter: oldToolset.Spec.SystemdMetricsExporter,
			Monitoring:             oldToolset.Spec.Monitoring,
			APIGateway:             oldToolset.Spec.APIGateway,
			KubeMetricsExporter:    oldToolset.Spec.KubeMetricsExporter,
			Reconciling:            oldToolset.Spec.Reconciling,
			MetricsPersisting:      oldToolset.Spec.MetricsPersisting,
			LogsPersisting:         oldToolset.Spec.LogsPersisting,
			MetricsServer:          oldToolset.Spec.MetricsServer,
		},
	}

	if oldToolset.Spec.LogCollection == nil {
		return newToolset, nil
	}
	newToolset.Spec.LogCollection = &latest.LogCollection{
		Deploy: oldToolset.Spec.LogCollection.Deploy,
	}

	if oldToolset.Spec.LogCollection.FluentdPVC != nil {
		newToolset.Spec.LogCollection.Fluentd = &latest.Fluentd{
			PVC: oldToolset.Spec.LogCollection.FluentdPVC,
		}
	}

	if oldToolset.Spec.LogCollection.Resources == nil &&
		oldToolset.Spec.LogCollection.Tolerations == nil &&
		oldToolset.Spec.LogCollection.NodeSelector == nil {
		return newToolset, nil
	}

	newToolset.Spec.LogCollection.Operator = &latest.Component{
		NodeSelector: oldToolset.Spec.LogCollection.NodeSelector,
		Tolerations:  oldToolset.Spec.LogCollection.Tolerations,
		Resources:    oldToolset.Spec.LogCollection.Resources,
	}
	return newToolset, nil
}
