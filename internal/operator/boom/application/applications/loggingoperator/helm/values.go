package helm

import corev1 "k8s.io/api/core/v1"

type Values struct {
	ReplicaCount     int                 `yaml:"replicaCount"`
	Image            Image               `yaml:"image"`
	ImagePullSecrets []string            `yaml:"imagePullSecrets"`
	NameOverride     string              `yaml:"nameOverride"`
	FullnameOverride string              `yaml:"fullnameOverride"`
	Resources        Resources           `yaml:"resources"`
	NodeSelector     map[string]string   `yaml:"nodeSelector,omitempty"`
	Tolerations      []corev1.Toleration `yaml:"tolerations"`
	Affinity         struct{}            `yaml:"affinity"`
	HTTP             HTTP                `yaml:"http"`
	RBAC             RBAC                `yaml:"rbac"`
}

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}

type Resources struct {
	Limits   ResourceDefinition `yaml:"limits,omitempty"`
	Requests ResourceDefinition `yaml:"requests,omitempty"`
}

type ResourceDefinition struct {
	CPU    string `yaml:"cpu,omitempty"`
	Memory string `yaml:"memory,omitempty"`
}

type HTTP struct {
	Port    int     `yaml:"port"`
	Service Service `yaml:"service"`
}

type Service struct {
	Type        string   `yaml:"type"`
	Annotations struct{} `yaml:"annotations"`
	Labels      struct{} `yaml:"labels"`
}

type RBAC struct {
	Enabled bool `yaml:"enabled"`
	PSP     PSP  `yaml:"psp"`
}

type PSP struct {
	Enabled bool `yaml:"enabled"`
}
