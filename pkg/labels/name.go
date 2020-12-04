package labels

import "errors"

var _ Labels = (*Name)(nil)

type Name struct {
	model InternalName
	*Component
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
	return &Name{
		Component: l,
		model: InternalName{
			InternalNameProp:  InternalNameProp{Name: name},
			InternalComponent: l.model,
		},
	}, nil
}

func MustForName(l *Component, name string) *Name {
	n, err := ForName(l, name)
	if err != nil {
		panic(err)
	}
	return n
}

func (l *Name) Equal(r comparable) bool {
	if right, ok := r.(*Name); ok {
		return l.model == right.model
	}
	return false
}

func (l *Name) MarshalYAML() (interface{}, error) {
	return l.model, nil
}
