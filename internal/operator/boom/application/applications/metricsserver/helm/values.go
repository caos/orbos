package helm

type Rbac struct {
	Create bool `yaml:"create"`
}
type ServiceAccount struct {
	Create                       bool   `yaml:"create"`
	Name                         string `yaml:"name,omitempty"`
	AutomountServiceAccountToken bool   `yaml:"automountServiceAccountToken"`
}
type APIService struct {
	Create bool `yaml:"create"`
}
type Image struct {
	Registry   string `yaml:"registry"`
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}
type HTTPGet struct {
	Path   string `yaml:"path,omitempty"`
	Port   string `yaml:"port,omitempty"`
	Scheme string `yaml:"scheme,omitempty"`
}
type Probe struct {
	Enabled          bool     `yaml:"enabled"`
	FailureThreshold int      `yaml:"failureThreshold"`
	HTTPGet          *HTTPGet `yaml:"httpGet,omitempty"`
	PeriodSeconds    int      `yaml:"periodSeconds"`
}
type Capabilities struct {
	Drop []string `yaml:"drop,omitempty"`
}
type ContainerSecurityContext struct {
	Enabled                bool `yaml:"enabled,omitempty"`
	ReadOnlyRootFilesystem bool `yaml:"readOnlyRootFilesystem,omitempty"`
	RunAsNonRoot           bool `yaml:"runAsNonRoot,omitempty"`
}
type PodSecurityContext struct {
	Enabled bool `yaml:"enabled,omitempty"`
}
type Service struct {
	Annotations map[string]string `yaml:"annotations,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Port        int               `yaml:"port,omitempty"`
	Type        string            `yaml:"type,omitempty"`
}
type PodDisruptionBudget struct {
	Enabled        bool        `yaml:"enabled"`
	MinAvailable   interface{} `yaml:"minAvailable,omitempty"`
	MaxUnavailable interface{} `yaml:"maxUnavailable,omitempty"`
}
type NodeAffinityPreset struct {
	Type   string   `yaml:"type"`
	Key    string   `yaml:"key"`
	Values []string `yaml:"values"`
}
type HostNetwork struct {
	Enabled bool `yaml:"enabled"`
}
type Values struct {
	Image                     *Image                    `yaml:"image,omitempty"`
	TestImage                 *Image                    `yaml:"testImage,omitempty"`
	HostAliases               []string                  `yaml:"hostAliases,omitempty"`
	Replicas                  int                       `yaml:"replicas,omitempty"`
	Rbac                      *Rbac                     `yaml:"rbac,omitempty"`
	ServiceAccount            *ServiceAccount           `yaml:"serviceAccount,omitempty"`
	APIService                *APIService               `yaml:"apiService,omitempty"`
	SecurePort                string                    `yaml:"securePort"`
	HostNetwork               *HostNetwork              `yaml:"hostNetwork,omitempty"`
	Command                   []string                  `yaml:"command"`
	ExtraArgs                 []string                  `yaml:"extraArgs"`
	PodLabels                 map[string]string         `yaml:"podLabels,omitempty"`
	PodAnnotations            map[string]string         `yaml:"podAnnotations,omitempty"`
	PodAffinityPreset         string                    `yaml:"podAffinityPreset"`
	PodAntiAffinityPreset     string                    `yaml:"podAntiAffinityPreset"`
	PodDisruptionBudget       *PodDisruptionBudget      `yaml:"podDisruptionBudget,omitempty"`
	NodeAffinityPreset        *NodeAffinityPreset       `yaml:"nodeAffinityPreset"`
	Affinity                  struct{}                  `yaml:"affinity,omitempty"`
	TopologySpreadConstraints []string                  `yaml:"topologySpreadConstraints"`
	NodeSelector              struct{}                  `yaml:"nodeSelector,omitempty"`
	Tolerations               []interface{}             `yaml:"tolerations,omitempty"`
	Service                   *Service                  `yaml:"service,omitempty"`
	Resources                 struct{}                  `yaml:"resources,omitempty"`
	LivenessProbe             *Probe                    `yaml:"livenessProbe,omitempty"`
	ReadinessProbe            *Probe                    `yaml:"readinessProbe,omitempty"`
	CustomLivenessProbe       struct{}                  `yaml:"customLivenessProbe"`
	CustomReadinessProbe      struct{}                  `yaml:"customReadinessProbe"`
	ContainerSecurityContext  *ContainerSecurityContext `yaml:"containerSecurityContext,omitempty"`
	PodSecurityContext        *PodSecurityContext       `yaml:"podSecurityContext,omitempty"`
}
