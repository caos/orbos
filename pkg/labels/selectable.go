package labels

import "gopkg.in/yaml.v3"

var selectProperty = InternalSelectProp{Select: true}

var _ Labels = (*Selectable)(nil)

type Selectable struct {
	model InternalSelectable
	base  *Name
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
		base: l,
		model: InternalSelectable{
			InternalSelectProp: selectProperty,
			InternalName:       l.model,
		},
	}
}

func (l *Selectable) UnmarshalYAML(node *yaml.Node) error {
	if err := node.Decode(&l.model); err != nil {
		return err
	}
	l.base = &Name{}
	return node.Decode(l.base)
}

func (l *Selectable) Major() int8 {
	return l.base.Major()
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
