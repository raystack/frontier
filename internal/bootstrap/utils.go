package bootstrap

func AppendIfUnique[T comparable](slice1 []T, slice2 []T) []T {
	for _, i := range slice2 {
		if !Contains(slice1, i) {
			slice1 = append(slice1, i)
		}
	}

	return slice1
}

func Contains[T comparable](s []T, e T) bool {
	for _, v := range s {
		if v == e {
			return true
		}
	}
	return false
}
