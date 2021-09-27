package helm

import (
	"github.com/caos/orbos/v5/pkg/kubernetes/k8s"
)

type Tracing struct {
	JaegerAgentHost interface{} `yaml:"jaegerAgentHost"`
}
type Kvstore struct {
	Store string `yaml:"store"`
}
type Ring struct {
	Kvstore           *Kvstore `yaml:"kvstore"`
	ReplicationFactor int      `yaml:"replication_factor"`
}
type Lifecycler struct {
	Ring *Ring `yaml:"ring"`
}
type Ingester struct {
	ChunkIdlePeriod   string      `yaml:"chunk_idle_period"`
	ChunkBlockSize    int         `yaml:"chunk_block_size"`
	ChunkRetainPeriod string      `yaml:"chunk_retain_period"`
	Lifecycler        *Lifecycler `yaml:"lifecycler"`
}
type LimitsConfig struct {
	EnforceMetricName      bool   `yaml:"enforce_metric_name"`
	RejectOldSamples       bool   `yaml:"reject_old_samples"`
	RejectOldSamplesMaxAge string `yaml:"reject_old_samples_max_age"`
}
type Index struct {
	Prefix string `yaml:"prefix"`
	Period string `yaml:"period"`
}
type Chunks struct {
	Prefix string `yaml:"prefix"`
	Period string `yaml:"period"`
}
type SchemaConfig struct {
	From        string  `yaml:"from"`
	Store       string  `yaml:"store"`
	ObjectStore string  `yaml:"object_store"`
	Schema      string  `yaml:"schema"`
	Index       *Index  `yaml:"index"`
	Chunks      *Chunks `yaml:"chunks"`
}
type SchemaConfigs struct {
	Configs []*SchemaConfig `yaml:"configs"`
}
type Server struct {
	HTTPListenPort int `yaml:"http_listen_port"`
}
type Boltdb struct {
	Directory string `yaml:"directory"`
}
type Filesystem struct {
	Directory string `yaml:"directory"`
}
type StorageConfig struct {
	Boltdb     *Boltdb     `yaml:"boltdb"`
	Filesystem *Filesystem `yaml:"filesystem"`
}
type ChunkStoreConfig struct {
	MaxLookBackPeriod string `yaml:"max_look_back_period"`
}
type TableManager struct {
	RetentionDeletesEnabled bool   `yaml:"retention_deletes_enabled"`
	RetentionPeriod         string `yaml:"retention_period"`
}
type Config struct {
	AuthEnabled      bool              `yaml:"auth_enabled"`
	Ingester         *Ingester         `yaml:"ingester"`
	LimitsConfig     *LimitsConfig     `yaml:"limits_config"`
	SchemaConfig     *SchemaConfigs    `yaml:"schema_config"`
	Server           *Server           `yaml:"server"`
	StorageConfig    *StorageConfig    `yaml:"storage_config"`
	ChunkStoreConfig *ChunkStoreConfig `yaml:"chunk_store_config"`
	TableManager     *TableManager     `yaml:"table_manager"`
}
type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}
type HTTPGet struct {
	Path string `yaml:"path"`
	Port string `yaml:"port"`
}
type LivenessProbe struct {
	HTTPGet             *HTTPGet `yaml:"httpGet"`
	InitialDelaySeconds int      `yaml:"initialDelaySeconds"`
}
type NetworkPolicy struct {
	Enabled bool `yaml:"enabled"`
}
type Persistence struct {
	Enabled          bool              `yaml:"enabled"`
	AccessModes      []string          `yaml:"accessModes"`
	Size             string            `yaml:"size"`
	Annotations      map[string]string `yaml:"annotations"`
	StorageClassName string            `yaml:"storageClassName"`
}

