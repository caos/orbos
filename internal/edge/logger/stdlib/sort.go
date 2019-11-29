package stdlib

type field struct {
	pos   uint8
	key   string
	value interface{}
}

type ByPosition []field

func (a ByPosition) Len() int      { return len(a) }
func (a ByPosition) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPosition) Less(i, j int) bool {
	less := a[i].pos < a[j].pos || a[i].pos == a[j].pos && a[i].key < a[j].key
	return less
}

func toStructFields(f map[string]interface{}) []field {

	sortableFields := make([]field, 0)

	for k, v := range f {
		sortableField := field{
			key:   k,
			value: v,
			pos:   9,
		}

		switch k {
		case "ts":
			sortableField.pos = 0
		case "msg":
			sortableField.pos = 1
		case "file":
			sortableField.pos = 2
		case "line":
			sortableField.pos = 3
		}

		sortableFields = append(sortableFields, sortableField)
	}

	return sortableFields
}
