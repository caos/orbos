package helm

func DefaultValues(imageTags map[string]string) *Values {
	return &Values{
		FullnameOverride: "logging-operator",
		ReplicaCount:     1,
		Image: Image{
			Repository: "banzaicloud/logging-operator",
			Tag:        imageTags["banzaicloud/logging-operator"],
			PullPolicy: "IfNotPresent",
		},
		HTTP: HTTP{
			Port: 8080,
			Service: Service{
				Type: "ClusterIP",
			},
		},
		RBAC: RBAC{
			Enabled: true,
			PSP: PSP{
				Enabled: true,
			},
		},
	}
}
