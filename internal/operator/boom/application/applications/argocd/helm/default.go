package helm

import (
	"strings"

	"github.com/caos/orbos/pkg/kubernetes/k8s"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func DefaultDexValues(imageTags map[string]string) *Dex {
	return &Dex{
		Enabled: true,
		Name:    "dex-server",
		Image: &Image{
			Repository:      "quay.io/dexidp/dex",
			Tag:             imageTags["quay.io/dexidp/dex"],
			ImagePullPolicy: "IfNotPresent",
		},
		ServiceAccount: &ServiceAccount{
			Create: true,
			Name:   "argocd-dex-server",
		},
		VolumeMounts: []*VolumeMount{
			&VolumeMount{
				Name:      "static-files",
				MountPath: "/shared",
			},
		},
		Volumes: []*Volume{
			&Volume{
				Name:     "static-files",
				EmptyDir: struct{}{},
			},
		},
		NodeSelector:      map[string]string{},
		ContainerPortHTTP: 5556,
		ServicePortHTTP:   5556,
		ContainerPortGrpc: 5557,
		ServicePortGrpc:   5557,
	}
}

func DefaultValues(imageTags map[string]string) *Values {
	knownHosts := []string{
		"bitbucket.org ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAubiN81eDcafrgMeLzaFPsw2kNvEcqTKl/VqLat/MaB33pZy0y3rJZtnqwR2qOOvbwKZYKiEO1O6VqNEBxKvJJelCq0dTXWT5pbO2gDXC6h6QDXCaHo6pOHGPUy+YBaGQRGuSusMEASYiWunYN0vCAI8QaXnWMXNMdFP3jHAJH0eDsoiGnLPBlBp4TNm6rYI74nMzgz3B9IikW4WVK+dc8KZJZWYjAuORU3jc1c/NPskD2ASinf8v3xnfXeukU0sJ5N6m5E8VLjObPEO+mN2t/FZTMZLiFqPWc/ALSqnMnnhwrNi2rbfg/rd/IpL8Le3pSBne8+seeFVBoGqzHM9yXw==",
		"github.com ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ==",
		"gitlab.com ecdsa-sha2-nistp256 AAAAE2VjZHNhLXNoYTItbmlzdHAyNTYAAAAIbmlzdHAyNTYAAABBBFSMqzJeV9rUzU4kWitGjeR4PWSa29SPqJ1fVkhtj3Hw9xjLVXVYrU9QlYWrOLXBpQ6KWjbjTDTdDkoohFzgbEY=",
		"gitlab.com ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIAfuCHKVTjquxvt6CM6tdG4SLp1Btn/nOeHHE5UOzRdf",
		"gitlab.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQCsj2bNKTBSpIYDEGk9KxsGh3mySTRgMtXL583qmBpzeQ+jqCMRgBqB98u3z++J1sKlXHWfM9dyhSevkMwSbhoR8XIq/U0tCNyokEi/ueaBMCvbcTHhO7FcwzY92WK4Yt0aGROY5qX2UKSeOvuP4D6TPqKF1onrSzH9bx9XUf2lEdWT/ia1NEKjunUqu1xOB/StKDHMoX4/OKyIzuS0q/T1zOATthvasJFoPrAjkohTyaDUz2LN5JoH839hViyEG82yB+MjcFV5MU3N1l1QL3cVUCh93xSaua1N85qivl+siMkPGbO5xR/En4iEY6K2XPASUEMaieWVNTRCtJ4S8H+9",
		"ssh.dev.azure.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H",
		"vs-ssh.visualstudio.com ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQC7Hr1oTWqNqOlzGJOfGJ4NakVyIzf1rXYd4d7wo6jBlkLvCA4odBlL0mDUyZ0/QUfTTqeu+tm22gOsv+VrVTMk6vwRU75gY/y9ut5Mb3bR5BV58dKXyq9A9UeB5Cakehn5Zgm6x1mKoVyf+FFn26iYqXJRgzIZZcZ5V6hrE0Qg39kZm4az48o0AUbf6Sp4SLdvnuMa2sVNwHBboS7EJkm57XQPVU3/QpyNLHbWDdzwtrlS+ez30S3AdYhLKEOxAG8weOnyrtLJAUen9mTkol8oII1edf7mWWbWVf0nBmly21+nZcmCTISQBtdcyPaEno7fFQMDD26/s0lfKob4Kw8H",
	}

	knownHostsStr := ""
	for _, v := range knownHosts {
		if knownHostsStr == "" {
			knownHostsStr = v
		} else {
			knownHostsStr = strings.Join([]string{knownHostsStr, v}, "\n")
		}
	}

	values := &Values{
		FullnameOverride: "argocd",
		InstallCRDs:      true,
		Global: &Global{
			Image: &Image{
				Repository:      "argoproj/argocd",
				Tag:             imageTags["argoproj/argocd"],
				ImagePullPolicy: "IfNotPresent",
			},
		},
		Controller: &Controller{
			Image: &Image{
				Repository:      "argoproj/argocd",
				Tag:             imageTags["argoproj/argocd"],
				ImagePullPolicy: "IfNotPresent",
			},
			Name: "application-controller",
			Args: &Args{
				StatusProcessors:    "20",
				OperationProcessors: "10",
			},
			LogLevel:      "info",
			ContainerPort: 8082,
			ReadinessProbe: &ReadinessProbe{
				FailureThreshold:    3,
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      1,
			},
			LivenessProbe: &LivenessProbe{
				FailureThreshold:    3,
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      1,
			},
			Service: &Service{
				Port: 8082,
			},
			ServiceAccount: &ServiceAccount{
				Create: true,
				Name:   "argocd-application-controller",
			},
			Metrics: &Metrics{
				Enabled: true,
				Service: &MetricsService{
					ServicePort: 8082,
				},
				ServiceMonitor: &ServiceMonitor{
					Enabled: false,
				},
				Rules: &Rules{
					Enabled: false,
				},
			},
			ClusterAdminAccess: &ClusterAdminAccess{
				Enabled: true,
			},
			NodeSelector: map[string]string{},
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
		Dex: &Dex{
			Enabled: false,
			ServiceAccount: &ServiceAccount{
				Create: false,
				Name:   "argocd-dex-server",
			},
			NodeSelector: map[string]string{},
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
		Redis: &Redis{
			Enabled: true,
			Name:    "redis",
			Image: &Image{
				Repository:      "redis",
				Tag:             imageTags["redis"],
				ImagePullPolicy: "IfNotPresent",
			},
			ContainerPort: 6379,
			ServicePort:   6379,
			NodeSelector:  map[string]string{},
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
		RedisHA: &RedisHA{
			Enabled:           false,
			Exporter:          &Exporter{Enabled: false},
			PersistenteVolume: &PV{Enabled: false},
			Redis: &RedisHAC{
				MasterGroupName: "argocd",
				Config:          &RedisHAConfig{Save: "\"\\\"\\\"\""},
			},
			HAProxy: &HAProxy{
				Enabled: false,
				Metrics: &HAProxyMetrics{Enabled: false},
			},
		},
		Server: &Server{
			Image: &Image{
				Repository:      "argoproj/argocd",
				Tag:             imageTags["argoproj/argocd"],
				ImagePullPolicy: "IfNotPresent",
			},
			Name:          "server",
			LogLevel:      "info",
			ContainerPort: 8080,
			ReadinessProbe: &ReadinessProbe{
				FailureThreshold:    3,
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      1,
			},
			LivenessProbe: &LivenessProbe{
				FailureThreshold:    3,
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      1,
			},
			Certificate: &Certificate{
				Enabled: false,
			},
			Service: &ServerService{
				Type:             "ClusterIP",
				ServicePortHTTP:  80,
				ServicePortHTTPS: 443,
			},
			Metrics: &Metrics{
				Enabled: true,
				Service: &MetricsService{
					ServicePort: 8083,
				},
				ServiceMonitor: &ServiceMonitor{
					Enabled: false,
				},
			},
			ServiceAccount: &ServiceAccount{
				Create: true,
				Name:   "argocd-server",
			},
			Ingress: &Ingress{
				Enabled: false,
			},
			Route: &Route{
				Enabled: false,
			},
			Config: &Config{
				URL:                         "https://argocd.example.com",
				ApplicationInstanceLabelKey: "argocd.argoproj.io/instance",
			},
			NodeSelector: map[string]string{},
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
		RepoServer: &RepoServer{
			Image: &Image{
				Repository:      "argoproj/argocd",
				Tag:             imageTags["argoproj/argocd"],
				ImagePullPolicy: "IfNotPresent",
			},
			Name:          "repo-server",
			LogLevel:      "info",
			ContainerPort: 8081,
			ReadinessProbe: &ReadinessProbe{
				FailureThreshold:    3,
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      1,
			},
			LivenessProbe: &LivenessProbe{
				FailureThreshold:    3,
				InitialDelaySeconds: 10,
				PeriodSeconds:       10,
				SuccessThreshold:    1,
				TimeoutSeconds:      1,
			},
			Service: &Service{
				Port: 8081,
			},
			Metrics: &Metrics{
				Enabled: true,
				Service: &MetricsService{
					ServicePort: 8084,
				},
				ServiceMonitor: &ServiceMonitor{
					Enabled: false,
				},
			},
			ServiceAccount: &ServiceAccount{
				Create: false,
			},
			NodeSelector: map[string]string{},
			Resources: &k8s.Resources{
				Limits: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("256Mi"),
				},
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("50m"),
					corev1.ResourceMemory: resource.MustParse("128Mi"),
				},
			},
		},
		Configs: &Configs{
			KnownHosts: &Data{
				Data: map[string]string{
					"ssh_known_hosts": knownHostsStr,
				},
			},
			Secret: &Secret{
				CreateSecret: true,
			},
		},
	}

	return values
}
