package apiserver

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	monitorlabels := labels.GetMonitorLabels(instanceName, "prometheus")

	metricRelabelings := make([]*servicemonitor.ConfigRelabeling, 0)
	relabeling := &servicemonitor.ConfigRelabeling{
		Action:       "keep",
		Regex:        "default;kubernetes;https",
		SourceLabels: []string{"__meta_kubernetes_namespace", "__meta_kubernetes_service_name", "__meta_kubernetes_endpoint_port_name"},
	}
	metricRelabelings = append(metricRelabelings, relabeling)

	endpoints := make([]*servicemonitor.ConfigEndpoint, 0)
	endpoint := &servicemonitor.ConfigEndpoint{
		Scheme:          "https",
		BearerTokenFile: "/var/run/secrets/kubernetes.io/serviceaccount/token",
		Port:            "https",
		Path:            "/metrics",
		TLSConfig: &servicemonitor.ConfigTLSConfig{
			CaFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
		MetricRelabelings: metricRelabelings,
	}
	endpoints = append(endpoints, endpoint)

	labels := map[string]string{
		"component": "apiserver",
		"provider":  "kubernetes",
	}

	return &servicemonitor.Config{
		Name:                  "kubernetes-apiservers-servicemonitor",
		Endpoints:             endpoints,
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: labels,
		NamespaceSelector:     []string{"default"},
		JobName:               "kubernetes-apiservers",
	}
}
