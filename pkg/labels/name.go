package labels

var _ Comparable = (*Name)(nil)

type Name struct {
	model InternalName
}

type InternalName struct {
	Name              string `yaml:"app.kubernetes.io/name"`
	InternalComponent `yaml:",inline"`
}

func ForName(l *Component, name string) (*Name, error) {
	return &Name{model: InternalName{
		Name:              name,
		InternalComponent: l.model,
	}}, nil
}

func (l *Name) Equal(r Comparable) bool {
	if right, ok := r.(*Name); ok {
		return l.model == right.model
	}
	return false
}

func (l *Name) MarshalYAML() (interface{}, error) {
	return l.model, nil
}
