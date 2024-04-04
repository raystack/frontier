package utils

import "strings"

func ExtractDomainFromEmail(email string) string {
	parts := strings.Split(email, "@")
	if len(parts) == 2 {
		return parts[1]
	}
	return ""
}
