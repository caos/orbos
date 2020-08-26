package helm

import (
	prometheusoperatorhelm "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/helm"
)

type Ingress struct {
	Enabled     bool              `yaml:"enabled"`
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
	Hosts       []interface{}     `yaml:"hosts"`
	Path        string            `yaml:"path"`
	TLS         []interface{}     `yaml:"tls"`
}

type Dashboards struct {
	Enabled bool   `yaml:"enabled"`
	Label   string `yaml:"label"`
}

type Datasources struct {
	Enabled                             bool              `yaml:"enabled"`
	DefaultDatasourceEnabled            bool              `yaml:"defaultDatasourceEnabled"`
	Annotations                         map[string]string `yaml:"annotations"`
	CreatePrometheusReplicasDatasources bool              `yaml:"createPrometheusReplicasDatasources"`
	Label                               string            `yaml:"label"`
}

type ServiceMonitor struct {
	Interval          string        `yaml:"interval"`
	SelfMonitor       bool          `yaml:"selfMonitor"`
	MetricRelabelings []interface{} `yaml:"metricRelabelings"`
	Relabelings       []interface{} `yaml:"relabelings"`
}

type Sidecar struct {
	Dashboards  *Dashboards  `yaml:"dashboards"`
	Datasources *Datasources `yaml:"datasources"`
}
type DashboardProviders struct {
	Providers *Providersyaml `yaml:"dashboardproviders.yaml"`
}
type Providersyaml struct {
	APIVersion int64       `yaml:"apiVersion"`
	Providers  []*Provider `yaml:"providers"`
}
type Provider struct {
	Name            string            `yaml:"name"`
	OrgID           int               `yaml:"ordId"`
	Folder          string            `yaml:"folder,omitempty"`
	Type            string            `yaml:"type"`
	DisableDeletion bool              `yaml:"disableDeletion"`
	Editable        bool              `yaml:"editable"`
	Options         map[string]string `yaml:"options"`
}

type Admin struct {
	ExistingSecret string `yaml:"existingSecret"`
	UserKey        string `yaml:"userKey"`
	PasswordKey    string `yaml:"passwordKey"`
}

type Service struct {
	Labels map[string]string `yaml:"labels,omitempty"`
}

type Persistence struct {
	Type             string   `yaml:"type"`
	Enabled          bool     `yaml:"enabled"`
	AccessModes      []string `yaml:"accessModes"`
	Size             string   `yaml:"size"`
	StorageClassName string   `yaml:"storageClassName"`
	Finalizers       []string `yaml:"finalizers"`
}

type Ini struct {
	Paths       map[string]string      `yaml:"paths,omitempty"`
	Analytics   map[string]bool        `yaml:"analytics,omitempty"`
	Log         map[string]string      `yaml:"log,omitempty"`
	GrafanaNet  map[string]interface{} `yaml:"grafana_net,omitempty"`
	AuthGoogle  map[string]string      `yaml:"auth.google,omitempty"`
	AuthGitlab  map[string]string      `yaml:"auth.gitlab,omitempty"`
	AuthGithub  map[string]string      `yaml:"auth.github,omitempty"`
	AuthGeneric map[string]string      `yaml:"auth.generic_oauth,omitempty"`
}

type Datasource struct {
	Name      string `yaml:"name"`
	Type      string `yaml:"type"`
	URL       string `yaml:"url"`
	Access    string `yaml:"access"`
	IsDefault bool   `yaml:"isDefault"`
}

type TestFramework struct {
	Enabled         bool   `yaml:"enabled"`
	Image           string `yaml:"image"`
	Tag             string `yaml:"tag"`
	SecurityContext struct {
	} `yaml:"securityContext"`
}
type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}
type GrafanaValues struct {
	FullnameOverride         string              `yaml:"fullnameOverride,omitempty"`
	Enabled                  bool                `yaml:"enabled"`
	DefaultDashboardsEnabled bool                `yaml:"defaultDashboardsEnabled"`
	AdminPassword            string              `yaml:"adminPassword"`
	Admin                    *Admin              `yaml:"admin"`
	Ingress                  *Ingress            `yaml:"ingress"`
	Sidecar                  *Sidecar            `yaml:"sidecar"`
	ExtraConfigmapMounts     []interface{}       `yaml:"extraConfigmapMounts"`
	AdditionalDataSources    []*Datasource       `yaml:"additionalDataSources"`
	ServiceMonitor           *ServiceMonitor     `yaml:"serviceMonitor"`
	DashboardProviders       *DashboardProviders `yaml:"dashboardProviders,omitempty"`
	DashboardsConfigMaps     map[string]string   `yaml:"dashboardsConfigMaps,omitempty"`
	Ini                      *Ini                `yaml:"grafana.ini,omitempty"`
	Persistence              *Persistence        `yaml:"persistence,omitempty"`
	TestFramework            *TestFramework      `yaml:"testFramework,omitempty"`
	Plugins                  []string            `yaml:"plugins,omitempty"`
	Image                    *Image              `yaml:"image,omitempty"`
	Env                      map[string]string   `yaml:"env,omitempty"`
	Service                  *Service            `yaml:"service,omitempty"`
	Labels                   map[string]string   `yaml:"labels,omitempty"`
	PodLabels                map[string]string   `yaml:"podLabels,omitempty"`
	NodeSelector             map[string]string   `yaml:"nodeSelector,omitempty"`
}

