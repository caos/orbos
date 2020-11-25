package labels

import "errors"

var _ Labels = (*Operator)(nil)

type Operator struct {
	model InternalOperator
}

type InternalPartofProp struct {
	PartOf string `yaml:"app.kubernetes.io/part-of,omitempty"`
}

type InternalManagedByProp struct {
	ManagedBy string `yaml:"app.kubernetes.io/managed-by,omitempty"`
}

type InternalOperator struct {
	InternalManagedByProp `yaml:",inline"`
	Version               string `yaml:"app.kubernetes.io/version"`
	InternalPartofProp    `yaml:",inline"`
}

func ForOperator(operator, version string) (*Operator, error) {

	if operator == "" || version == "" {
		return nil, errors.New("operator or version must not be nil")
	}

	return &Operator{model: InternalOperator{
		Version:               version,
		InternalPartofProp:    InternalPartofProp{PartOf: "ORBOS"},
		InternalManagedByProp: InternalManagedByProp{ManagedBy: operator},
	}}, nil
}

func (l *Operator) Equal(r Comparable) bool {
	if right, ok := r.(*Operator); ok {
		return l.model == right.model
	}
	return false
}

func (l *Operator) MarshalYAML() (interface{}, error) {
	return nil, errors.New("type *labels.Operator is not serializable")
}