type Rbac struct {
	Create     bool `yaml:"create"`
	PspEnabled bool `yaml:"pspEnabled"`
}
type ReadinessProbe struct {
	HTTPGet             *HTTPGet `yaml:"httpGet"`
	InitialDelaySeconds int      `yaml:"initialDelaySeconds"`
}
type SecurityContext struct {
	FsGroup      int  `yaml:"fsGroup"`
	RunAsGroup   int  `yaml:"runAsGroup"`
	RunAsNonRoot bool `yaml:"runAsNonRoot"`
	RunAsUser    int  `yaml:"runAsUser"`
}
type Service struct {
	Type        string            `yaml:"type"`
	NodePort    interface{}       `yaml:"nodePort"`
	Port        int               `yaml:"port"`
	Annotations map[string]string `yaml:"annotations"`
	Labels      map[string]string `yaml:"labels"`
}
type ServiceAccount struct {
	Create      bool              `yaml:"create"`
	Name        interface{}       `yaml:"name"`
	Annotations map[string]string `yaml:"annotations"`
}
type UpdateStrategy struct {
	Type string `yaml:"type"`
}
type ServiceMonitor struct {
	Enabled          bool              `yaml:"enabled"`
	Interval         string            `yaml:"interval"`
	AdditionalLabels map[string]string `yaml:"additionalLabels"`
}

type Values struct {
	FullNameOverride              string            `yaml:"fullNameOverride,omitempty"`
	Affinity                      struct{}          `yaml:"affinity"`
	Annotations                   map[string]string `yaml:"annotations"`
	Tracing                       *Tracing          `yaml:"tracing"`
	Config                        *Config           `yaml:"config"`
	Image                         *Image            `yaml:"image"`
	ExtraArgs                     struct{}          `yaml:"extraArgs"`
	LivenessProbe                 *LivenessProbe    `yaml:"livenessProbe"`
	NetworkPolicy                 *NetworkPolicy    `yaml:"networkPolicy"`
	Client                        struct{}          `yaml:"client"`
	NodeSelector                  map[string]string `yaml:"nodeSelector"`
	Persistence                   *Persistence      `yaml:"persistence"`
	PodLabels                     map[string]string `yaml:"podLabels"`
	PodAnnotations                map[string]string `yaml:"podAnnotations"`
	PodManagementPolicy           string            `yaml:"podManagementPolicy"`
	Rbac                          *Rbac             `yaml:"rbac"`
	ReadinessProbe                *ReadinessProbe   `yaml:"readinessProbe"`
	Replicas                      int               `yaml:"replicas"`
	Resources                     *k8s.Resources    `yaml:"resources"`
	SecurityContext               *SecurityContext  `yaml:"securityContext"`
	Service                       *Service          `yaml:"service"`
	ServiceAccount                *ServiceAccount   `yaml:"serviceAccount"`
	TerminationGracePeriodSeconds int               `yaml:"terminationGracePeriodSeconds"`
	Tolerations                   k8s.Tolerations   `yaml:"tolerations"`
	PodDisruptionBudget           struct{}          `yaml:"podDisruptionBudget"`
	UpdateStrategy                *UpdateStrategy   `yaml:"updateStrategy"`
	ServiceMonitor                *ServiceMonitor   `yaml:"serviceMonitor"`
	InitContainers                []interface{}     `yaml:"initContainers"`
	ExtraContainers               []interface{}     `yaml:"extraContainers"`
	ExtraVolumes                  []interface{}     `yaml:"extraVolumes"`
	ExtraVolumeMounts             []interface{}     `yaml:"extraVolumeMounts"`
	ExtraPorts                    []interface{}     `yaml:"extraPorts"`
	Env                           []*Env            `yaml:"env,omitempty"`
}

type Toleration struct {
	Effect            string `yaml:"effect,omitempty"`
	Key               string `yaml:"key,omitempty"`
	Operator          string `yaml:"operator,omitempty"`
	TolerationSeconds int    `yaml:"tolerationSeconds,omitempty"`
	Value             string `yaml:"value,omitempty"`
}
type Env struct {
	Name  string
	Value string
}
