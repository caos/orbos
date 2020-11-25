package labels

import "errors"

var _ Labels = (*API)(nil)

type API struct {
	model InternalAPI
}

type InternalAPI struct {
	Kind             string `yaml:"orbos.ch/kind"`
	ApiVersion       string `yaml:"orbos.ch/apiversion"`
	InternalOperator `yaml:",inline"`
}

func ForAPI(l *Operator, kind, version string) (*API, error) {
	if kind == "" || version == "" {
		return nil, errors.New("kind and version must not be nil")
	}

	return &API{model: InternalAPI{
		Kind:             kind,
		ApiVersion:       version,
		InternalOperator: l.model,
	}}, nil
}

func (l *API) Equal(r comparable) bool {
	if right, ok := r.(*API); ok {
		return l.model == right.model
	}
	return false
}

func (l *API) MarshalYAML() (interface{}, error) {
	return nil, errors.New("type *labels.API is not serializable")
}
