package config

import (
	"strings"

	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"

	lokiinfo "github.com/caos/orbos/internal/operator/boom/application/applications/logspersisting/info"
	prometheusinfo "github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/info"
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

func New(spec *toolsetslatest.ToolsetSpec) *Config {
	dashboardProviders := make([]*Provider, 0)
	if spec.Monitoring != nil && spec.Monitoring.DashboardProviders != nil {
		for _, provider := range spec.Monitoring.DashboardProviders {
			confProvider := &Provider{
				ConfigMaps: provider.ConfigMaps,
				Folder:     provider.Folder,
			}
			dashboardProviders = append(dashboardProviders, confProvider)
		}
	}

	datasources := make([]*Datasource, 0)
	if spec.Monitoring != nil && spec.Monitoring.Datasources != nil {
		for _, datasource := range spec.Monitoring.Datasources {
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
		DashboardProviders: dashboardProviders,
		Datasources:        datasources,
	}
	if spec.Monitoring != nil {
		conf.Deploy = spec.Monitoring.Deploy
	}

	providers := getGrafanaDashboards(DashboardsDirectoryPath, spec)

	for _, provider := range providers {
		conf.AddDashboardProvider(provider)
	}

	if spec.MetricCollection != nil && spec.MetricsPersisting != nil && spec.MetricCollection.Deploy && spec.MetricsPersisting.Deploy {
		serviceName := strings.Join([]string{prometheusinfo.GetInstanceName(), "prometheus"}, "-")
		datasourceProm := strings.Join([]string{"http://", serviceName, ".", prometheusinfo.GetNamespace(), ":9090"}, "")
		conf.AddDatasourceURL(serviceName, "prometheus", datasourceProm)
	}

	if spec.LogsPersisting != nil && spec.LogsPersisting.Deploy {
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
