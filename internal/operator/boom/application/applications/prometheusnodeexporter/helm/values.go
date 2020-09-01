package helm

import corev1 "k8s.io/api/core/v1"

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}

type Service struct {
	Type        string            `yaml:"type"`
	Port        int               `yaml:"port"`
	TargetPort  int               `yaml:"targetPort"`
	NodePort    interface{}       `yaml:"nodePort"`
	Annotations map[string]string `yaml:"annotations"`
}

type Prometheus struct {
	Monitor *Monitor `yaml:"monitor"`
}

type Monitor struct {
	Enabled          bool              `yaml:"enabled"`
	AdditionalLabels map[string]string `yaml:"additionalLabels"`
	Namespace        string            `yaml:"namespace"`
	ScrapeTimeout    string            `yaml:"scrapeTimeout"`
}

type ServiceAccount struct {
	Create           bool          `yaml:"create"`
	Name             interface{}   `yaml:"name"`
	ImagePullSecrets []interface{} `yaml:"imagePullSecrets"`
}

type SecurityContext struct {
	RunAsNonRoot bool `yaml:"runAsNonRoot"`
	RunAsUser    int  `yaml:"runAsUser"`
}

type Rbac struct {
	Create     bool `yaml:"create"`
	PspEnabled bool `yaml:"pspEnabled"`
}

type Toleration struct {
	Effect   string `yaml:"effect"`
	Operator string `yaml:"operator"`
}

type Values struct {
	FullnameOverride      string                       `yaml:"fullnameOverride,omitempty"`
	Image                 *Image                       `yaml:"image"`
	Service               *Service                     `yaml:"service"`
	Prometheus            *Prometheus                  `yaml:"prometheus"`
	ServiceAccount        *ServiceAccount              `yaml:"serviceAccount"`
	SecurityContext       *SecurityContext             `yaml:"securityContext"`
	Rbac                  *Rbac                        `yaml:"rbac"`
	Endpoints             []interface{}                `yaml:"endpoints"`
	HostNetwork           bool                         `yaml:"hostNetwork"`
	Affinity              interface{}                  `yaml:"affinity"`
	NodeSelector          map[string]string            `yaml:"nodeSelector"`
	Tolerations           []*Toleration                `yaml:"tolerations"`
	ExtraArgs             []string                     `yaml:"extraArgs"`
	ExtraHostVolumeMounts interface{}                  `yaml:"extraHostVolumeMounts"`
	Configmaps            interface{}                  `yaml:"configmaps"`
	PodLabels             map[string]string            `yaml:"podLabels"`
	Resources             *corev1.ResourceRequirements `yaml:"resources"`
}
