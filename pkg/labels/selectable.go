package labels

import "gopkg.in/yaml.v3"

var (
	_          IDLabels = (*Selectable)(nil)
	selectable          = InternalSelectableProp{Selectable: "yes"}
)

type Selectable struct {
	model InternalSelectable
	base  *Name
}

type InternalSelectableProp struct {
	Selectable string `yaml:"orbos.ch/selectable"`
}

type InternalSelectable struct {
	InternalSelectableProp `yaml:",inline"`
	InternalName           `yaml:",inline"`
}

func AsSelectable(l *Name) *Selectable {
	return &Selectable{
		base: l,
		model: InternalSelectable{
			InternalName:           l.model,
			InternalSelectableProp: selectable,
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

/*
func (l *Selectable) Major() int8 {
	return l.base.Major()
}
*/

func (l *Selectable) Name() string {
	return l.base.Name()
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

func MustForNameK8SMap(l *Component, name string) map[string]string {
	return MustK8sMap(MustForName(l, name))
}
