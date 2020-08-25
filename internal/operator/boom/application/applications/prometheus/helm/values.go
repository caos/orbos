package helm

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	prometheusoperator "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/helm"
)

type Service struct {
	Annotations              map[string]string `yaml:"annotations,omitempty"`
	Labels                   map[string]string `yaml:"labels,omitempty"`
	ClusterIP                string            `yaml:"clusterIP,omitempty"`
	Port                     int               `yaml:"port,omitempty"`
	TargetPort               int               `yaml:"targetPort,omitempty"`
	ExternalIPs              []interface{}     `yaml:"externalIPs,omitempty"`
	NodePort                 int               `yaml:"nodePort,omitempty"`
	LoadBalancerIP           string            `yaml:"loadBalancerIP,omitempty"`
	LoadBalancerSourceRanges []interface{}     `yaml:"loadBalancerSourceRanges,omitempty"`
	Type                     string            `yaml:"type,omitempty"`
	SessionAffinity          string            `yaml:"sessionAffinity,omitempty"`
}

type ServicePerReplica struct {
	Enabled                  bool              `yaml:"enabled,omitempty"`
	Annotations              map[string]string `yaml:"annotations,omitempty"`
	Port                     int               `yaml:"port,omitempty"`
	TargetPort               int               `yaml:"targetPort,omitempty"`
	NodePort                 int               `yaml:"nodePort,omitempty"`
	LoadBalancerSourceRanges []interface{}     `yaml:"loadBalancerSourceRanges,omitempty"`
	Type                     string            `yaml:"type,omitempty"`
}
type PodDisruptionBudget struct {
	Enabled        bool   `yaml:"enabled,omitempty"`
	MinAvailable   int    `yaml:"minAvailable,omitempty"`
	MaxUnavailable string `yaml:"maxUnavailable,omitempty"`
}

