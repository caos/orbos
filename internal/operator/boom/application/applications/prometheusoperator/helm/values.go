package helm

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"
	corev1 "k8s.io/api/core/v1"
)

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
	Rules       *Rules            `yaml:"rules"`
	Labels      map[string]string `yaml:"labels"`
	Annotations map[string]string `yaml:"annotations"`
}

type Global struct {
	Rbac             *Rbac         `yaml:"rbac"`
	ImagePullSecrets []interface{} `yaml:"imagePullSecrets"`
}

type Rbac struct {
	Create     bool `yaml:"create"`
	PspEnabled bool `yaml:"pspEnabled"`
}

type DisabledTool struct {
	Enabled bool `yaml:"enabled"`
}

type TLSProxy struct {
	Enabled   bool        `yaml:"enabled"`
	Image     *Image      `yaml:"image"`
	Resources interface{} `yaml:"resources"`
}

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}

type Patch struct {
	Enabled           bool              `yaml:"enabled"`
	Image             *Image            `yaml:"image"`
	PriorityClassName string            `yaml:"priorityClassName"`
	PodAnnotations    map[string]string `yaml:"podAnnotations"`
	NodeSelector      map[string]string `yaml:"nodeSelector"`
}

type AdmissionWebhooks struct {
	FailurePolicy string `yaml:"failurePolicy"`
	Enabled       bool   `yaml:"enabled"`
	Patch         *Patch `yaml:"patch"`
}

type ServiceAccount struct {
	Create bool   `yaml:"create"`
	Name   string `yaml:"name"`
}

type Service struct {
	Annotations              map[string]string `yaml:"annotations"`
	Labels                   map[string]string `yaml:"labels"`
	ClusterIP                string            `yaml:"clusterIP"`
	NodePort                 int               `yaml:"nodePort"`
	NodePortTLS              int               `yaml:"nodePortTls"`
	AdditionalPorts          []interface{}     `yaml:"additionalPorts"`
	LoadBalancerIP           string            `yaml:"loadBalancerIP"`
	LoadBalancerSourceRanges []interface{}     `yaml:"loadBalancerSourceRanges"`
	Type                     string            `yaml:"type"`
	ExternalIPs              []interface{}     `yaml:"externalIPs"`
}

type KubeletService struct {
	Enabled   bool   `yaml:"enabled"`
	Namespace string `yaml:"namespace"`
}

type ServiceMonitor struct {
	Interval          string        `yaml:"interval"`
	SelfMonitor       bool          `yaml:"selfMonitor"`
	MetricRelabelings []interface{} `yaml:"metricRelabelings"`
	Relabelings       []interface{} `yaml:"relabelings"`
}

type SecurityContext struct {
	RunAsNonRoot bool `yaml:"runAsNonRoot"`
	RunAsUser    int  `yaml:"runAsUser"`
}

type PrometheusOperatorValues struct {
	Enabled                       bool                `yaml:"enabled"`
	TLSProxy                      *TLSProxy           `yaml:"tlsProxy"`
	AdmissionWebhooks             *AdmissionWebhooks  `yaml:"admissionWebhooks"`
	DenyNamespaces                []string            `yaml:"denyNamespaces"`
	ServiceAccount                *ServiceAccount     `yaml:"serviceAccount"`
	Service                       *Service            `yaml:"service"`
	CreateCustomResource          bool                `yaml:"createCustomResource"`
	CrdAPIGroup                   string              `yaml:"crdApiGroup"`
	CleanupCustomResource         bool                `yaml:"cleanupCustomResource"`
	PodLabels                     map[string]string   `yaml:"podLabels"`
	PodAnnotations                map[string]string   `yaml:"podAnnotations"`
	KubeletService                *KubeletService     `yaml:"kubeletService"`
	ServiceMonitor                *ServiceMonitor     `yaml:"serviceMonitor"`
	NodeSelector                  map[string]string   `yaml:"nodeSelector"`
	Tolerations                   []corev1.Toleration `yaml:"tolerations"`
	Affinity                      struct{}            `yaml:"affinity"`
	SecurityContext               *SecurityContext    `yaml:"securityContext"`
	Image                         *Image              `yaml:"image"`
	ConfigmapReloadImage          *Image              `yaml:"configmapReloadImage"`
	PrometheusConfigReloaderImage *Image              `yaml:"prometheusConfigReloaderImage"`
	ConfigReloaderCPU             string              `yaml:"configReloaderCpu"`
	ConfigReloaderMemory          string              `yaml:"configReloaderMemory"`
	HyperkubeImage                *Image              `yaml:"hyperkubeImage"`
	Resources                     *k8s.Resources      `yaml:"resources"`
}

type DisabledToolServicePerReplica struct {
	Enabled           bool
	ServicePerReplica *DisabledTool `yaml:"servicePerReplica"`
	IngressPerReplica *DisabledTool `yaml:"ingressPerReplica"`
}

type Values struct {
	NameOverride              string                         `yaml:"nameOverride,omitempty"`
	FullnameOverride          string                         `yaml:"fullnameOverride,omitempty"`
	CommonLabels              map[string]string              `yaml:"commonLabels"`
	DefaultRules              *DefaultRules                  `yaml:"defaultRules"`
	AdditionalPrometheusRules []interface{}                  `yaml:"additionalPrometheusRules"`
	Global                    *Global                        `yaml:"global"`
	Alertmanager              *DisabledToolServicePerReplica `yaml:"alertmanager"`
	Grafana                   *DisabledTool                  `yaml:"grafana"`
	KubeAPIServer             *DisabledTool                  `yaml:"kubeApiServer"`
	Kubelet                   *DisabledTool                  `yaml:"kubelet"`
	KubeControllerManager     *DisabledTool                  `yaml:"kubeControllerManager"`
	CoreDNS                   *DisabledTool                  `yaml:"coreDns"`
	KubeDNS                   *DisabledTool                  `yaml:"kubeDns"`
	KubeEtcd                  *DisabledTool                  `yaml:"kubeEtcd"`
	KubeScheduler             *DisabledTool                  `yaml:"kubeScheduler"`
	KubeProxy                 *DisabledTool                  `yaml:"kubeProxy"`
	KubeStateMetricsScrap     *DisabledTool                  `yaml:"kubeStateMetrics"`
	KubeStateMetrics          *DisabledTool                  `yaml:"kube-state-metrics"`
	NodeExporter              *DisabledTool                  `yaml:"nodeExporter"`
	PrometheusNodeExporter    *DisabledTool                  `yaml:"prometheus-node-exporter"`
	PrometheusOperator        *PrometheusOperatorValues      `yaml:"prometheusOperator"`
	Prometheus                *DisabledToolServicePerReplica `yaml:"prometheus"`
}
