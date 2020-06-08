package ambassador

import (
	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	argocdnet "github.com/caos/orbos/internal/operator/boom/application/applications/argocd/network"
	grafananet "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/network"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/crds"
	"github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
)

func (a *Ambassador) HelmPreApplySteps(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) ([]interface{}, error) {

	ret := make([]interface{}, 0)
	if toolsetCRDSpec.Argocd.Network != nil {
		host := crds.GetHostFromConfig(argocdnet.GetHostConfig(toolsetCRDSpec.Argocd.Network))
		ret = append(ret, host)
		mapping := crds.GetMappingFromConfig(argocdnet.GetMappingConfig(toolsetCRDSpec.Argocd.Network))
		ret = append(ret, mapping)
	}

	if toolsetCRDSpec.Grafana.Network != nil {
		host := crds.GetHostFromConfig(grafananet.GetHostConfig(toolsetCRDSpec.Grafana.Network))
		ret = append(ret, host)
		mapping := crds.GetMappingFromConfig(grafananet.GetMappingConfig(toolsetCRDSpec.Grafana.Network))
		ret = append(ret, mapping)
	}

	return ret, nil
}

func (a *Ambassador) HelmMutate(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec, resultFilePath string) error {

	if err := helper.DeleteFirstResourceFromYaml(resultFilePath, "v1", "Pod", "ambassador-test-ready"); err != nil {
		return err
	}

	return nil
}

func (a *Ambassador) SpecToHelmValues(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) interface{} {
	spec := toolsetCRDSpec.Ambassador
	imageTags := helm.GetImageTags()

	values := helm.DefaultValues(imageTags)
	if spec.ReplicaCount != 0 {
		values.ReplicaCount = spec.ReplicaCount
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

	return values
}

func (a *Ambassador) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (a *Ambassador) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
