package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "minio-operator",
		Version: "4.0.11",
		Index: &chart.Index{
			Name: "minio",
			URL:  "operator.min.io",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"minio/console":  "v0.6.8",
		"minio/operator": "v4.0.10",
		"minio/minio":    "RELEASE.2021-04-06T23-11-00Z",
	}
}
