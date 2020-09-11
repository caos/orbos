package helm

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/k8s"
	corev1 "k8s.io/api/core/v1"
)

type AdminService struct {
	Annotations map[string]string `yaml:"annotations"`
	Create      bool              `yaml:"create"`
	Port        int               `yaml:"port"`
	Type        string            `yaml:"type"`
}
type AuthService struct {
	Create                 bool        `yaml:"create"`
	OptionalConfigurations interface{} `yaml:"optional_configurations"`
}
type Resource struct {
	Name                     string `yaml:"name"`
	TargetAverageUtilization int    `yaml:"targetAverageUtilization"`
}
type Metrics []struct {
	Resource *Resource `yaml:"resource"`
	Type     string    `yaml:"type"`
}
type Autoscaling struct {
	Enabled     bool     `yaml:"enabled"`
	MaxReplicas int      `yaml:"maxReplicas"`
	Metrics     *Metrics `yaml:"metrics"`
	MinReplicas int      `yaml:"minReplicas"`
}
type Crds struct {
	Create  bool `yaml:"create"`
	Enabled bool `yaml:"enabled"`
	Keep    bool `yaml:"keep"`
}
type DeploymentStrategy struct {
	Type string `yaml:"type"`
}
type Image struct {
	PullPolicy string `yaml:"pullPolicy"`
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}
type LicenseKey struct {
	CreateSecret bool        `yaml:"createSecret"`
	Value        interface{} `yaml:"value"`
}
type LivenessProbe struct {
	FailureThreshold    int `yaml:"failureThreshold"`
	InitialDelaySeconds int `yaml:"initialDelaySeconds"`
	PeriodSeconds       int `yaml:"periodSeconds"`
}
type PrometheusExporter struct {
	Enabled    bool           `yaml:"enabled"`
	PullPolicy string         `yaml:"pullPolicy"`
	Repository string         `yaml:"repository"`
	Resources  *k8s.Resources `yaml:"resources"`
	Tag        string         `yaml:"tag"`
}
type RateLimit struct {
	Create bool `yaml:"create"`
}
type Rbac struct {
	Create              bool     `yaml:"create"`
	PodSecurityPolicies struct{} `yaml:"podSecurityPolicies"`
}
type ReadinessProbe struct {
	FailureThreshold    int `yaml:"failureThreshold"`
	InitialDelaySeconds int `yaml:"initialDelaySeconds"`
	PeriodSeconds       int `yaml:"periodSeconds"`
}
type RedisAnnotations struct {
	Deployment map[string]string `yaml:"deployment"`
	Service    map[string]string `yaml:"service"`
}
type Redis struct {
	Annotations  *RedisAnnotations `yaml:"annotations"`
	Create       bool              `yaml:"create"`
	Resources    *k8s.Resources    `yaml:"resources"`
	NodeSelector map[string]string `yaml:"nodeSelector"`
}
type Scope struct {
	SingleNamespace bool `yaml:"singleNamespace"`
}
type Security struct {
	PodSecurityContext       *PodSecurityContext       `yaml:"podSecurityContext"`
	ContainerSecurityContext *ContainerSecurityContext `yaml:"containerSecurityContext"`
}
type PodSecurityContext struct {
	RunAsUser int `yaml:"runAsUser"`
}
type ContainerSecurityContext struct {
	AllowPrivilegeEscalation bool `yaml:"allowPrivilegeEscalation"`
}
type Port struct {
	Name       string `yaml:"name"`
	Port       uint16 `yaml:"port,omitempty"`
	TargetPort uint16 `yaml:"targetPort,omitempty"`
	NodePort   uint16 `yaml:"nodePort,omitempty"`
}
type Service struct {
	Annotations    interface{} `yaml:"annotations,omitempty"`
	Ports          []*Port     `yaml:"ports"`
	Type           string      `yaml:"type"`
	LoadBalancerIP string      `yaml:"loadBalancerIP,omitempty"`
}
type ServiceAccount struct {
	Create bool        `yaml:"create"`
	Name   interface{} `yaml:"name"`
}

type Values struct {
	AdminService           *AdminService       `yaml:"adminService"`
	Affinity               *k8s.Affinity       `yaml:"affinity"`
	AmbassadorConfig       string              `yaml:"ambassadorConfig"`
	AuthService            *AuthService        `yaml:"authService"`
	Autoscaling            *Autoscaling        `yaml:"autoscaling"`
	Crds                   *Crds               `yaml:"crds"`
	CreateDevPortalMapping bool                `yaml:"createDevPortalMappings"`
	DaemonSet              bool                `yaml:"daemonSet"`
	DeploymentAnnotations  map[string]string   `yaml:"deploymentAnnotations"`
	DeploymentStrategy     *DeploymentStrategy `yaml:"deploymentStrategy"`
	DNSPolicy              string              `yaml:"dnsPolicy"`
	Env                    map[string]string   `yaml:"env"`
	FullnameOverride       string              `yaml:"fullnameOverride"`
	HostNetwork            bool                `yaml:"hostNetwork"`
	Image                  *Image              `yaml:"image"`
	ImagePullSecrets       []interface{}       `yaml:"imagePullSecrets"`
	InitContainers         []interface{}       `yaml:"initContainers"`
	LicenseKey             *LicenseKey         `yaml:"licenseKey"`
	LivenessProbe          *LivenessProbe      `yaml:"livenessProbe"`
	NameOverride           string              `yaml:"nameOverride"`
	NodeSelector           map[string]string   `yaml:"nodeSelector"`
	PodAnnotations         map[string]string   `yaml:"podAnnotations"`
	PodDisruptionBudget    struct{}            `yaml:"podDisruptionBudget"`
	PodLabels              map[string]string   `yaml:"podLabels"`
	PriorityClassName      string              `yaml:"priorityClassName"`
	PrometheusExporter     *PrometheusExporter `yaml:"prometheusExporter"`
	RateLimit              *RateLimit          `yaml:"rateLimit"`
	Rbac                   *Rbac               `yaml:"rbac"`
	ReadinessProbe         *ReadinessProbe     `yaml:"readinessProbe"`
	Redis                  *Redis              `yaml:"redis"`
	RedisURL               interface{}         `yaml:"redisURL"`
	ReplicaCount           int                 `yaml:"replicaCount"`
	Resources              *k8s.Resources      `yaml:"resources"`
	Scope                  *Scope              `yaml:"scope"`
	Security               *Security           `yaml:"security"`
	Service                *Service            `yaml:"service"`
	ServiceAccount         *ServiceAccount     `yaml:"serviceAccount"`
	SidecarContainers      []interface{}       `yaml:"sidecarContainers"`
	Tolerations            []corev1.Toleration `yaml:"tolerations"`
	VolumeMounts           []interface{}       `yaml:"volumeMounts"`
	Volumes                []interface{}       `yaml:"volumes"`
}
