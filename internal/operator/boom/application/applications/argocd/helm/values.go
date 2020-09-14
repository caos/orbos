package helm

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/k8s"
	corev1 "k8s.io/api/core/v1"
)

type Image struct {
	Repository      string `yaml:"repository"`
	Tag             string `yaml:"tag"`
	ImagePullPolicy string `yaml:"imagePullPolicy"`
}
type Global struct {
	Image           *Image   `yaml:"image"`
	SecurityContext struct{} `yaml:"securityContext"`
}
type Args struct {
	StatusProcessors    string `yaml:"statusProcessors"`
	OperationProcessors string `yaml:"operationProcessors"`
}

type ReadinessProbe struct {
	FailureThreshold    int `yaml:"failureThreshold"`
	InitialDelaySeconds int `yaml:"initialDelaySeconds"`
	PeriodSeconds       int `yaml:"periodSeconds"`
	SuccessThreshold    int `yaml:"successThreshold"`
	TimeoutSeconds      int `yaml:"timeoutSeconds"`
}
type LivenessProbe struct {
	FailureThreshold    int `yaml:"failureThreshold"`
	InitialDelaySeconds int `yaml:"initialDelaySeconds"`
	PeriodSeconds       int `yaml:"periodSeconds"`
	SuccessThreshold    int `yaml:"successThreshold"`
	TimeoutSeconds      int `yaml:"timeoutSeconds"`
}
type Service struct {
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
	Port        int               `yaml:"port"`
}
type ServiceAccount struct {
	Create bool   `yaml:"create"`
	Name   string `yaml:"name"`
}
type MetricsService struct {
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
	ServicePort int               `yaml:"servicePort"`
}
type ServiceMonitor struct {
	Enabled bool `yaml:"enabled"`
}
type Rules struct {
	Enabled bool          `yaml:"enabled"`
	Spec    []interface{} `yaml:"spec"`
}
type Metrics struct {
	Enabled        bool            `yaml:"enabled"`
	Service        *MetricsService `yaml:"service"`
	ServiceMonitor *ServiceMonitor `yaml:"serviceMonitor"`
	Rules          *Rules          `yaml:"rules"`
}
type ClusterAdminAccess struct {
	Enabled bool `yaml:"enabled"`
}
type ConfigMap struct {
	Name        string `yaml:"name,omitempty"`
	DefaultMode int    `yaml:"defaultMode,omitempty"`
}
type VolumeSecret struct {
	SecretName  string  `yaml:"secretName,omitempty"`
	Items       []*Item `yaml:"items,omitempty"`
	DefaultMode int     `yaml:"defaultMode"`
}

type Item struct {
	Key  string `yaml:"key"`
	Path string `yaml:"path"`
}

type Volume struct {
	Secret    *VolumeSecret `yaml:"secret,omitempty"`
	ConfigMap *ConfigMap    `yaml:"configMap,omitempty"`
	Name      string        `yaml:"name"`
	EmptyDir  struct{}      `yaml:"emptyDir,omitempty"`
}

type VolumeMount struct {
	Name      string `yaml:"name"`
	MountPath string `yaml:"mountPath,omitempty"`
	SubPath   string `yaml:"subPath,omitempty"`
	ReadOnly  bool   `yaml:"readOnly,omitempty"`
}
type Controller struct {
	Name               string              `yaml:"name"`
	Image              *Image              `yaml:"image"`
	Args               *Args               `yaml:"args"`
	LogLevel           string              `yaml:"logLevel"`
	ExtraArgs          struct{}            `yaml:"extraArgs"`
	Env                []interface{}       `yaml:"env"`
	PodAnnotations     map[string]string   `yaml:"podAnnotations"`
	PodLabels          map[string]string   `yaml:"podLabels"`
	ContainerPort      int                 `yaml:"containerPort"`
	ReadinessProbe     *ReadinessProbe     `yaml:"readinessProbe"`
	LivenessProbe      *LivenessProbe      `yaml:"livenessProbe"`
	VolumeMounts       []*VolumeMount      `yaml:"volumeMounts"`
	Volumes            []*Volume           `yaml:"volumes"`
	Service            *Service            `yaml:"service"`
	NodeSelector       map[string]string   `yaml:"nodeSelector"`
	Tolerations        []corev1.Toleration `yaml:"tolerations"`
	Affinity           struct{}            `yaml:"affinity"`
	PriorityClassName  string              `yaml:"priorityClassName"`
	Resources          *k8s.Resources      `yaml:"resources"`
	ServiceAccount     *ServiceAccount     `yaml:"serviceAccount"`
	Metrics            *Metrics            `yaml:"metrics"`
	ClusterAdminAccess *ClusterAdminAccess `yaml:"clusterAdminAccess"`
}
type Dex struct {
	Enabled           bool                `yaml:"enabled"`
	Name              string              `yaml:"name"`
	Image             *Image              `yaml:"image"`
	InitImage         *Image              `yaml:"initImage,omitempty"`
	Env               []interface{}       `yaml:"env"`
	ServiceAccount    *ServiceAccount     `yaml:"serviceAccount"`
	VolumeMounts      []*VolumeMount      `yaml:"volumeMounts"`
	Volumes           []*Volume           `yaml:"volumes"`
	ContainerPortHTTP int                 `yaml:"containerPortHttp"`
	ServicePortHTTP   int                 `yaml:"servicePortHttp"`
	ContainerPortGrpc int                 `yaml:"containerPortGrpc"`
	ServicePortGrpc   int                 `yaml:"servicePortGrpc"`
	NodeSelector      map[string]string   `yaml:"nodeSelector"`
	Tolerations       []corev1.Toleration `yaml:"tolerations"`
	Affinity          struct{}            `yaml:"affinity"`
	PriorityClassName string              `yaml:"priorityClassName"`
	Resources         *k8s.Resources      `yaml:"resources"`
}

