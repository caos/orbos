package ambassador

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/crds"
	"github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/helm"
	argocdnet "github.com/caos/orbos/internal/operator/boom/application/applications/argocd/network"
	grafananet "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/network"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"
)

func (a *Ambassador) HelmPreApplySteps(monitor mntr.Monitor, toolsetCRDSpec *toolsetslatest.ToolsetSpec) ([]interface{}, error) {

	ret := make([]interface{}, 0)
	if toolsetCRDSpec.Reconciling != nil && toolsetCRDSpec.Reconciling.Network != nil {
		host := crds.GetHostFromConfig(argocdnet.GetHostConfig(toolsetCRDSpec.Reconciling.Network))
		ret = append(ret, host)
		mapping := crds.GetMappingFromConfig(argocdnet.GetMappingConfig(toolsetCRDSpec.Reconciling.Network))
		ret = append(ret, mapping)
	}

	if toolsetCRDSpec.Monitoring != nil && toolsetCRDSpec.Monitoring.Network != nil {
		host := crds.GetHostFromConfig(grafananet.GetHostConfig(toolsetCRDSpec.Monitoring.Network))
		ret = append(ret, host)
		mapping := crds.GetMappingFromConfig(grafananet.GetMappingConfig(toolsetCRDSpec.Monitoring.Network))
		ret = append(ret, mapping)
	}

	return ret, nil
}

func (a *Ambassador) HelmMutate(monitor mntr.Monitor, toolsetCRDSpec *toolsetslatest.ToolsetSpec, resultFilePath string) error {

	if err := helper.DeleteFirstResourceFromYaml(resultFilePath, "v1", "Pod", "ambassador-test-ready"); err != nil {
		return err
	}

	return nil
}

func (a *Ambassador) SpecToHelmValues(monitor mntr.Monitor, toolsetCRDSpec *toolsetslatest.ToolsetSpec) interface{} {
	imageTags := helm.GetImageTags()
	if toolsetCRDSpec != nil && toolsetCRDSpec.APIGateway != nil {
		helper.OverwriteExistingValues(imageTags, map[string]string{
			"quay.io/datawire/aes": toolsetCRDSpec.APIGateway.OverwriteVersion,
		})
	}
	values := helm.DefaultValues(imageTags)

	spec := toolsetCRDSpec.APIGateway

	if spec == nil {
		return values
	}

	if spec.ReplicaCount != 0 {
		values.ReplicaCount = spec.ReplicaCount
	}

	if spec.Affinity != nil {
		values.Affinity = spec.Affinity
	}

	if spec.Service != nil {
		values.Service.Type = spec.Service.Type
		values.Service.LoadBalancerIP = spec.Service.LoadBalancerIP
		if spec.Service.Ports != nil {
			ports := make([]*helm.Port, 0)
			for _, v := range spec.Service.Ports {
				ports = append(ports, &helm.Port{
					Name:       v.Name,
					Port:       v.Port,
					TargetPort: v.TargetPort,
					NodePort:   v.NodePort,
				})
			}
			values.Service.Ports = ports
		}
	}

	if spec.NodeSelector != nil {
		for k, v := range spec.NodeSelector {
			values.NodeSelector[k] = v
			values.Redis.NodeSelector[k] = v
		}
	}

	if spec.Tolerations != nil {
		for _, tol := range spec.Tolerations {
			values.Tolerations = append(values.Tolerations, tol)
		}
	}

	values.CreateDevPortalMapping = spec.ActivateDevPortal

	if spec.Resources != nil {
		values.Resources = spec.Resources
	}

	// default is false
	values.Service.Annotations.Module.Config.EnableGRPCWeb = spec.GRPCWeb
	// default is true
	if spec.ProxyProtocol != nil {
		values.Service.Annotations.Module.Config.UseProxyProto = *spec.ProxyProtocol
	}

	if spec.Caching != nil {
		if spec.Caching.Enable {
			values.Redis.Create = true
		}

		if spec.Caching.Resources != nil {
			values.Redis.Resources = spec.Caching.Resources
		}
	}

	licenceKey, err := helper.GetSecretValue(spec.LicenceKey, spec.ExistingLicenceKey)
	if err != nil {
		monitor.Debug("No licence key found")
		return values
	}
	if licenceKey != "" {
		values.LicenseKey.Value = licenceKey
	}

	return values
}

func (a *Ambassador) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (a *Ambassador) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
