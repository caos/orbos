package helm

import "github.com/caos/orbos/pkg/kubernetes/k8s"

type Image struct {
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
	PullPolicy string `yaml:"pullPolicy"`
}
type SecurityContext struct {
	RunAsUser    int  `yaml:"runAsUser"`
	RunAsGroup   int  `yaml:"runAsGroup"`
	RunAsNonRoot bool `yaml:"runAsNonRoot"`
	FsGroup      int  `yaml:"fsGroup"`
}
type Requests struct {
	CPU              string `yaml:"cpu"`
	Memory           string `yaml:"memory"`
	EphemeralStorage string `yaml:"ephemeral-storage"`
}
type Operator struct {
	ClusterDomain    string           `yaml:"clusterDomain"`
	NsToWatch        string           `yaml:"nsToWatch"`
	Image            *Image           `yaml:"image"`
	ImagePullSecrets []interface{}    `yaml:"imagePullSecrets"`
	ReplicaCount     int              `yaml:"replicaCount"`
	SecurityContext  *SecurityContext `yaml:"securityContext"`
	Resources        *k8s.Resources   `yaml:"resources"`
}
type Console struct {
	Image        *Image         `yaml:"image"`
	ReplicaCount int            `yaml:"replicaCount"`
	Resources    *k8s.Resources `yaml:"resources"`
}
type Pools struct {
	Servers          int               `yaml:"servers"`
	VolumesPerServer int               `yaml:"volumesPerServer"`
	Size             string            `yaml:"size"`
	StorageClassName string            `yaml:"storageClassName"`
	Tolerations      struct{}          `yaml:"tolerations"`
	NodeSelector     map[string]string `yaml:"nodeSelector"`
	Affinity         struct{}          `yaml:"affinity"`
	Resources        *k8s.Resources    `yaml:"resources"`
	SecurityContext  struct{}          `yaml:"securityContext"`
}
type Secrets struct {
	Enabled   bool   `yaml:"enabled"`
	Name      string `yaml:"name"`
	AccessKey string `yaml:"accessKey"`
	SecretKey string `yaml:"secretKey"`
}
type Metrics struct {
	Enabled bool `yaml:"enabled"`
	Port    int  `yaml:"port"`
}
type Certificate struct {
	ExternalCertSecret struct{} `yaml:"externalCertSecret"`
	RequestAutoCert    bool     `yaml:"requestAutoCert"`
	CertConfig         struct{} `yaml:"certConfig"`
}
type S3 struct {
	BucketDNS bool `yaml:"bucketDNS"`
}
type ConsoleTenantSecrets struct {
	Enabled    bool   `yaml:"enabled"`
	Name       string `yaml:"name"`
	Passphrase string `yaml:"passphrase"`
	Salt       string `yaml:"salt"`
	AccessKey  string `yaml:"accessKey"`
	SecretKey  string `yaml:"secretKey"`
}
type ConsoleTenant struct {
	Image        *Image                `yaml:"image"`
	ReplicaCount int                   `yaml:"replicaCount"`
	Secrets      *ConsoleTenantSecrets `yaml:"secrets"`
}
type Tenants struct {
	Name                string         `yaml:"name"`
	Namespace           string         `yaml:"namespace"`
	Image               *Image         `yaml:"image"`
	ImagePullSecrets    []interface{}  `yaml:"imagePullSecrets"`
	Scheduler           struct{}       `yaml:"scheduler"`
	Pools               []*Pools       `yaml:"pools"`
	MountPath           string         `yaml:"mountPath"`
	SubPath             string         `yaml:"subPath"`
	Secrets             *Secrets       `yaml:"secrets"`
	Metrics             *Metrics       `yaml:"metrics"`
	Certificate         *Certificate   `yaml:"certificate"`
	S3                  *S3            `yaml:"s3"`
	PodManagementPolicy string         `yaml:"podManagementPolicy"`
	ServiceMetadata     struct{}       `yaml:"serviceMetadata"`
	Env                 struct{}       `yaml:"env"`
	PriorityClassName   string         `yaml:"priorityClassName"`
	Console             *ConsoleTenant `yaml:"console"`
}
type Values struct {
	Operator *Operator  `yaml:"operator"`
	Console  *Console   `yaml:"console"`
	Tenants  []*Tenants `yaml:"tenants"`
}
