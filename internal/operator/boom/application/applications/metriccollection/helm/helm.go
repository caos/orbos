package helm

import "github.com/caos/orbos/v5/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "prometheus-operator",
		Version: "9.3.1",
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"grafana/grafana":                           "7.1.5",
		"quay.io/prometheus/alertmanager":           "v0.21.0",
		"squareup/ghostunnel":                       "v1.5.2",
		"jettech/kube-webhook-certgen":              "v1.2.1",
		"quay.io/coreos/prometheus-operator":        "v0.38.2",
		"quay.io/coreos/configmap-reload":           "v0.0.1",
		"quay.io/coreos/prometheus-config-reloader": "v0.38.1",
		"k8s.gcr.io/hyperkube":                      "v1.12.1",
		"quay.io/prometheus/prometheus":             "v2.20.1",
	}
}
