package labels

import "errors"

var _ Labels = (*Name)(nil)

type Name struct {
	model InternalName
}

type InternalNameProp struct {
	Name string `yaml:"app.kubernetes.io/name,omitempty"`
}

type InternalName struct {
	InternalNameProp  `yaml:",inline"`
	InternalComponent `yaml:",inline"`
}

func ForName(l *Component, name string) (*Name, error) {
	if name == "" {
		return nil, errors.New("name must not be nil")
	}
	return &Name{model: InternalName{
		InternalNameProp:  InternalNameProp{Name: name},
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
