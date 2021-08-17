package logging

type ConfigOutput struct {
	Name                      string
	Namespace                 string
	URL                       string
	ClusterOutput             bool
	RemoveKeys                []string
	Labels                    map[string]string
	ExtraLabels               map[string]string
	ExtractKubernetesLabels   bool
	ConfigureKubernetesLabels bool
	EnabledNamespaces         []string
	Username                  *SecretKeyRef
	Password                  *SecretKeyRef
}

type Buffer struct {
	Timekey       string `yaml:"timekey"`
	TimekeyWait   string `yaml:"timekey_wait"`
	TimekeyUseUtc bool   `yaml:"timekey_use_utc"`
}

type Loki struct {
	URL                       string            `yaml:"url"`
	ConfigureKubernetesLabels bool              `yaml:"configure_kubernetes_labels,omitempty"`
	ExtractKubernetesLabels   bool              `yaml:"extract_kubernetes_labels,omitempty"`
	ExtraLabels               map[string]string `yaml:"extra_labels,omitempty"`
	Labels                    map[string]string `yaml:"labels,omitempty"`
	RemoveKeys                []string          `yaml:"remove_keys,omitempty"`
	Username                  *Value            `yaml:"username"`
	Password                  *Value            `yaml:"password"`
	Buffer                    *Buffer           `yaml:"buffer"`
}

type Value struct {
	ValueFrom *ValueFrom `yaml:"valueFrom,omitempty"`
}

type ValueFrom struct {
	SecretKeyRef *SecretKeyRef `yaml:"secretKeyRef,omitempty"`
}

type SecretKeyRef struct {
	Key  string `yaml:"key,omitempty"`
	Name string `yaml:"name,omitempty"`
}

type OutputSpec struct {
	Loki              *Loki    `yaml:"loki"`
	EnabledNamespaces []string `yaml:"enabledNamespaces,omitempty"`
}

type Output struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Metadata   *Metadata   `yaml:"metadata"`
	Spec       *OutputSpec `yaml:"spec"`
}

func NewOutput(clusterOutput bool, conf *ConfigOutput) *Output {
	kind := "Output"
	meta := &Metadata{
		Name:      conf.Name,
		Namespace: conf.Namespace,
	}
	if clusterOutput {
		kind = "ClusterOutput"
		meta.Namespace = ""
	}

	var username, password *Value
	if conf.Username != nil {
		username = &Value{
			ValueFrom: &ValueFrom{
				SecretKeyRef: &SecretKeyRef{
					Key:  conf.Username.Key,
					Name: conf.Username.Name,
				},
			},
		}
	}
	if conf.Password != nil {
		password = &Value{
			ValueFrom: &ValueFrom{
				SecretKeyRef: &SecretKeyRef{
					Key:  conf.Password.Key,
					Name: conf.Password.Name,
				},
			},
		}
	}

	return &Output{
		APIVersion: "logging.banzaicloud.io/v1beta1",
		Kind:       kind,
		Metadata:   meta,
		Spec: &OutputSpec{
			EnabledNamespaces: conf.EnabledNamespaces,
			Loki: &Loki{
				URL:                       conf.URL,
				ExtractKubernetesLabels:   conf.ExtractKubernetesLabels,
				ConfigureKubernetesLabels: conf.ConfigureKubernetesLabels,
				ExtraLabels:               conf.ExtraLabels,
				Labels:                    conf.Labels,
				RemoveKeys:                conf.RemoveKeys,
				Buffer: &Buffer{
					Timekey:       "1m",
					TimekeyWait:   "30s",
					TimekeyUseUtc: true,
				},
				Username: username,
				Password: password,
			},
		},
	}
}
