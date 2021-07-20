package config

import "github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/helm"

func getScrapeConfigs() []*helm.AdditionalScrapeConfig {

	adconfigs := make([]*helm.AdditionalScrapeConfig, 0)

	adconfigs = append(adconfigs, getNodes())
	adconfigs = append(adconfigs, getCadvisor())
	adconfigs = append(adconfigs, getEtcd())

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

func getEtcd() *helm.AdditionalScrapeConfig {

	sdconfig := &helm.KubernetesSdConfig{
		Role: "node",
	}

	relabelings := []*helm.RelabelConfig{{
		Action:       "replace",
		SourceLabels: []string{"__address__"},
		TargetLabel:  "__address__",
		Regex:        "(.*):.*",
		Replacement:  "${1}:2381",
	}, {
		Action:       "keep",
		Regex:        "true",
		SourceLabels: []string{"__meta_kubernetes_node_labelpresent_node_role_kubernetes_io_master"},
	}}

	metricRelabelConfigs := []*helm.RelabelConfig{{
		Action:       "keep",
		Regex:        "etcd_server_has_leader",
		SourceLabels: []string{"__name__"},
	}, {
		Action:       "replace",
		SourceLabels: []string{"__name__"},
		TargetLabel:  "__name__",
		Regex:        "etcd_server_has_leader",
		Replacement:  "dist_etcd_server_has_leader",
	}}

	return &helm.AdditionalScrapeConfig{
		JobName:              "caos_remote_etcd",
		Scheme:               "http",
		KubernetesSdConfigs:  []*helm.KubernetesSdConfig{sdconfig},
		RelabelConfigs:       relabelings,
		MetricRelabelConfigs: metricRelabelConfigs,
	}
}
