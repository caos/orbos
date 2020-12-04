package labels

import "errors"

var _ Labels = (*Component)(nil)

type Component struct {
	model InternalComponent
	*API
}

func ForComponent(l *API, component string) (*Component, error) {
	if component == "" {
		return nil, errors.New("component must not be nil")
	}
	return &Component{
		API: l,
		model: InternalComponent{
			InternalComponentProp: InternalComponentProp{Component: component},
			InternalAPI:           l.model,
		},
	}, nil
}

func MustForComponent(l *API, component string) *Component {
	c, err := ForComponent(l, component)
	if err != nil {
		panic(err)
	}
	return c
}

func (l *Component) Equal(r comparable) bool {
	if right, ok := r.(*Component); ok {
		return l.model == right.model
	}
	return false
}

func (l *Component) MarshalYAML() (interface{}, error) {
	return nil, errors.New("type *labels.Component is not serializable")
}

type InternalComponentProp struct {
	Component string `yaml:"app.kubernetes.io/component"`
}

type InternalComponent struct {
	InternalComponentProp `yaml:",inline"`
	InternalAPI           `yaml:",inline"`
}
