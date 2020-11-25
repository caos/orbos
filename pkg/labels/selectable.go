package labels

var selectProperty = InternalSelectProp{Select: true}

var _ Labels = (*Selectable)(nil)

type Selectable struct {
	model InternalSelectable
}

type InternalSelectProp struct {
	Select bool `yaml:"orbos.ch/select"`
}

type InternalSelectable struct {
	InternalSelectProp `yaml:",inline"`
	InternalName       `yaml:",inline"`
}

func AsSelectable(l *Name) *Selectable {
	return &Selectable{model: InternalSelectable{
		InternalSelectProp: selectProperty,
		InternalName:       l.model,
	}}
}

func (l *Selectable) Equal(r Comparable) bool {
	if right, ok := r.(*Selectable); ok {
		return l.model == right.model
	}
	return false
}

func (l *Selectable) MarshalYAML() (interface{}, error) {
	return l.model, nil
}
