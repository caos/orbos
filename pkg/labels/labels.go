package labels

type Comparable interface {
	Equal(Comparable) bool
}
