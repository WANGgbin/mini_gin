package mini_gin

func MergeParam(origin *map[string]string, delta map[string]string) {
	if len(delta) == 0 {
		return
	}

	if *origin == nil {
		*origin = delta
		return
	}

	for key, value := range delta {
		(*origin)[key] = value
	}
}
