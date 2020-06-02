package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "metrics-server",
		Version: "2.11.1",
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"k8s.gcr.io/metrics-server-amd64": "v0.3.6",
	}
}
