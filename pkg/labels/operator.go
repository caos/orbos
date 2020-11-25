package labels

import "errors"

var _ Comparable = (*Operator)(nil)

type Operator struct {
	model InternalOperator
}

type InternalOperator struct {
	Version   string `yaml:"app.kubernetes.io/version"`
	PartOf    string `yaml:"app.kubernetes.io/part-of"`
	ManagedBy string `yaml:"app.kubernetes.io/managed-by"`
}

func ForOperator(operator, version string) (*Operator, error) {
	return &Operator{model: InternalOperator{
		Version:   version,
		PartOf:    "ORBOS",
		ManagedBy: operator,
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
