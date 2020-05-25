package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "logging-operator",
		Version: "2.7.2",
		Index: &chart.Index{
			Name: "banzaicloud-stable",
			URL:  "kubernetes-charts.banzaicloud.com",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"banzaicloud/logging-operator": "2.7.0",
	}
}
