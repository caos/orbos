package labels

var selectProperty = InternalSelectProp{Select: true}

var _ Labels = (*Selectable)(nil)

type Selectable struct {
	model InternalSelectable
	*Name
}

type InternalSelectProp struct {
	Select bool `yaml:"orbos.ch/select"`
}

type InternalSelectable struct {
	InternalSelectProp `yaml:",inline"`
	InternalName       `yaml:",inline"`
}

func AsSelectable(l *Name) *Selectable {
	return &Selectable{
		Name: l,
		model: InternalSelectable{
			InternalSelectProp: selectProperty,
			InternalName:       l.model,
		},
	}
}

func (l *Selectable) Equal(r comparable) bool {
	if right, ok := r.(*Selectable); ok {
		return l.model == right.model
	}
	return false
}

func (l *Selectable) MarshalYAML() (interface{}, error) {
	return l.model, nil
}

func MustForNameAsSelectableK8SMap(l *Component, name string) map[string]string {
	return MustK8sMap(AsSelectable(MustForName(l, name)))
}

func MustForNameK8SMap(l *Component, name string) map[string]string {
	return MustK8sMap(MustForName(l, name))
}
