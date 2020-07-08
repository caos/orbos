package grafana

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/admin"
	"path/filepath"
	"sort"

	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/auth"
	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/config"
	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/internal/utils/kubectl"
	"github.com/caos/orbos/internal/utils/kustomize"
	"github.com/caos/orbos/mntr"
)

func (g *Grafana) HelmMutate(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec, resultFilePath string) error {
	if toolsetCRDSpec.KubeMetricsExporter != nil && toolsetCRDSpec.KubeMetricsExporter.Deploy &&
		toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.KubeStateMetrics) {

		if err := helper.DeleteFirstResourceFromYaml(resultFilePath, "v1", "ConfigMap", "grafana-persistentvolumesusage"); err != nil {
			return err
		}
	}

	return nil
}

func (g *Grafana) HelmPreApplySteps(monitor mntr.Monitor, spec *toolsetsv1beta2.ToolsetSpec) ([]interface{}, error) {
	config := config.New(spec)

	folders := make([]string, 0)
	for _, provider := range config.DashboardProviders {
		folders = append(folders, provider.Folder)
	}

	outs, err := getKustomizeOutput(folders)
	if err != nil {
		return nil, err
	}

	ret := make([]interface{}, len(outs))
	for k, v := range outs {
		ret[k] = v
	}

	if spec.Monitoring != nil && spec.Monitoring.Admin != nil {
		ret = append(ret, admin.GetSecrets(spec.Monitoring.Admin)...)
	}
	return ret, nil
}

type ProviderSorter []*helm.Provider

func (a ProviderSorter) Len() int           { return len(a) }
func (a ProviderSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ProviderSorter) Less(i, j int) bool { return a[i].Name < a[j].Name }

type AlphaSorter []string

func (a AlphaSorter) Len() int           { return len(a) }
func (a AlphaSorter) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a AlphaSorter) Less(i, j int) bool { return a[i] < a[j] }

func (g *Grafana) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta2.ToolsetSpec) interface{} {
	version, err := kubectl.NewVersion().GetKubeVersion(monitor)
	if err != nil {
		return nil
	}

	values := helm.DefaultValues(g.GetImageTags())
	conf := config.New(toolset)

	values.KubeTargetVersionOverride = version

	providers := make([]*helm.Provider, 0)
	dashboards := make(map[string]string, 0)
	datasources := make([]*helm.Datasource, 0)

	//internal datasources
	if conf.Datasources != nil {
		for _, datasource := range conf.Datasources {
			valuesDatasource := &helm.Datasource{
				Name:      datasource.Name,
				Type:      datasource.Type,
				URL:       datasource.Url,
				Access:    datasource.Access,
				IsDefault: datasource.IsDefault,
			}
			datasources = append(datasources, valuesDatasource)
		}
	}

	//internal dashboards
	if conf.DashboardProviders != nil {
		for _, provider := range conf.DashboardProviders {
			sort.Sort(AlphaSorter(provider.ConfigMaps))
			for _, configmap := range provider.ConfigMaps {
				providers = append(providers, getProvider(configmap))
				dashboards[configmap] = configmap
			}
		}
	}

	if len(providers) > 0 {
		sort.Sort(ProviderSorter(providers))
		values.Grafana.DashboardProviders = &helm.DashboardProviders{
			Providers: &helm.Providersyaml{
				APIVersion: 1,
				Providers:  providers,
			},
		}

		values.Grafana.DashboardsConfigMaps = dashboards
	}
	if len(datasources) > 0 {
		values.Grafana.AdditionalDataSources = datasources
	}

	if toolset.Monitoring != nil {
		if toolset.Monitoring.Admin != nil {
			values.Grafana.Admin = admin.GetConfig(toolset.Monitoring.Admin)
		}

		if toolset.Monitoring.Storage != nil {
			values.Grafana.Persistence.Enabled = true
			values.Grafana.Persistence.Size = toolset.Monitoring.Storage.Size
			values.Grafana.Persistence.StorageClassName = toolset.Monitoring.Storage.StorageClass

			if toolset.Monitoring.Storage.AccessModes != nil {
				values.Grafana.Persistence.AccessModes = toolset.Monitoring.Storage.AccessModes
			}
		}

		if toolset.Monitoring.Network != nil && toolset.Monitoring.Network.Domain != "" {
			values.Grafana.Env["GF_SERVER_DOMAIN"] = toolset.Monitoring.Network.Domain

			if toolset.Monitoring.Auth != nil {
				if toolset.Monitoring.Auth.Google != nil {
					google, err := auth.GetGoogleAuthConfig(toolset.Monitoring.Auth.Google)
					if err == nil && google != nil {
						values.Grafana.Ini.AuthGoogle = google
					}
				}

				if toolset.Monitoring.Auth.Github != nil {
					github, err := auth.GetGithubAuthConfig(toolset.Monitoring.Auth.Github)
					if err == nil && github != nil {
						values.Grafana.Ini.AuthGithub = github
					}
				}

				if toolset.Monitoring.Auth.Gitlab != nil {
					gitlab, err := auth.GetGitlabAuthConfig(toolset.Monitoring.Auth.Gitlab)
					if err == nil && gitlab != nil {
						values.Grafana.Ini.AuthGitlab = gitlab
					}
				}

				if toolset.Monitoring.Auth.GenericOAuth != nil {
					generic, err := auth.GetGenericOAuthConfig(toolset.Monitoring.Auth.GenericOAuth)
					if err == nil && generic != nil {
						values.Grafana.Ini.AuthGeneric = generic
					}
				}
			}
		}

		if toolset.Monitoring.Plugins != nil && len(toolset.Monitoring.Plugins) > 0 {
			values.Grafana.Plugins = append(values.Grafana.Plugins, toolset.Monitoring.Plugins...)
		}
	}

	return values
}

func getKustomizeOutput(folders []string) ([]string, error) {
	ret := make([]string, len(folders))
	for n, folder := range folders {

		cmd, err := kustomize.New(folder)
		if err != nil {
			return nil, err
		}
		execcmd := cmd.Build()

		out, err := execcmd.Output()
		if err != nil {
			return nil, err
		}
		ret[n] = string(out)
	}
	return ret, nil
}

func getProvider(configmapName string) *helm.Provider {
	return &helm.Provider{
		Name:            configmapName,
		Type:            "file",
		DisableDeletion: false,
		Editable:        false,
		Options: map[string]string{
			"path": filepath.Join("/var/lib/grafana/dashboards", configmapName),
		},
	}
}

func (g *Grafana) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (g *Grafana) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
