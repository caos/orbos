package helm

import "github.com/caos/orbos/v5/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "kube-state-metrics",
		Version: "2.9.1",
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"quay.io/coreos/kube-state-metrics": "v1.9.7",
	}
}
