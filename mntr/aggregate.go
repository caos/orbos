package mntr

import "sort"

type Field struct {
	Pos   uint8
	Key   string
	Value interface{}
}

type ByPosition []*Field

func (a ByPosition) Len() int      { return len(a) }
func (a ByPosition) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByPosition) Less(i, j int) bool {
	less := a[i].Pos < a[j].Pos || a[i].Pos == a[j].Pos && a[i].Key < a[j].Key
	return less
}

func AggregateCommitFields(f map[string]string) []*Field {
	sortableFields := make([]*Field, 0)

	for k, v := range f {
		sortableField := &Field{
			Key:   k,
			Value: v,
			Pos:   2,
		}

		switch k {
		case "msg":
			sortableField.Pos = 0
		case "dbg":
			sortableField.Pos = 0
		case "evt":
			sortableField.Pos = 0
		case "err":
			sortableField.Pos = 1
		case "ts":
			continue
		case "src":
			continue
		}
		sortableFields = append(sortableFields, sortableField)
	}

	sort.Sort(ByPosition(sortableFields))

	return sortableFields
}

func AggregateLogFields(f map[string]string) []*Field {

	sortableFields := make([]*Field, 0)

	for k, v := range f {
		sortableField := &Field{
			Key:   k,
			Value: v,
			Pos:   9,
		}

		switch k {
		case "ts":
			sortableField.Pos = 0
		case "msg":
			sortableField.Pos = 1
		case "err":
			sortableField.Pos = 1
		case "dbg":
			sortableField.Pos = 1
		case "evt":
			sortableField.Pos = 1
		case "src":
			sortableField.Pos = 2
		}

		sortableFields = append(sortableFields, sortableField)
	}

	sort.Sort(ByPosition(sortableFields))

	return sortableFields
}
