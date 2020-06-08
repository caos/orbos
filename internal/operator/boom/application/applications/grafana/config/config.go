package config

import (
	"strings"

	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	lokiinfo "github.com/caos/orbos/internal/operator/boom/application/applications/loki/info"
	prometheusinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/info"
)

var (
	DashboardsDirectoryPath string = "../../dashboards"
)

type Datasource struct {
	Name      string
	Type      string
	Url       string
	Access    string
	IsDefault bool
}

type Provider struct {
	ConfigMaps []string
	Folder     string
}

type Config struct {
	Deploy             bool
	Datasources        []*Datasource
	DashboardProviders []*Provider
	Ini                map[string]interface{}
}

func New(spec *toolsetsv1beta1.ToolsetSpec) *Config {
	dashboardProviders := make([]*Provider, 0)
	if spec.Grafana.DashboardProviders != nil {
		for _, provider := range spec.Grafana.DashboardProviders {
			confProvider := &Provider{
				ConfigMaps: provider.ConfigMaps,
				Folder:     provider.Folder,
			}
			dashboardProviders = append(dashboardProviders, confProvider)
		}
	}

	datasources := make([]*Datasource, 0)
	if spec.Grafana.Datasources != nil {
		for _, datasource := range spec.Grafana.Datasources {
			confDatasource := &Datasource{
				Name:      datasource.Name,
				Type:      datasource.Type,
				Url:       datasource.Url,
				Access:    datasource.Access,
				IsDefault: datasource.IsDefault,
			}
			datasources = append(datasources, confDatasource)
		}
	}

	conf := &Config{
		Deploy:             spec.Grafana.Deploy,
		DashboardProviders: dashboardProviders,
		Datasources:        datasources,
	}

	providers := getGrafanaDashboards(DashboardsDirectoryPath, spec)

	for _, provider := range providers {
		conf.AddDashboardProvider(provider)
	}

	if spec.PrometheusOperator.Deploy && spec.Prometheus.Deploy {
		serviceName := strings.Join([]string{prometheusinfo.GetInstanceName(), "prometheus"}, "-")
		datasourceProm := strings.Join([]string{"http://", serviceName, ".", prometheusinfo.GetNamespace(), ":9090"}, "")
		conf.AddDatasourceURL(serviceName, "prometheus", datasourceProm)
	}

	if spec.Loki.Deploy {
		serviceName := lokiinfo.GetName().String()
		datasourceLoki := strings.Join([]string{"http://", serviceName, ".", lokiinfo.GetNamespace(), ":3100"}, "")
		conf.AddDatasourceURL(serviceName, "loki", datasourceLoki)
	}

	return conf
}

func (c *Config) AddDashboardProvider(provider *Provider) {
	c.DashboardProviders = append(c.DashboardProviders, provider)
}

func (c *Config) AddDatasourceURL(name, datatype, url string) {
	c.Datasources = append(c.Datasources, &Datasource{
		Name:   name,
		Type:   datatype,
		Url:    url,
		Access: "proxy",
	})
}
