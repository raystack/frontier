package utils

func DefaultStringIfEmpty(str string, defaultString string) string {
	if str != "" {
		return str
	}
	return defaultString
}
