package helper

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