type Redis struct {
	Enabled           bool                `yaml:"enabled"`
	Name              string              `yaml:"name"`
	Image             *Image              `yaml:"image"`
	ContainerPort     int                 `yaml:"containerPort"`
	ServicePort       int                 `yaml:"servicePort"`
	Env               []interface{}       `yaml:"env"`
	NodeSelector      map[string]string   `yaml:"nodeSelector"`
	Tolerations       []corev1.Toleration `yaml:"tolerations"`
	Affinity          struct{}            `yaml:"affinity"`
	PriorityClassName string              `yaml:"priorityClassName"`
	Resources         *k8s.Resources      `yaml:"resources"`
	VolumeMounts      []*VolumeMount      `yaml:"volumeMounts"`
	Volumes           []*Volume           `yaml:"volumes"`
}
type Certificate struct {
	Enabled         bool          `yaml:"enabled"`
	Domain          string        `yaml:"domain"`
	Issuer          struct{}      `yaml:"issuer"`
	AdditionalHosts []interface{} `yaml:"additionalHosts"`
}
type ServerService struct {
	Annotations      map[string]string `yaml:"annotations"`
	Labels           map[string]string `yaml:"labels"`
	Type             string            `yaml:"type"`
	ServicePortHTTP  int               `yaml:"servicePortHttp"`
	ServicePortHTTPS int               `yaml:"servicePortHttps"`
}
type Ingress struct {
	Enabled     bool              `yaml:"enabled"`
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
	Hosts       []interface{}     `yaml:"hosts"`
	Paths       []string          `yaml:"paths"`
	TLS         []interface{}     `yaml:"tls"`
}

type Route struct {
	Enabled  bool   `yaml:"enabled"`
	Hostname string `yaml:"hostname"`
}

type Config struct {
	URL                         string `yaml:"url"`
	ApplicationInstanceLabelKey string `yaml:"application.instanceLabelKey"`
	OIDC                        string `yaml:"oidc.config,omitempty"`
	Dex                         string `yaml:"dex.config,omitempty"`
	Repositories                string `yaml:"repositories,omitempty"`
	RepositoryCredentials       string `yaml:"repository.credentials,omitempty"`
	ConfigManagementPlugins     string `yaml:"configManagementPlugins,omitempty"`
}

