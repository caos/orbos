package labels

var _ Labels = (*Selector)(nil)

type Selector struct {
	model InternalSelector
	*Name
}

type InternalSelector struct {
	InternalSelectProp    `yaml:",inline"`
	InternalNameProp      `yaml:",inline"`
	InternalComponentProp `yaml:",inline"`
	InternalManagedByProp `yaml:",inline"`
	InternalPartofProp    `yaml:",inline"`
}

func DeriveSelector(l *Name, open bool) *Selector {
	selector := &Selector{
		Name: l,
		model: InternalSelector{
			InternalSelectProp:    selectProperty,
			InternalComponentProp: l.model.InternalComponentProp,
		},
	}

	if !open {
		selector.model.InternalPartofProp = l.model.InternalPartofProp
		selector.model.InternalNameProp = l.model.InternalNameProp
		selector.model.InternalManagedByProp = l.model.InternalManagedByProp
	}

	return selector
}

func (l *Selector) Equal(r comparable) bool {
	if right, ok := r.(*Selector); ok {
		return l.model == right.model
	}
	return false
}

func (l *Selector) MarshalYAML() (interface{}, error) {
	return l.model, nil
}
