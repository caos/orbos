package logging

type Storage struct {
	StorageClassName string
	AccessModes      []string
	Storage          string
}

type Config struct {
	Name             string
	Namespace        string
	ControlNamespace string
	Replicas         int
	FluentdPVC       *Storage
	FluentbitPVC     *Storage
}
type Requests struct {
	Storage string `yaml:"storage,omitempty"`
}
type Resources struct {
	Requests *Requests `yaml:"requests,omitempty"`
}
type PvcSpec struct {
	AccessModes      []string   `yaml:"accessModes,omitempty"`
	Resources        *Resources `yaml:"resources,omitempty"`
	StorageClassName string     `yaml:"storageClassName,omitempty"`
}
type Pvc struct {
	PvcSpec *PvcSpec `yaml:"spec,omitempty"`
}
type KubernetesStorage struct {
	Pvc *Pvc `yaml:"pvc,omitempty"`
}
type Scaling struct {
	Replicas int `yaml:"replicas"`
}
type Fluentd struct {
	Metrics             *Metrics           `yaml:"metrics,omitempty"`
	BufferStorageVolume *KubernetesStorage `yaml:"bufferStorageVolume,omitempty"`
	LogLevel            string             `yaml:"logLevel,omitempty"`
	DisablePvc          bool               `yaml:"disablePvc"`
	Scaling             *Scaling           `yaml:"scaling,omitempty"`
}
type Metrics struct {
	Port int `yaml:"port"`
}
type Image struct {
	PullPolicy string `yaml:"pullPolicy"`
	Repository string `yaml:"repository"`
	Tag        string `yaml:"tag"`
}

type FilterKubernetes struct {
	KubeTagPrefix string `yaml:"Kube_Tag_Prefix"`
}

type Fluentbit struct {
	Metrics             *Metrics           `yaml:"metrics,omitempty"`
	FilterKubernetes    *FilterKubernetes  `yaml:"filterKubernetes,omitempty"`
	Image               *Image             `yaml:"image,omitempty"`
	BufferStorageVolume *KubernetesStorage `yaml:"bufferStorageVolume,omitempty"`
}
type Spec struct {
	Fluentd                                      *Fluentd   `yaml:"fluentd"`
	Fluentbit                                    *Fluentbit `yaml:"fluentbit"`
	ControlNamespace                             string     `yaml:"controlNamespace"`
	EnableRecreateWorkloadOnImmutableFieldChange bool       `yaml:"enableRecreateWorkloadOnImmutableFieldChange"`
	FlowConfigCheckDisabled                      bool       `yaml:"flowConfigCheckDisabled"`
}
type Metadata struct {
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}
type Logging struct {
	APIVersion string    `yaml:"apiVersion"`
	Kind       string    `yaml:"kind"`
	Metadata   *Metadata `yaml:"metadata"`
	Spec       *Spec     `yaml:"spec"`
}

func New(conf *Config) *Logging {
	values := &Logging{
		APIVersion: "logging.banzaicloud.io/v1beta1",
		Kind:       "Logging",
		Metadata: &Metadata{
			Name:      conf.Name,
			Namespace: conf.Namespace,
		},
		Spec: &Spec{
			FlowConfigCheckDisabled:                      true,
			EnableRecreateWorkloadOnImmutableFieldChange: true,
			ControlNamespace:                             conf.ControlNamespace,
			Fluentd: &Fluentd{
				Metrics: &Metrics{
					Port: 8080,
				},
				DisablePvc: true,
			},
			Fluentbit: &Fluentbit{
				Metrics: &Metrics{
					Port: 8080,
				},
				Image: &Image{
					Repository: "fluent/fluent-bit",
					Tag:        "1.3.6",
					PullPolicy: "IfNotPresent",
				},
			},
		},
	}
	if conf.FluentdPVC != nil {
		values.Spec.Fluentd.BufferStorageVolume = &KubernetesStorage{
			Pvc: &Pvc{
				PvcSpec: &PvcSpec{
					StorageClassName: conf.FluentdPVC.StorageClassName,
					Resources: &Resources{
						Requests: &Requests{
							Storage: conf.FluentdPVC.Storage,
						},
					},
				},
			},
		}
		values.Spec.Fluentd.DisablePvc = false

		if conf.FluentdPVC.AccessModes != nil {
			values.Spec.Fluentd.BufferStorageVolume.Pvc.PvcSpec.AccessModes = conf.FluentdPVC.AccessModes
		} else {
			values.Spec.Fluentd.BufferStorageVolume.Pvc.PvcSpec.AccessModes = []string{"ReadWriteOnce"}
		}
	}

	if conf.Replicas != 0 {
		values.Spec.Fluentd.Scaling = &Scaling{
			Replicas: conf.Replicas,
		}
	}
	return values
}
