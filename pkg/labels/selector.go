package labels

import "gopkg.in/yaml.v3"

var _ Labels = (*Selector)(nil)

type Selector struct {
	model InternalSelector
	base  *Name
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
		base: l,
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

func (l *Selector) Major() int8 {
	return l.base.Major()
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

func (l *Selector) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&l.model); err != nil {
		return err
	}
	l.base = &Name{}
	return node.Decode(l.base)
}
