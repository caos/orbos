package helm

import "github.com/caos/orbos/internal/operator/boom/templator/helm/chart"

func GetChartInfo() *chart.Chart {
	return &chart.Chart{
		Name:    "prometheus-operator",
		Version: "8.10.2",
	}
}

func GetImageTags() map[string]string {
	return map[string]string{
		"quay.io/prometheus/alertmanager":           "v0.20.0",
		"squareup/ghostunnel":                       "v1.5.2",
		"jettech/kube-webhook-certgen":              "v1.0.0",
		"quay.io/coreos/prometheus-operator":        "v0.36.0",
		"quay.io/coreos/configmap-reload":           "v0.0.1",
		"quay.io/coreos/prometheus-config-reloader": "v0.36.0",
		"k8s.gcr.io/hyperkube":                      "v1.12.1",
		"quay.io/prometheus/prometheus":             "v2.15.2",
	}
}
