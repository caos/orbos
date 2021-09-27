package helm

import "github.com/caos/orbos/v5/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "ambassador",
		Version: "6.5.9",
		Index: &chart.Index{
			Name: "datawire",
			URL:  "www.getambassador.io",
		},
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"quay.io/datawire/aes": "1.8.0",
		"prom/statsd-exporter": "v0.8.1",
	}
}
