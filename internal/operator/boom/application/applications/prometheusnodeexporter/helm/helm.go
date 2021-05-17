package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "prometheus-node-exporter",
		Version: "1.11.1",
		Index: &chart.Index{
			Name: "prometheus-community",
			URL:  "prometheus-community.github.io/helm-charts",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"quay.io/prometheus/node-exporter": "v1.0.1",
	}
}
