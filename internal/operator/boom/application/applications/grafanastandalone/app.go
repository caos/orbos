package grafanastandalone

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana"
	"path/filepath"

	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/utils/kustomize"
)

var (
	applicationName = "grafanastandalone"
)

type Grafana struct {
	ApplicationDirectoryPath string
	monitor                  mntr.Monitor
	spec                     *grafana.Grafana
}

func New(monitor mntr.Monitor, toolsDirectoryPath string) *Grafana {
	lo := &Grafana{
		ApplicationDirectoryPath: filepath.Join(toolsDirectoryPath, applicationName),
		monitor:                  monitor,
	}

	return lo
}

func specToValues(imageTags map[string]string, spec *grafana.Grafana) *Values {
	values := &Values{
		FullnameOverride: "grafana",
		Rbac: &Rbac{
			Create:         true,
			PspEnabled:     false,
			PspUseAppArmor: true,
			Namespaced:     false,
		},
		ServiceAccount: &ServiceAccount{
			Create: true,
		},
		Replicas: 1,
		DeploymentStrategy: &DeploymentStrategy{
			Type: "RollingUpdate",
		},
		ReadinessProbe: &ReadinessProbe{
			HTTPGet: &HTTPGet{
				Port: 3000,
				Path: "/api/health",
			},
		},
		LivenessProbe: &LivenessProbe{
			HTTPGet: &HTTPGet{
				Port: 3000,
				Path: "/api/health",
			},
			InitialDelaySeconds: 60,
			TimeoutSeconds:      30,
			FailureThreshold:    10,
		},
		Image: &Image{
			Repository: "grafana/grafana",
			Tag:        imageTags["grafana/grafana"],
			PullPolicy: "IfNotPresent",
		},
		TestFramework: &TestFramework{
			Enabled: true,
			Image:   "dduportal/bats",
			Tag:     imageTags["dduportal/bats"],
		},
		SecurityContext: &SecurityContext{
			RunAsUser: 472,
			FsGroup:   472,
		},
		DownloadDashboardsImage: &DownloadDashboardsImage{
			Repository: "appropriate/curl",
			Tag:        imageTags["appropriate/curl"],
			PullPolicy: "IfNotPresent",
		},
		DownloadDashboards: &DownloadDashboards{},
		PodPortName:        "grafana",
		Service: &Service{
			Type:       "ClusterIP",
			Port:       80,
			TargetPort: 3000,
			PortName:   "service",
		},
		Ingress: &Ingress{
			Enabled: false,
		},
		Persistence: &Persistence{
			Type:        "pvc",
			Enabled:     false,
			AccessModes: []string{"ReadWriteOnce"},
			Size:        "10Gi",
			Finalizers:  []string{"kubernetes.io/pvc-protection"},
		},
		InitChownData: &InitChownData{
			Enabled: true,
			Image: &Image{
				Repository: "busybox",
				Tag:        imageTags["busybox"],
				PullPolicy: "IfNotPresent",
			},
		},
		AdminUser:     "admin",
		AdminPassword: "admin",
		Admin: &Admin{
			ExistingSecret: "",
			UserKey:        "admin-user",
			PasswordKey:    "admin-password",
		},
		// Datasources             *Datasources             `yaml:"datasources"`
		// Dashboards              *Dashboards              `yaml:"dashboards"`
		// DashboardsConfigMaps    map[string]string        `yaml:"dashboardsConfigMaps"`
		GrafanaIni: &GrafanaIni{
			Paths: &Paths{
				Data:         "/var/lib/grafana/data",
				Logs:         "/var/log/grafana",
				Plugins:      "/var/lib/grafana/plugins",
				Provisioning: "/etc/grafana/provisioning",
			},
			Analytics: &Analytics{
				CheckForUpdates: true,
			},
			Log: &Log{
				Mode: "console",
			},
			GrafanaNet: &GrafanaNet{
				URL: "https://grafana.net",
			},
		},
		Ldap: &Ldap{
			Enabled: false,
		},
		SMTP: &SMTP{
			ExistingSecret: "",
			UserKey:        "user",
			PasswordKey:    "password",
		},
		Sidecar: &Sidecar{
			Image:           "kiwigrid/k8s-sidecar:0.1.20",
			ImagePullPolicy: "IfNotPresent",
			Dashboards: &DashboardsSidecar{
				Enabled: false,
			},
			Datasources: &DatasourcesSidecar{
				Enabled: false,
			},
		},
	}

	if spec.Datasources != nil {
		datasources := make([]*Datasource, 0)
		for _, datasource := range spec.Datasources {
			valuesDatasource := &Datasource{
				Name:      datasource.Name,
				Type:      datasource.Type,
				URL:       datasource.Url,
				Access:    datasource.Access,
				IsDefault: datasource.IsDefault,
			}
			datasources = append(datasources, valuesDatasource)
		}
		values.Datasources = &Datasources{
			Datasources: &Datasourcesyaml{
				APIVersion:  1,
				Datasources: datasources,
			},
		}
	}

	if spec.DashboardProviders != nil {
		providers := make([]*Provider, 0)
		dashboards := make(map[string]string, 0)
		for _, provider := range spec.DashboardProviders {
			for _, configmap := range provider.ConfigMaps {
				providers = append(providers, getProvider(configmap))
				dashboards[configmap] = configmap
			}
		}
		values.DashboardProviders = &DashboardProviders{
			Providers: &Providersyaml{
				APIVersion: 1,
				Providers:  providers,
			},
		}
		values.DashboardsConfigMaps = dashboards
	}

	return values
}

func getKustomizeOutput(folders []string) ([]string, error) {
	ret := make([]string, len(folders))
	for n, folder := range folders {

		cmd, err := kustomize.New(folder)
		if err != nil {
			return nil, err
		}
		execcmd := cmd.Build()

		out, err := execcmd.Output()
		if err != nil {
			return nil, err
		}
		ret[n] = string(out)
	}
	return ret, nil
}

func getProvider(appName string) *Provider {
	return &Provider{
		Name:            appName,
		Type:            "file",
		DisableDeletion: false,
		Editable:        true,
		Options: map[string]string{
			"path": filepath.Join("/var/lib/grafana/dashboards", appName),
		},
	}
}
