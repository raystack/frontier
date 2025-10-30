package utils

// Coalesce returns the first non-empty string from the provided values.
// If the first value is non-empty, it is returned; otherwise the second value is returned.
func Coalesce(preferred, fallback string) string {
	if preferred != "" {
		return preferred
	}
	return fallback
}
