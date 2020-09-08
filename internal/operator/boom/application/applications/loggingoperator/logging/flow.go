package logging

type FlowConfig struct {
	Name           string
	Namespace      string
	SelectLabels   map[string]string
	Outputs        []string
	ClusterOutputs []string
	ParserType     string
}

type Parse struct {
	Type string `yaml:"type"`
}
type Parser struct {
	RemoveKeyNameField bool   `yaml:"remove_key_name_field"`
	ReserveData        bool   `yaml:"reserve_data"`
	Parse              *Parse `yaml:"parse"`
}
type TagNormaliser struct {
	Format string
}

type Filter struct {
	Parser        *Parser        `yaml:"parser,omitempty"`
	TagNormaliser *TagNormaliser `yaml:"tag_normaliser,omitempty"`
}

type FlowSpec struct {
	Filters           []*Filter `yaml:"filters,omitempty"`
	Match             []*Match  `yaml:"match,omitempty"`
	OutputRefs        []string  `yaml:"localOutputRefs"`
	ClusterOutputRefs []string  `yaml:"globalOutputRefs"`
}
type Match struct {
	Select  *Select `yaml:"select"`
	Exclude *Select `yaml:"exlcude"`
}
type Select struct {
	Labels     map[string]string `yaml:"labels,omitempty"`
	Hosts      []string          `yaml:"hosts,omitempty"`
	Namespaces []string          `yaml:"namespaces,omitempty"`
}

type Flow struct {
	APIVersion string    `yaml:"apiVersion"`
	Kind       string    `yaml:"kind"`
	Metadata   *Metadata `yaml:"metadata"`
	Spec       *FlowSpec `yaml:"spec"`
}

func NewFlow(conf *FlowConfig) *Flow {
	return &Flow{
		APIVersion: "logging.banzaicloud.io/v1beta1",
		Kind:       "Flow",
		Metadata: &Metadata{
			Name:      conf.Name,
			Namespace: conf.Namespace,
		},
		Spec: &FlowSpec{
			Filters: []*Filter{
				{
					Parser: &Parser{
						RemoveKeyNameField: true,
						ReserveData:        true,
						Parse: &Parse{
							Type: conf.ParserType,
						},
					},
				},
				{
					TagNormaliser: &TagNormaliser{Format: "${namespace}.${container}.${pod}"},
				},
				{
					TagNormaliser: &TagNormaliser{Format: "${namespace_name}.${container_name}.${pod_name}"},
				},
			},
			Match: []*Match{
				{
					Select: &Select{
						Labels: conf.SelectLabels,
					},
				},
			},
			OutputRefs:        conf.Outputs,
			ClusterOutputRefs: conf.ClusterOutputs,
		},
	}
}
