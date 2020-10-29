package helm

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"
	corev1 "k8s.io/api/core/v1"
)

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}

type Service struct {
	Port           int               `yaml:"port"`
	Type           string            `yaml:"type"`
	NodePort       int               `yaml:"nodePort"`
	LoadBalancerIP string            `yaml:"loadBalancerIP"`
	Annotations    map[string]string `yaml:"annotations"`
}

type Rbac struct {
	Create bool `yaml:"create"`
}

type ServiceAccount struct {
	Create           bool          `yaml:"create"`
	Name             interface{}   `yaml:"name"`
	ImagePullSecrets []interface{} `yaml:"imagePullSecrets"`
}

type Prometheus struct {
	Monitor *Monitor `yaml:"monitor"`
}
type Monitor struct {
	Enabled          bool              `yaml:"enabled"`
	AdditionalLabels map[string]string `yaml:"additionalLabels"`
	Namespace        string            `yaml:"namespace"`
	HonorLabels      bool              `yaml:"honorLabels"`
}

type PodSecurityPolicy struct {
	Enabled     bool              `yaml:"enabled"`
	Annotations map[string]string `yaml:"annotations"`
}

type SecurityContext struct {
	Enabled   bool `yaml:"enabled"`
	RunAsUser int  `yaml:"runAsUser"`
	FsGroup   int  `yaml:"fsGroup"`
}

type Collectors struct {
	Certificatesigningrequests bool `yaml:"certificatesigningrequests"`
	Configmaps                 bool `yaml:"configmaps"`
	Cronjobs                   bool `yaml:"cronjobs"`
	Daemonsets                 bool `yaml:"daemonsets"`
	Deployments                bool `yaml:"deployments"`
	Endpoints                  bool `yaml:"endpoints"`
	Horizontalpodautoscalers   bool `yaml:"horizontalpodautoscalers"`
	Ingresses                  bool `yaml:"ingresses"`
	Jobs                       bool `yaml:"jobs"`
	Limitranges                bool `yaml:"limitranges"`
	Namespaces                 bool `yaml:"namespaces"`
	Nodes                      bool `yaml:"nodes"`
	Persistentvolumeclaims     bool `yaml:"persistentvolumeclaims"`
	Persistentvolumes          bool `yaml:"persistentvolumes"`
	Poddisruptionbudgets       bool `yaml:"poddisruptionbudgets"`
	Pods                       bool `yaml:"pods"`
	Replicasets                bool `yaml:"replicasets"`
	Replicationcontrollers     bool `yaml:"replicationcontrollers"`
	Resourcequotas             bool `yaml:"resourcequotas"`
	Secrets                    bool `yaml:"secrets"`
	Services                   bool `yaml:"services"`
	Statefulsets               bool `yaml:"statefulsets"`
	Storageclasses             bool `yaml:"storageclasses"`
	Verticalpodautoscalers     bool `yaml:"verticalpodautoscalers"`
}

type Selector struct {
	MatchLabels map[string]string `yaml:"matchLabels,omitempty"`
}

type PodDisruptionBudget struct {
	MaxUnavailable int       `yaml:"maxUnavailable,omitempty"`
	Selector       *Selector `yaml:"selector,omitempty"`
}

type Values struct {
	FullnameOverride    string               `yaml:"fullnameOverride,omitempty"`
	PrometheusScrape    bool                 `yaml:"prometheusScrape"`
	Image               *Image               `yaml:"image"`
	Replicas            int                  `yaml:"replicas"`
	Service             *Service             `yaml:"service"`
	CustomLabels        map[string]string    `yaml:"customLabels"`
	HostNetwork         bool                 `yaml:"hostNetwork"`
	Rbac                *Rbac                `yaml:"rbac"`
	ServiceAccount      *ServiceAccount      `yaml:"serviceAccount"`
	Prometheus          *Prometheus          `yaml:"prometheus"`
	PodSecurityPolicy   *PodSecurityPolicy   `yaml:"podSecurityPolicy"`
	SecurityContext     *SecurityContext     `yaml:"securityContext"`
	NodeSelector        map[string]string    `yaml:"nodeSelector"`
	Affinity            *k8s.Affinity        `yaml:"affinity"`
	Tolerations         []corev1.Toleration  `yaml:"tolerations"`
	PodAnnotations      map[string]string    `yaml:"podAnnotations"`
	Collectors          *Collectors          `yaml:"collectors"`
	PodDisruptionBudget *PodDisruptionBudget `yaml:"podDisruptionBudget"`
	Resources           *k8s.Resources       `yaml:"resources"`
}
