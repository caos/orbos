package helm

import (
	"github.com/caos/orbos/v5/pkg/kubernetes/k8s"
)

type Values struct {
	ReplicaCount     int               `yaml:"replicaCount"`
	Image            Image             `yaml:"image"`
	ImagePullSecrets []string          `yaml:"imagePullSecrets"`
	NameOverride     string            `yaml:"nameOverride"`
	FullnameOverride string            `yaml:"fullnameOverride"`
	Resources        *k8s.Resources    `yaml:"resources"`
	NodeSelector     map[string]string `yaml:"nodeSelector"`
	Tolerations      k8s.Tolerations   `yaml:"tolerations"`
	HTTP             HTTP              `yaml:"http"`
	RBAC             RBAC              `yaml:"rbac"`
}

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
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