type Ingress struct {
	Enabled     bool              `yaml:"enabled,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Hosts       []interface{}     `yaml:"hosts,omitempty"`
	Paths       []interface{}     `yaml:"paths,omitempty"`
	TLS         []interface{}     `yaml:"tls,omitempty"`
}
type IngressPerReplica struct {
	Enabled       bool              `yaml:"enabled,omitempty"`
	Annotations   map[string]string `yaml:"annotations,omitempty"`
	Labels        map[string]string `yaml:"labels,omitempty"`
	HostPrefix    string            `yaml:"hostPrefix,omitempty"`
	HostDomain    string            `yaml:"hostDomain,omitempty"`
	Paths         []interface{}     `yaml:"paths,omitempty"`
	TLSSecretName string            `yaml:"tlsSecretName,omitempty"`
}

type PodSecurityPolicy struct {
	AllowedCapabilities []interface{} `yaml:"allowedCapabilities,omitempty"`
}

type ServiceMonitor struct {
	Interval          string        `yaml:"interval,omitempty"`
	SelfMonitor       bool          `yaml:"selfMonitor,omitempty"`
	BearerTokenFile   interface{}   `yaml:"bearerTokenFile,omitempty"`
	MetricRelabelings []interface{} `yaml:"metricRelabelings,omitempty"`
	Relabelings       []interface{} `yaml:"relabelings,omitempty"`
}

type Image struct {
	Repository string `yaml:"repository,omitempty"`
	Tag        string `yaml:"tag,omitempty"`
}

type NamespaceSelector struct {
	Any        bool     `yaml:"any,omitempty"`
	MatchNames []string `yaml:"matchNames,omitempty"`
}

type MonitorSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
}

type Query struct {
	LookbackDelta  string `yaml:"lookbackDelta,omitempty"`
	MaxConcurrency int32  `yaml:"maxConcurrency,omitempty"`
	MaxSamples     int32  `yaml:"maxSamples,omitempty"`
	Timeout        string `yaml:"timeout,omitempty"`
}
type PodMetadata struct {
	Labels map[string]string `yaml:"labels,omitempty"`
}
type StorageSpec struct {
	VolumeClaimTemplate *VolumeClaimTemplate `yaml:"volumeClaimTemplate,omitempty"`
}
type VolumeClaimTemplate struct {
	Spec     *VolumeClaimTemplateSpec `yaml:"spec,omitempty"`
	selector struct{}                 `yaml:"selector,omitempty"`
}
type VolumeClaimTemplateSpec struct {
	StorageClassName string     `yaml:"storageClassName,omitempty"`
	AccessModes      []string   `yaml:"accessModes,omitempty"`
	Resources        *Resources `yaml:"resources,omitempty"`
}
type Resources struct {
	Requests *Request `yaml:"requests,omitempty"`
}

type Request struct {
	Storage string `yaml:"storage,omitempty"`
}
type SecurityContext struct {
	RunAsNonRoot bool `yaml:"runAsNonRoot,omitempty"`
	RunAsUser    int  `yaml:"runAsUser,omitempty"`
	FsGroup      int  `yaml:"fsGroup,omitempty"`
}

type KubernetesSdConfig struct {
	Role string `yaml:"role,omitempty"`
}
type TLSConfig struct {
	CaFile   string `yaml:"ca_file,omitempty"`
	CertFile string `yaml:"cert_file,omitempty"`
	KeyFile  string `yaml:"key_file,omitempty"`
}
type RelabelConfig struct {
	Action       string   `yaml:"action,omitempty"`
	Regex        string   `yaml:"regex,omitempty"`
	SourceLabels []string `yaml:"source_labels,omitempty"`
	TargetLabel  string   `yaml:"target_label,omitempty"`
	Replacement  string   `yaml:"replacement,omitempty"`
	Modulus      uint64   `yaml:"modulus,omitempty"`
	Separator    string   `yaml:"separator,omitempty"`
}

type ValuesRelabelConfig struct {
	Action       string   `yaml:"action,omitempty"`
	Regex        string   `yaml:"regex,omitempty"`
	SourceLabels []string `yaml:"sourceLabels,omitempty"`
	TargetLabel  string   `yaml:"targetLabel,omitempty"`
	Replacement  string   `yaml:"replacement,omitempty"`
	Modulus      uint64   `yaml:"modulus,omitempty"`
	Separator    string   `yaml:"separator,omitempty"`
}

type AdditionalScrapeConfig struct {
	JobName              string                `yaml:"job_name,omitempty"`
	KubernetesSdConfigs  []*KubernetesSdConfig `yaml:"kubernetes_sd_configs,omitempty"`
	Scheme               string                `yaml:"scheme,omitempty"`
	TLSConfig            *TLSConfig            `yaml:"tls_config,omitempty"`
	RelabelConfigs       []*RelabelConfig      `yaml:"relabel_configs,omitempty"`
	MetricRelabelConfigs []*RelabelConfig      `yaml:"metric_relabel_configs,omitempty"`
	BearerTokenFile      string                `yaml:"bearer_token_file,omitempty"`
}

type PrometheusSpec struct {
	ScrapeInterval                          string                    `yaml:"scrapeInterval,omitempty"`
	EvaluationInterval                      string                    `yaml:"evaluationInterval,omitempty"`
	ListenLocal                             bool                      `yaml:"listenLocal,omitempty"`
	EnableAdminAPI                          bool                      `yaml:"enableAdminAPI,omitempty"`
	Image                                   *Image                    `yaml:"image,omitempty"`
	Tolerations                             []interface{}             `yaml:"tolerations,omitempty"`
	AlertingEndpoints                       []interface{}             `yaml:"alertingEndpoints,omitempty"`
	ExternalLabels                          map[string]string         `yaml:"externalLabels,omitempty"`
	ReplicaExternalLabelName                string                    `yaml:"replicaExternalLabelName,omitempty"`
	ReplicaExternalLabelNameClear           bool                      `yaml:"replicaExternalLabelNameClear,omitempty"`
	PrometheusExternalLabelName             string                    `yaml:"prometheusExternalLabelName,omitempty"`
	PrometheusExternalLabelNameClear        bool                      `yaml:"prometheusExternalLabelNameClear,omitempty"`
	ExternalURL                             string                    `yaml:"externalUrl,omitempty"`
	NodeSelector                            map[string]string         `yaml:"nodeSelector,omitempty"`
	Secrets                                 []interface{}             `yaml:"secrets,omitempty"`
	ConfigMaps                              []interface{}             `yaml:"configMaps,omitempty"`
	Query                                   *Query                    `yaml:"query,omitempty"`
	RuleNamespaceSelector                   *NamespaceSelector        `yaml:"ruleNamespaceSelector,omitempty"`
	RuleSelectorNilUsesHelmValues           bool                      `yaml:"ruleSelectorNilUsesHelmValues,omitempty"`
	RuleSelector                            *RuleSelector             `yaml:"ruleSelector,omitempty"`
	ServiceMonitorSelectorNilUsesHelmValues bool                      `yaml:"serviceMonitorSelectorNilUsesHelmValues,omitempty"`
	ServiceMonitorSelector                  *MonitorSelector          `yaml:"serviceMonitorSelector,omitempty"`
	ServiceMonitorNamespaceSelector         *NamespaceSelector        `yaml:"serviceMonitorNamespaceSelector,omitempty"`
	PodMonitorSelectorNilUsesHelmValues     bool                      `yaml:"podMonitorSelectorNilUsesHelmValues,omitempty"`
	PodMonitorSelector                      *MonitorSelector          `yaml:"podMonitorSelector,omitempty"`
	PodMonitorNamespaceSelector             *NamespaceSelector        `yaml:"podMonitorNamespaceSelector,omitempty"`
	Retention                               string                    `yaml:"retention,omitempty"`
	RetentionSize                           string                    `yaml:"retentionSize,omitempty"`
	WalCompression                          bool                      `yaml:"walCompression,omitempty"`
	Paused                                  bool                      `yaml:"paused,omitempty"`
	Replicas                                int                       `yaml:"replicas,omitempty"`
	LogLevel                                string                    `yaml:"logLevel,omitempty"`
	LogFormat                               string                    `yaml:"logFormat,omitempty"`
	RoutePrefix                             string                    `yaml:"routePrefix,omitempty"`
	PodMetadata                             *PodMetadata              `yaml:"podMetadata,omitempty"`
	PodAntiAffinity                         string                    `yaml:"podAntiAffinity,omitempty"`
	PodAntiAffinityTopologyKey              string                    `yaml:"podAntiAffinityTopologyKey,omitempty"`
	Affinity                                struct{}                  `yaml:"affinity,omitempty"`
	RemoteRead                              []interface{}             `yaml:"remoteRead,omitempty"`
	RemoteWrite                             []*RemoteWrite            `yaml:"remoteWrite,omitempty"`
	RemoteWriteDashboards                   bool                      `yaml:"remoteWriteDashboards,omitempty"`
	Resources                               struct{}                  `yaml:"resources,omitempty"`
	StorageSpec                             *StorageSpec              `yaml:"storageSpec,omitempty"`
	AdditionalScrapeConfigs                 []*AdditionalScrapeConfig `yaml:"additionalScrapeConfigs,omitempty"`
	AdditionalAlertManagerConfigs           []interface{}             `yaml:"additionalAlertManagerConfigs,omitempty"`
	AdditionalAlertRelabelConfigs           []interface{}             `yaml:"additionalAlertRelabelConfigs,omitempty"`
	SecurityContext                         *SecurityContext          `yaml:"securityContext,omitempty"`
	PriorityClassName                       string                    `yaml:"priorityClassName,omitempty"`
	Thanos                                  struct{}                  `yaml:"thanos,omitempty"`
	Containers                              []interface{}             `yaml:"containers,omitempty"`
	AdditionalScrapeConfigsExternal         bool                      `yaml:"additionalScrapeConfigsExternal,omitempty"`
}

type RemoteWrite struct {
	URL                 string                 `yaml:"url,omitempty"`
	BasicAuth           *BasicAuth             `yaml:"basicAuth,omitempty"`
	WriteRelabelConfigs []*ValuesRelabelConfig `yaml:"writeRelabelConfigs,omitempty"`
}

type BasicAuth struct {
	Username *SecretKeySelector `yaml:"username,omitempty"`
	Password *SecretKeySelector `yaml:"password,omitempty"`
}

type SecretKeySelector struct {
	Name string `yaml:"name,omitempty"`
	Key  string `yaml:"key,omitempty"`
}

type ServiceAccount struct {
	Create bool   `yaml:"create,omitempty"`
	Name   string `yaml:"name,omitempty"`
}
type RuleSelector struct {
	MatchLabels map[string]string `yaml:"matchLabels"`
}
type PrometheusValues struct {
	Enabled                   bool                     `yaml:"enabled,omitempty"`
	Annotations               map[string]string        `yaml:"annotations,omitempty"`
	ServiceAccount            *ServiceAccount          `yaml:"serviceAccount,omitempty"`
	Service                   *Service                 `yaml:"service,omitempty"`
	ServicePerReplica         *ServicePerReplica       `yaml:"servicePerReplica,omitempty"`
	PodDisruptionBudget       *PodDisruptionBudget     `yaml:"podDisruptionBudget,omitempty"`
	Ingress                   *Ingress                 `yaml:"ingress,omitempty"`
	IngressPerReplica         *IngressPerReplica       `yaml:"ingressPerReplica,omitempty"`
	PodSecurityPolicy         *PodSecurityPolicy       `yaml:"podSecurityPolicy,omitempty"`
	ServiceMonitor            *ServiceMonitor          `yaml:"serviceMonitor,omitempty"`
	PrometheusSpec            *PrometheusSpec          `yaml:"prometheusSpec,omitempty"`
	AdditionalServiceMonitors []*servicemonitor.Values `yaml:"additionalServiceMonitors,omitempty"`
	AdditionalPodMonitors     []interface{}            `yaml:"additionalPodMonitors,omitempty"`
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

type AdditionalPrometheusRules struct {
	Name             string            `yaml:"name,omitempty"`
	Groups           []*Group          `yaml:"groups,omitempty"`
	AdditionalLabels map[string]string `yaml:"additionalLabels,omitempty"`
}

type Group struct {
	Name  string  `yaml:"name,omitempty"`
	Rules []*Rule `yaml:"rules,omitempty"`
}

type Rule struct {
	Expr        string            `yaml:"expr,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Record      string            `yaml:"record,omitempty"`
	Alert       string            `yaml:"alert,omitempty"`
	For         string            `yaml:"for,omitempty"`
}

