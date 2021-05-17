package helm

func DefaultValues(imageTags map[string]string, image string) *Values {
	return &Values{
		TestImage: &Image{
			Registry:   "docker.io",
			Repository: image,
			Tag:        imageTags[image],
			PullPolicy: "IfNotPresent",
		},
		Image: &Image{
			Registry:   "docker.io",
			Repository: image,
			Tag:        imageTags[image],
			PullPolicy: "IfNotPresent",
		},
		HostAliases: []string{},
		Replicas:    1,
		Rbac: &Rbac{
			Create: true,
		},
		ServiceAccount: &ServiceAccount{
			Create:                       true,
			Name:                         "",
			AutomountServiceAccountToken: true,
		},
		APIService:  &APIService{Create: false},
		SecurePort:  "8443",
		HostNetwork: &HostNetwork{Enabled: false},
		Command:     []string{"metrics-server"},
		ExtraArgs: []string{
			"--kubelet-insecure-tls",
			"--kubelet-preferred-address-types=InternalIP,Hostname,InternalDNS,ExternalDNS,ExternalIP",
		},
		PodLabels:             map[string]string{},
		PodAnnotations:        map[string]string{},
		PodAffinityPreset:     "",
		PodAntiAffinityPreset: "soft",
		PodDisruptionBudget: &PodDisruptionBudget{
			Enabled:        false,
			MinAvailable:   nil,
			MaxUnavailable: nil,
		},
		NodeAffinityPreset: &NodeAffinityPreset{
			Type:   "",
			Key:    "",
			Values: []string{},
		},
		Affinity:                  struct{}{},
		TopologySpreadConstraints: []string{},
		NodeSelector:              struct{}{},
		Tolerations:               nil,
		Service: &Service{
			Annotations: map[string]string{},
			Labels:      map[string]string{},
			Port:        443,
			Type:        "ClusterIP",
		},
		Resources: struct{}{},
		LivenessProbe: &Probe{
			Enabled:          true,
			FailureThreshold: 3,
			HTTPGet: &HTTPGet{
				Path:   "/livez",
				Port:   "https",
				Scheme: "HTTPS",
			},
			PeriodSeconds: 10,
		},
		ReadinessProbe: &Probe{
			Enabled:          true,
			FailureThreshold: 3,
			HTTPGet: &HTTPGet{
				Path:   "/readyz",
				Port:   "https",
				Scheme: "HTTPS",
			},
			PeriodSeconds: 10,
		},
		ContainerSecurityContext: &ContainerSecurityContext{
			Enabled:                true,
			ReadOnlyRootFilesystem: false,
			RunAsNonRoot:           true,
		},
		PodSecurityContext: &PodSecurityContext{
			Enabled: false,
		},
	}
}