type Server struct {
	Name                   string              `yaml:"name"`
	Image                  *Image              `yaml:"image"`
	ExtraArgs              []string            `yaml:"extraArgs"`
	Env                    []interface{}       `yaml:"env"`
	LogLevel               string              `yaml:"logLevel"`
	PodAnnotations         map[string]string   `yaml:"podAnnotations"`
	PodLabels              map[string]string   `yaml:"podLabels"`
	ContainerPort          int                 `yaml:"containerPort"`
	ReadinessProbe         *ReadinessProbe     `yaml:"readinessProbe"`
	LivenessProbe          *LivenessProbe      `yaml:"livenessProbe"`
	VolumeMounts           []*VolumeMount      `yaml:"volumeMounts"`
	Volumes                []*Volume           `yaml:"volumes"`
	NodeSelector           map[string]string   `yaml:"nodeSelector"`
	Tolerations            []corev1.Toleration `yaml:"tolerations"`
	Affinity               struct{}            `yaml:"affinity"`
	PriorityClassName      string              `yaml:"priorityClassName"`
	Resources              *k8s.Resources      `yaml:"resources"`
	Certificate            *Certificate        `yaml:"certificate"`
	Service                *ServerService      `yaml:"service"`
	Metrics                *Metrics            `yaml:"metrics"`
	ServiceAccount         *ServiceAccount     `yaml:"serviceAccount"`
	Ingress                *Ingress            `yaml:"ingress"`
	Route                  *Route              `yaml:"route"`
	Config                 *Config             `yaml:"config"`
	RbacConfig             *RbacConfig         `yaml:"rbacConfig,omitempty"`
	AdditionalApplications []interface{}       `yaml:"additionalApplications"`
	AdditionalProjects     []interface{}       `yaml:"additionalProjects"`
}
type RbacConfig struct {
	Csv     string `yaml:"policy.csv,omitempty"`
	Default string `yaml:"policy.default,omitempty"`
	Scopes  string `yaml:"scopes,omitempty"`
}

type RepoServer struct {
	Name              string              `yaml:"name"`
	Image             *Image              `yaml:"image"`
	ExtraArgs         []string            `yaml:"extraArgs"`
	Env               []interface{}       `yaml:"env"`
	LogLevel          string              `yaml:"logLevel"`
	PodAnnotations    map[string]string   `yaml:"podAnnotations"`
	PodLabels         map[string]string   `yaml:"podLabels"`
	ContainerPort     int                 `yaml:"containerPort"`
	ReadinessProbe    *ReadinessProbe     `yaml:"readinessProbe"`
	LivenessProbe     *LivenessProbe      `yaml:"livenessProbe"`
	VolumeMounts      []*VolumeMount      `yaml:"volumeMounts"`
	Volumes           []*Volume           `yaml:"volumes"`
	NodeSelector      map[string]string   `yaml:"nodeSelector"`
	Tolerations       []corev1.Toleration `yaml:"tolerations"`
	Affinity          struct{}            `yaml:"affinity"`
	PriorityClassName string              `yaml:"priorityClassName"`
	Resources         *k8s.Resources      `yaml:"resources"`
	Service           *Service            `yaml:"service"`
	Metrics           *Metrics            `yaml:"metrics"`
	ServiceAccount    *ServiceAccount     `yaml:"serviceAccount"`
}
type Data struct {
	Data map[string]string `yaml:"data"`
}
type Secret struct {
	CreateSecret          bool   `yaml:"createSecret"`
	GithubSecret          string `yaml:"githubSecret"`
	GitlabSecret          string `yaml:"gitlabSecret"`
	BitbucketServerSecret string `yaml:"bitbucketServerSecret"`
	BitbucketUUD          string `yaml:"bitbucketUUÌD"`
	GogsSecret            string `yaml:"gogsSecret"`
	ArgocdServerTLSConfig struct {
	} `yaml:"argocdServerTlsConfig"`
}
type Configs struct {
	KnownHosts *Data   `yaml:"knownHosts"`
	TLSCerts   *Data   `yaml:"tlsCerts"`
	Secret     *Secret `yaml:"secret"`
}

type RedisHAConfig struct {
	Save string
}

type RedisHAC struct {
	MasterGroupName string
	Config          *RedisHAConfig
}

type HAProxyMetrics struct {
	Enabled bool
}

type HAProxy struct {
	Enabled bool
	Metrics *HAProxyMetrics
}

type Exporter struct {
	Enabled bool
}

type RedisHA struct {
	Enabled           bool
	Exporter          *Exporter
	PersistenteVolume *PV
	Redis             *RedisHAC
	HAProxy           *HAProxy
}

type PV struct {
	Enabled bool
}

type Values struct {
	NameOverride     string      `yaml:"nameOverride,omitempty"`
	FullnameOverride string      `yaml:"fullnameOverride,omitempty"`
	InstallCRDs      bool        `yaml:"installCRDs"`
	Global           *Global     `yaml:"global"`
	Controller       *Controller `yaml:"controller"`
	Dex              *Dex        `yaml:"dex"`
	Redis            *Redis      `yaml:"redis"`
	RedisHA          *RedisHA    `yaml:"redis-ha"`
	Server           *Server     `yaml:"server"`
	RepoServer       *RepoServer `yaml:"repoServer"`
	Configs          *Configs    `yaml:"configs"`
}
