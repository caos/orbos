package format

import "sort"

type field struct {
	pos   uint8
	key   string
	value interface{}
}

type byPosition []field

func (a byPosition) Len() int      { return len(a) }
func (a byPosition) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a byPosition) Less(i, j int) bool {
	less := a[i].pos < a[j].pos || a[i].pos == a[j].pos && a[i].key < a[j].key
	return less
}

func toCommitFields(f map[string]string) []field {
	sortableFields := make([]field, 0)

	for k, v := range f {
		sortableField := field{
			key:   k,
			value: v,
			pos:   2,
		}

		switch k {
		case "msg":
			sortableField.pos = 0
		case "err":
			sortableField.pos = 1
		case "ts":
			continue
		case "file":
			continue
		case "line":
			continue
		}
		sortableFields = append(sortableFields, sortableField)
	}

	sort.Sort(byPosition(sortableFields))

	return sortableFields
}

func toLogFields(f map[string]string) []field {

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

	sort.Sort(byPosition(sortableFields))

	return sortableFields
}
