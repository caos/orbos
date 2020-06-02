package helm

import prometheusoperator "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/helm"

import "github.com/caos/orbos/internal/operator/boom/application/applications/grafanastandalone"

func DefaultValues(imageTags map[string]string) *Values {
	grafana := &GrafanaValues{
		Image: &grafanastandalone.Image{
			Repository: "grafana/grafana",
			Tag:        imageTags["grafana/grafana"],
			PullPolicy: "IfNotPresent",
		},
		FullnameOverride:         "grafana",
		Enabled:                  true,
		DefaultDashboardsEnabled: true,
		AdminPassword:            "admin",
		Ingress: &Ingress{
			Enabled: false,
		},
		Sidecar: &Sidecar{
			Dashboards: &Dashboards{
				Enabled: true,
				Label:   "grafana_dashboard",
			},
			Datasources: &Datasources{
				Enabled: true,
				Label:   "grafana_datasource",
			},
		},
		ServiceMonitor: &ServiceMonitor{
			SelfMonitor: false,
		},
		Persistence: &Persistence{
			Type:        "pvc",
			Enabled:     false,
			AccessModes: []string{"ReadWriteOnce"},
			Size:        "10Gi",
			Finalizers:  []string{"kubernetes.io/pvc-protection"},
		},
		TestFramework: &grafanastandalone.TestFramework{
			Enabled: false,
			Image:   "dduportal/bats",
			Tag:     imageTags["dduportal/bats"],
		},
		Plugins: []string{"grafana-piechart-panel"},
		Ini: &Ini{
			Paths: map[string]string{
				"data":         "/var/lib/grafana/data",
				"logs":         "/var/log/grafana",
				"plugins":      "/var/lib/grafana/plugins",
				"provisioning": "/etc/grafana/provisioning",
			},
			Analytics: map[string]bool{
				"check_for_updates": true,
			},
			Log: map[string]string{
				"mode": "console",
			},
			GrafanaNet: map[string]interface{}{
				"url": "https://grafana.net",
			},
		},
		Env: map[string]string{
			"GF_SERVER_ROOT_URL": "%(protocol)s://%(domain)s/",
		},
	}

	return &Values{
		DefaultRules: &DefaultRules{
			Create: false,
			Rules: &Rules{
				Alertmanager:                false,
				Etcd:                        false,
				General:                     false,
				K8S:                         false,
				KubeApiserver:               false,
				KubePrometheusNodeAlerting:  false,
				KubePrometheusNodeRecording: false,
				KubernetesAbsent:            false,
				KubernetesApps:              false,
				KubernetesResources:         false,
				KubernetesStorage:           false,
				KubernetesSystem:            false,
				KubeScheduler:               false,
				Network:                     false,
				Node:                        false,
				Prometheus:                  false,
				PrometheusOperator:          false,
				Time:                        false,
			},
		},
		Global: &Global{
			Rbac: &Rbac{
				Create:     false,
				PspEnabled: false,
			},
		},
		FullnameOverride: "grafana",
		Alertmanager: &DisabledToolServicePerReplica{
			Enabled:           false,
			ServicePerReplica: &DisabledTool{Enabled: false},
			IngressPerReplica: &DisabledTool{Enabled: false},
		},
		Grafana: grafana,
		KubeAPIServer: &DisabledTool{
			Enabled: false,
		},
		Kubelet: &DisabledTool{
			Enabled: false,
		},
		KubeControllerManager: &DisabledTool{
			Enabled: false,
		},
		CoreDNS: &DisabledTool{
			Enabled: false,
		},
		KubeDNS: &DisabledTool{
			Enabled: false,
		},
		KubeEtcd: &DisabledTool{
			Enabled: false,
		},
		KubeScheduler: &DisabledTool{
			Enabled: false,
		},
		KubeProxy: &DisabledTool{
			Enabled: false,
		},
		KubeStateMetricsScrap: &DisabledTool{
			Enabled: false,
		},
		KubeStateMetrics: &DisabledTool{
			Enabled: false,
		},
		NodeExporter: &DisabledTool{
			Enabled: false,
		},
		PrometheusNodeExporter: &DisabledTool{
			Enabled: false,
		},
		PrometheusOperator: &prometheusoperator.PrometheusOperatorValues{
			Enabled: false,
			TLSProxy: &prometheusoperator.TLSProxy{
				Enabled: false,
				Image: &prometheusoperator.Image{
					Repository: "squareup/ghostunnel",
					Tag:        imageTags["squareup/ghostunnel"],
					PullPolicy: "IfNotPresent",
				},
			},
			AdmissionWebhooks: &prometheusoperator.AdmissionWebhooks{
				FailurePolicy: "Fail",
				Enabled:       false,
				Patch: &prometheusoperator.Patch{
					Enabled: false,
					Image: &prometheusoperator.Image{
						Repository: "jettech/kube-webhook-certgen",
						Tag:        imageTags["jettech/kube-webhook-certgen"],
						PullPolicy: "IfNotPresent",
					},
					PriorityClassName: "",
				},
			},
			ServiceAccount: &prometheusoperator.ServiceAccount{
				Create: false,
			},
			ServiceMonitor: &prometheusoperator.ServiceMonitor{
				Interval:    "",
				SelfMonitor: false,
			},
			CreateCustomResource: true,
			KubeletService: &prometheusoperator.KubeletService{
				Enabled: false,
			},
		},
		Prometheus: &DisabledToolServicePerReplica{
			Enabled:           false,
			ServicePerReplica: &DisabledTool{Enabled: false},
			IngressPerReplica: &DisabledTool{Enabled: false},
		},
	}
}
