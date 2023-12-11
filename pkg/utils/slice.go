package utils

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

func ContainsFunc[T any](s []T, f func(T) bool) bool {
	for _, v := range s {
		if f(v) {
			return true
		}
	}
	return false
}

func Map[T1, T2 any](s []T1, f func(T1) T2) []T2 {
	result := make([]T2, len(s))
	for i, v := range s {
		result[i] = f(v)
	}
	return result
}

func Filter[T any](s []T, f func(T) bool) []T {
	result := make([]T, 0)
	for _, v := range s {
		if f(v) {
			result = append(result, v)
		}
	}
	return result
}

func Deduplicate[T comparable](s []T) []T {
	// Create a map of all unique elements.
	seen := make(map[T]struct{})
	j := 0
	for _, v := range s {
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		s[j] = v
		j++
	}

	// Slice the map to create a new slice of unique elements.
	return s[:j]
}
