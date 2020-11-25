package labels

import "errors"

var _ Comparable = (*Component)(nil)

type Component struct {
	model InternalComponent
}

type InternalComponent struct {
	Component   string `yaml:"app.kubernetes.io/component"`
	InternalAPI `yaml:",inline"`
}

func ForComponent(l *API, component string) (*Component, error) {
	return &Component{model: InternalComponent{
		Component:   component,
		InternalAPI: l.model,
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
