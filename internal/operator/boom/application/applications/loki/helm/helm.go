package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "loki",
		Version: "0.29.0",
		Index: &chart.Index{
			Name: "loki",
			URL:  "grafana.github.io/loki/charts",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"grafana/loki": "1.5.0",
	}
}
