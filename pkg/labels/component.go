package labels

import "errors"

var _ Labels = (*Component)(nil)

type Component struct {
	model InternalComponent
}

type InternalComponentProp struct {
	Component string `yaml:"app.kubernetes.io/component"`
}

type InternalComponent struct {
	InternalComponentProp `yaml:",inline"`
	InternalAPI           `yaml:",inline"`
}

func ForComponent(l *API, component string) (*Component, error) {
	if component == "" {
		return nil, errors.New("component must not be nil")
	}
	return &Component{model: InternalComponent{
		InternalComponentProp: InternalComponentProp{Component: component},
		InternalAPI:           l.model,
	}}, nil
}

func (l *Component) Equal(r Comparable) bool {
	if right, ok := r.(*Component); ok {
		return l.model == right.model
	}
	return false
}

func (l *Component) MarshalYAML() (interface{}, error) {
	return nil, errors.New("type *labels.Component is not serializable")
}
