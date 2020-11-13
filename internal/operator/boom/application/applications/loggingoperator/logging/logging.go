package logging

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"
)

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
	NodeSelector     map[string]string
	Tolerations      k8s.Tolerations
	Fluentd          *latest.Fluentd
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
	NodeSelector        map[string]string  `yaml:"nodeSelector,omitempty"`
	Tolerations         k8s.Tolerations    `yaml:"tolerations,omitempty"`
	Resources           *k8s.Resources     `yaml:"resources,omitempty"`
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
	NodeSelector        map[string]string  `yaml:"nodeSelector,omitempty"`
	Tolerations         k8s.Tolerations    `yaml:"tolerations,omitempty"`
	Resources           *k8s.Resources     `yaml:"resources,omitempty"`
}
type Spec struct {
	Fluentd                                      *Fluentd   `yaml:"fluentd"`
	Fluentbit                                    *Fluentbit `yaml:"fluentbit"`
	WatchNamespaces                              []string   `yaml:"watchNamespaces"`
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

func New(spec *latest.LogCollection) *Logging {
	values := &Logging{
		APIVersion: "logging.banzaicloud.io/v1beta1",
		Kind:       "Logging",
		Metadata: &Metadata{
			Name:      "logging",
			Namespace: "caos-system",
		},
		Spec: &Spec{
			FlowConfigCheckDisabled:                      true,
			EnableRecreateWorkloadOnImmutableFieldChange: true,
			ControlNamespace:                             "caos-system",
			Fluentd: &Fluentd{
				Metrics: &Metrics{
					Port: 8080,
				},
				DisablePvc:   true,
				Tolerations:  k8s.Tolerations{},
				NodeSelector: map[string]string{},
				Scaling: &Scaling{
					Replicas: 1,
				},
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
				Tolerations:  k8s.Tolerations{},
				NodeSelector: map[string]string{},
			},
		},
	}

	if spec == nil {
		return values
	}

	if spec.WatchNamespaces != nil {
		values.Spec.WatchNamespaces = spec.WatchNamespaces
	}

	if spec.Fluentd != nil {
		if spec.Fluentd.Replicas != nil {
			values.Spec.Fluentd.Scaling.Replicas = *spec.Fluentd.Replicas
		}
		if spec.Fluentd.NodeSelector != nil {
			for k, v := range spec.Fluentd.NodeSelector {
				values.Spec.Fluentd.NodeSelector[k] = v
			}
		}

		if spec.Fluentd.Tolerations != nil {
			for _, tol := range spec.Fluentd.Tolerations {
				values.Spec.Fluentd.Tolerations = append(spec.Fluentd.Tolerations, tol)
			}
		}

		if spec.Fluentd.Resources != nil {
			values.Spec.Fluentd.Resources = spec.Fluentd.Resources
		}

		if spec.Fluentd.PVC != nil {
			values.Spec.Fluentd.DisablePvc = false
			values.Spec.Fluentd.BufferStorageVolume = &KubernetesStorage{
				Pvc: &Pvc{
					PvcSpec: &PvcSpec{
						StorageClassName: spec.Fluentd.PVC.StorageClass,
						Resources: &Resources{
							Requests: &Requests{
								Storage: spec.Fluentd.PVC.Size,
							},
						},
						AccessModes: []string{"ReadWriteOnce"},
					},
				},
			}
			if spec.Fluentd.PVC.AccessModes != nil {
				values.Spec.Fluentd.BufferStorageVolume.Pvc.PvcSpec.AccessModes = spec.Fluentd.PVC.AccessModes
			}
		}
	}

	if spec.Fluentbit == nil {
		return values
	}

	if spec.Fluentbit.Resources != nil {
		values.Spec.Fluentbit.Resources = spec.Fluentbit.Resources
	}

	if spec.Fluentbit.NodeSelector != nil {
		for k, v := range spec.Fluentbit.NodeSelector {
			values.Spec.Fluentbit.NodeSelector[k] = v
		}
	}

	if spec.Fluentbit.Tolerations != nil {
		for _, tol := range spec.Fluentbit.Tolerations {
			values.Spec.Fluentbit.Tolerations = append(spec.Fluentbit.Tolerations, tol)
		}
	}

	return values
}
