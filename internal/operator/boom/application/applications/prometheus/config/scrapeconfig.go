package config

import "github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/helm"

func getScrapeConfigs() []*helm.AdditionalScrapeConfig {

	adconfigs := make([]*helm.AdditionalScrapeConfig, 0)

	adconfigs = append(adconfigs, getNodes())
	adconfigs = append(adconfigs, getCadvisor())

	return adconfigs

}

func getNodes() *helm.AdditionalScrapeConfig {
	relabelings := make([]*helm.RelabelConfig, 0)
	relabeling := &helm.RelabelConfig{
		Action: "labelmap",
		Regex:  "__meta_kubernetes_node_label_(.+)",
	}
	relabelings = append(relabelings, relabeling)
	relabeling = &helm.RelabelConfig{
		TargetLabel: "__address__",
		Replacement: "kubernetes.default.svc:443",
	}
	relabelings = append(relabelings, relabeling)
	relabeling = &helm.RelabelConfig{
		SourceLabels: []string{"__meta_kubernetes_node_name"},
		Regex:        "(.+)",
		TargetLabel:  "__metrics_path__",
		Replacement:  "/api/v1/nodes/${1}/proxy/metrics",
	}
	relabelings = append(relabelings, relabeling)

	relabeling = &helm.RelabelConfig{
		TargetLabel: "metrics_path",
		Replacement: "/metrics",
	}
	relabelings = append(relabelings, relabeling)

	sdconfig := &helm.KubernetesSdConfig{
		Role: "node",
	}

	return &helm.AdditionalScrapeConfig{
		JobName:             "kubernetes-nodes",
		Scheme:              "https",
		KubernetesSdConfigs: []*helm.KubernetesSdConfig{sdconfig},
		BearerTokenFile:     "/var/run/secrets/kubernetes.io/serviceaccount/token",
		TLSConfig: &helm.TLSConfig{
			CaFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
		RelabelConfigs: relabelings,
	}
}

func getCadvisor() *helm.AdditionalScrapeConfig {
	relabelings := make([]*helm.RelabelConfig, 0)
	relabeling := &helm.RelabelConfig{
		Action: "labelmap",
		Regex:  "__meta_kubernetes_node_label_(.+)",
	}
	relabelings = append(relabelings, relabeling)
	relabeling = &helm.RelabelConfig{
		TargetLabel: "__address__",
		Replacement: "kubernetes.default.svc:443",
	}
	relabelings = append(relabelings, relabeling)
	relabeling = &helm.RelabelConfig{
		SourceLabels: []string{"__meta_kubernetes_node_name"},
		Regex:        "(.+)",
		TargetLabel:  "__metrics_path__",
		Replacement:  "/api/v1/nodes/${1}/proxy/metrics/cadvisor",
	}
	relabelings = append(relabelings, relabeling)

	relabeling = &helm.RelabelConfig{
		TargetLabel: "metrics_path",
		Replacement: "/metrics/cadvisor",
	}
	relabelings = append(relabelings, relabeling)

	sdconfig := &helm.KubernetesSdConfig{
		Role: "node",
	}

	return &helm.AdditionalScrapeConfig{
		JobName:             "kubelet",
		Scheme:              "https",
		KubernetesSdConfigs: []*helm.KubernetesSdConfig{sdconfig},
		BearerTokenFile:     "/var/run/secrets/kubernetes.io/serviceaccount/token",
		TLSConfig: &helm.TLSConfig{
			CaFile: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt",
		},
		RelabelConfigs: relabelings,
	}
}
