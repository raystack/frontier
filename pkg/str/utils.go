package str

import "strings"

func DefaultStringIfEmpty(str string, defaultString string) string {
	if str != "" {
		return str
	}
	return defaultString
}

func IsStringEmpty(str string) bool {
	return strings.TrimSpace(str) == ""
}
