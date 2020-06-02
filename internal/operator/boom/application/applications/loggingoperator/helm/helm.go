package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "logging-operator",
		Version: "3.2.1",
		Index: &chart.Index{
			Name: "banzaicloud-stable",
			URL:  "kubernetes-charts.banzaicloud.com",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"banzaicloud/logging-operator": "3.2.0",
	}
}
