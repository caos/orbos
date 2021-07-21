package crds

type MappingConfig struct {
	Name      string
	Namespace string
	Prefix    string
	Service   string
	Host      string
}
type MappingSpec struct {
	Prefix  string `yaml:"prefix"`
	Service string `yaml:"service"`
	Host    string `yaml:"host"`
}
type Mapping struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   *Metadata    `yaml:"metadata"`
	Spec       *MappingSpec `yaml:"spec"`
}

func GetMappingFromConfig(conf *MappingConfig) *Mapping {

	var metadata *Metadata
	if conf.Namespace != "" {
		metadata = &Metadata{
			Name:      conf.Name,
			Namespace: conf.Namespace,
		}
	} else {
		metadata = &Metadata{
			Name: conf.Name,
		}
	}

	return &Mapping{
		APIVersion: "getambassador.io/v2",
		Kind:       "Mapping",
		Metadata:   metadata,
		Spec: &MappingSpec{
			Prefix:  conf.Prefix,
			Service: conf.Service,
			Host:    conf.Host,
		},
	}
}