type Rules struct {
	Alertmanager                bool `yaml:"alertmanager"`
	Etcd                        bool `yaml:"etcd"`
	General                     bool `yaml:"general"`
	K8S                         bool `yaml:"k8s"`
	KubeApiserver               bool `yaml:"kubeApiserver"`
	KubePrometheusNodeAlerting  bool `yaml:"kubePrometheusNodeAlerting"`
	KubePrometheusNodeRecording bool `yaml:"kubePrometheusNodeRecording"`
	KubernetesAbsent            bool `yaml:"kubernetesAbsent"`
	KubernetesApps              bool `yaml:"kubernetesApps"`
	KubernetesResources         bool `yaml:"kubernetesResources"`
	KubernetesStorage           bool `yaml:"kubernetesStorage"`
	KubernetesSystem            bool `yaml:"kubernetesSystem"`
	KubeScheduler               bool `yaml:"kubeScheduler"`
	Network                     bool `yaml:"network"`
	Node                        bool `yaml:"node"`
	Prometheus                  bool `yaml:"prometheus"`
	PrometheusOperator          bool `yaml:"prometheusOperator"`
	Time                        bool `yaml:"time"`
}

type DefaultRules struct {
	Create      bool              `yaml:"create"`
	Rules       *Rules            `yaml:"rules,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type Values struct {
	DefaultRules              *DefaultRules                                    `yaml:"defaultRules,omitempty"`
	Global                    *Global                                          `yaml:"global,omitempty"`
	KubeTargetVersionOverride string                                           `yaml:"kubeTargetVersionOverride,omitempty"`
	NameOverride              string                                           `yaml:"nameOverride,omitempty"`
	FullnameOverride          string                                           `yaml:"fullnameOverride,omitempty"`
	CommonLabels              map[string]string                                `yaml:"commonLabels,omitempty"`
	Alertmanager              *DisabledTool                                    `yaml:"alertmanager,omitempty"`
	Grafana                   *GrafanaValues                                   `yaml:"grafana,omitempty"`
	KubeAPIServer             *DisabledTool                                    `yaml:"kubeApiServer,omitempty"`
	Kubelet                   *DisabledTool                                    `yaml:"kubelet,omitempty"`
	KubeControllerManager     *DisabledTool                                    `yaml:"kubeControllerManager,omitempty"`
	CoreDNS                   *DisabledTool                                    `yaml:"coreDns,omitempty"`
	KubeDNS                   *DisabledTool                                    `yaml:"kubeDns,omitempty"`
	KubeEtcd                  *DisabledTool                                    `yaml:"kubeEtcd,omitempty"`
	KubeScheduler             *DisabledTool                                    `yaml:"kubeScheduler,omitempty"`
	KubeProxy                 *DisabledTool                                    `yaml:"kubeProxy,omitempty"`
	KubeStateMetricsScrap     *DisabledTool                                    `yaml:"kubeStateMetrics,omitempty"`
	KubeStateMetrics          *DisabledTool                                    `yaml:"kube-state-metrics,omitempty"`
	NodeExporter              *DisabledTool                                    `yaml:"nodeExporter,omitempty"`
	PrometheusNodeExporter    *DisabledTool                                    `yaml:"prometheus-node-exporter,omitempty"`
	PrometheusOperator        *prometheusoperatorhelm.PrometheusOperatorValues `yaml:"prometheusOperator,omitempty"`
	Prometheus                *DisabledTool                                    `yaml:"prometheus,omitempty"`
}

type Global struct {
	Rbac             *Rbac         `yaml:"rbac,omitempty"`
	ImagePullSecrets []interface{} `yaml:"imagePullSecrets,omitempty"`
}

type Rbac struct {
	Create     bool `yaml:"create,omitempty"`
	PspEnabled bool `yaml:"pspEnabled,omitempty"`
}

type DisabledTool struct {
	Enabled bool `yaml:"enabled"`
}
