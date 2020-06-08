package helm

func DefaultValues(imageTags map[string]string) *Values {
	return &Values{
		FullnameOverride: "node-exporter",
		Image: &Image{
			Repository: "quay.io/prometheus/node-exporter",
			Tag:        imageTags["quay.io/prometheus/node-exporter"],
			PullPolicy: "IfNotPresent",
		},
		Service: &Service{
			Type:        "ClusterIP",
			Port:        9100,
			TargetPort:  9100,
			NodePort:    "",
			Annotations: map[string]string{"prometheus.io/scrape": "true"},
		},
		Prometheus: &Prometheus{
			Monitor: &Monitor{
				Enabled:          false,
				AdditionalLabels: map[string]string{},
				Namespace:        "",
				ScrapeTimeout:    "10s",
			},
		},
		ServiceAccount: &ServiceAccount{
			Create: true,
		},
		SecurityContext: &SecurityContext{
			RunAsNonRoot: true,
			RunAsUser:    65534,
		},
		Rbac: &Rbac{
			Create:     true,
			PspEnabled: true,
		},
		HostNetwork: false,
		Tolerations: []*Toleration{&Toleration{
			Effect:   "NoSchedule",
			Operator: "Exists",
		}},
	}
}