type Values struct {
	KubeTargetVersionOverride string                                       `yaml:"kubeTargetVersionOverride,omitempty"`
	NameOverride              string                                       `yaml:"nameOverride,omitempty"`
	FullnameOverride          string                                       `yaml:"fullnameOverride,omitempty"`
	CommonLabels              map[string]string                            `yaml:"commonLabels,omitempty"`
	DefaultRules              *DefaultRules                                `yaml:"defaultRules,omitempty"`
	AdditionalPrometheusRules []*AdditionalPrometheusRules                 `yaml:"additionalPrometheusRules,omitempty"`
	Global                    *Global                                      `yaml:"global,omitempty"`
	Alertmanager              *DisabledTool                                `yaml:"alertmanager,omitempty"`
	Grafana                   *DisabledTool                                `yaml:"grafana,omitempty"`
	KubeAPIServer             *DisabledTool                                `yaml:"kubeApiServer,omitempty"`
	Kubelet                   *DisabledTool                                `yaml:"kubelet,omitempty"`
	KubeControllerManager     *DisabledTool                                `yaml:"kubeControllerManager,omitempty"`
	CoreDNS                   *DisabledTool                                `yaml:"coreDns,omitempty"`
	KubeDNS                   *DisabledTool                                `yaml:"kubeDns,omitempty"`
	KubeEtcd                  *DisabledTool                                `yaml:"kubeEtcd,omitempty"`
	KubeScheduler             *DisabledTool                                `yaml:"kubeScheduler,omitempty"`
	KubeProxy                 *DisabledTool                                `yaml:"kubeProxy,omitempty"`
	KubeStateMetricsScrap     *DisabledTool                                `yaml:"kubeStateMetrics,omitempty"`
	KubeStateMetrics          *DisabledTool                                `yaml:"kube-state-metrics,omitempty"`
	NodeExporter              *DisabledTool                                `yaml:"nodeExporter,omitempty"`
	PrometheusNodeExporter    *DisabledTool                                `yaml:"prometheus-node-exporter,omitempty"`
	PrometheusOperator        *prometheusoperator.PrometheusOperatorValues `yaml:"prometheusOperator,omitempty"`
	Prometheus                *PrometheusValues                            `yaml:"prometheus,omitempty"`
}
