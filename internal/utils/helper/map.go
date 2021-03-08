package helper

func OverwriteExistingKey(values map[string]string, first *string, second string) {
	if *first != "" && second != "" {
		value, ok := values[*first]
		if ok && value != "" {
			delete(values, *first)
			values[second] = value
			*first = second
		}
	}
}

func OverwriteExistingValues(first map[string]string, second map[string]string) {
	for k, v := range second {
		if v != "" {
			_, ok := first[k]
			if ok {
				first[k] = ""
				first[k] = v
			}
		}
	}
}
