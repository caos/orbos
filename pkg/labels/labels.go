package labels

import "gopkg.in/yaml.v3"

type Labels interface {
	Comparable
	yaml.Marshaler
}

type Comparable interface {
	Equal(Comparable) bool
}
